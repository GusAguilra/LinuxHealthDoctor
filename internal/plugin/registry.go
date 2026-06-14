package plugin

import (
	"fmt"
	"sort"
	"sync"

	"github.com/linuxhealthdoctor/lhd/internal/core"
)

type ListOption func(*listOptions)

type listOptions struct {
	categories []core.Component
	tags       []string
	ids        []string
}

func WithCategories(categories ...core.Component) ListOption {
	return func(o *listOptions) {
		o.categories = categories
	}
}

func WithTags(tags ...string) ListOption {
	return func(o *listOptions) {
		o.tags = tags
	}
}

func WithIDs(ids ...string) ListOption {
	return func(o *listOptions) {
		o.ids = ids
	}
}

type Registry struct {
	mu         sync.RWMutex
	checkers   map[string]Checker
	categories map[core.Component][]Checker
}

var global = &Registry{
	checkers:   make(map[string]Checker),
	categories: make(map[core.Component][]Checker),
}

func Register(c Checker) {
	global.Register(c)
}

func Get(id string) Checker {
	return global.Get(id)
}

func List(opts ...ListOption) []Checker {
	return global.List(opts...)
}

func ExecutionPlan(categories []core.Component) ([][]Checker, error) {
	return global.ExecutionPlan(categories)
}

func (r *Registry) Register(c Checker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkers[c.ID()] = c
	r.categories[c.Category()] = append(r.categories[c.Category()], c)
}

func (r *Registry) Get(id string) Checker {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.checkers[id]
}

func (r *Registry) List(opts ...ListOption) []Checker {
	r.mu.RLock()
	defer r.mu.RUnlock()

	opt := &listOptions{}
	for _, o := range opts {
		o(opt)
	}

	var result []Checker
	for _, c := range r.checkers {
		if len(opt.categories) > 0 {
			found := false
			for _, cat := range opt.categories {
				if c.Category() == cat {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if len(opt.tags) > 0 {
			if !hasIntersection(c.Tags(), opt.tags) {
				continue
			}
		}
		if len(opt.ids) > 0 {
			found := false
			for _, id := range opt.ids {
				if c.ID() == id {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		result = append(result, c)
	}
	return result
}

func (r *Registry) ExecutionPlan(categories []core.Component) ([][]Checker, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	checkers := r.List(WithCategories(categories...))
	if len(checkers) == 0 {
		return nil, fmt.Errorf("no checks found for categories: %v", categories)
	}

	depGraph := make(map[string]map[string]bool)
	for _, c := range checkers {
		if _, ok := depGraph[c.ID()]; !ok {
			depGraph[c.ID()] = make(map[string]bool)
		}
		for _, dep := range c.Dependencies() {
			if r.checkers[dep] != nil {
				depGraph[c.ID()][dep] = true
			}
		}
	}

	var layers [][]Checker
	remaining := make(map[string]Checker)
	for _, c := range checkers {
		remaining[c.ID()] = c
	}

	for len(remaining) > 0 {
		var layer []Checker
		for id, c := range remaining {
			if len(depGraph[id]) == 0 {
				layer = append(layer, c)
			}
		}
		if len(layer) == 0 {
			return nil, fmt.Errorf("circular dependency detected in plugin execution plan")
		}
		sort.Slice(layer, func(i, j int) bool {
			return layer[i].ID() < layer[j].ID()
		})
		layers = append(layers, layer)
		for _, c := range layer {
			delete(remaining, c.ID())
			for _, deps := range depGraph {
				delete(deps, c.ID())
			}
		}
	}

	return layers, nil
}

func hasIntersection(a, b []string) bool {
	set := make(map[string]bool)
	for _, s := range a {
		set[s] = true
	}
	for _, s := range b {
		if set[s] {
			return true
		}
	}
	return false
}
