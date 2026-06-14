package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/linuxhealthdoctor/lhd/internal/core"
)

type ExecutionContext struct {
	ctx         context.Context
	cancel      context.CancelFunc
	parentCtx   context.Context
	eventBus    core.EventBus
	checkers    []Checker
	layers      [][]Checker
	results     []*core.CheckResult
	mu          sync.Mutex
	maxParallel int
	timeout     time.Duration
	progress    chan ProgressEvent
	progDone    chan struct{}
}

type ProgressEvent struct {
	Total     int
	Completed int
	Current   string
	Status    core.CheckStatus
}

func NewExecutionContext(ctx context.Context, categories []core.Component, opts ...ContextOption) (*ExecutionContext, error) {
	layers, err := global.ExecutionPlan(categories)
	if err != nil {
		return nil, err
	}

	timeout := 30 * time.Second
	execCtx, cancel := context.WithTimeout(ctx, timeout)

	ec := &ExecutionContext{
		ctx:         execCtx,
		cancel:      cancel,
		parentCtx:   ctx,
		eventBus:    core.NewInMemoryEventBus(),
		maxParallel: core.MaxParallelDefault,
		timeout:     timeout,
		progress:    make(chan ProgressEvent, 100),
		progDone:    make(chan struct{}),
	}

	for _, opt := range opts {
		opt(ec)
	}

	for _, layer := range layers {
		ec.checkers = append(ec.checkers, layer...)
	}

	ec.layers = layers

	return ec, nil
}

type ContextOption func(*ExecutionContext)

func WithMaxParallel(n int) ContextOption {
	return func(ec *ExecutionContext) {
		ec.maxParallel = n
	}
}

func WithTimeout(d time.Duration) ContextOption {
	return func(ec *ExecutionContext) {
		ec.timeout = d
		ec.cancel()
		ec.ctx, ec.cancel = context.WithTimeout(ec.parentCtx, d)
	}
}

func WithEventBus(bus core.EventBus) ContextOption {
	return func(ec *ExecutionContext) {
		ec.eventBus = bus
	}
}

func (ec *ExecutionContext) Run() *core.AggregatedResult {
	result := core.NewAggregatedResult()
	total := len(ec.checkers)
	completed := 0

	for _, layer := range ec.layers {
		if ec.ctx.Err() != nil {
			break
		}

		var wg sync.WaitGroup
		semaphore := make(chan struct{}, ec.maxParallel)
		var layerMu sync.Mutex
		var layerResults []*core.CheckResult

		for _, checker := range layer {
			wg.Add(1)
			semaphore <- struct{}{}
			go func(c Checker) {
				defer wg.Done()
				defer func() { <-semaphore }()

				ec.eventBus.Publish(string(core.EventCheckStart), &core.Event{
					Type:    core.EventCheckStart,
					Source:  c.ID(),
					Message: fmt.Sprintf("Starting check: %s", c.Name()),
				})

				start := time.Now()
				checkResult, err := c.Check(ec.ctx, &CheckRequest{})
				duration := time.Since(start)

				if err != nil {
					checkResult = &core.CheckResult{
						ID:      c.ID(),
						Status:  core.StatusError,
						Error:   err,
						Message: fmt.Sprintf("%s: %v", c.Name(), err),
					}
				} else if checkResult == nil {
					checkResult = &core.CheckResult{
						ID:      c.ID(),
						Status:  core.StatusError,
						Message: fmt.Sprintf("%s: returned nil result", c.Name()),
					}
				} else if checkResult.Status == core.StatusError && checkResult.Message == "" && checkResult.Error != nil {
					checkResult.Message = checkResult.Error.Error()
				}
				checkResult.Duration = duration

				ec.eventBus.Publish(string(core.EventCheckComplete), &core.Event{
					Type:     core.EventCheckComplete,
					Source:   c.ID(),
					Severity: checkResult.Severity,
					Message:  fmt.Sprintf("Check %s: %s", c.Name(), checkResult.Status),
				})

				layerMu.Lock()
				layerResults = append(layerResults, checkResult)
				layerMu.Unlock()

				ec.mu.Lock()
				completed++
				select {
				case ec.progress <- ProgressEvent{
					Total:     total,
					Completed: completed,
					Current:   c.Name(),
					Status:    checkResult.Status,
				}:
				default:
				}
				ec.mu.Unlock()
			}(checker)
		}
		wg.Wait()

		for _, r := range layerResults {
			result.AddResult(r)
		}
	}

	result.TotalChecks = len(result.AllResults)
	result.HealthScore = result.CalculateHealthScore()
	result.Duration = time.Since(result.Timestamp)

	ec.cancel()
	close(ec.progress)
	close(ec.progDone)

	return result
}

func NewFilteredExecutionContext(ec *ExecutionContext, ids []string) *ExecutionContext {
	idSet := make(map[string]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}
	filtered := make([]Checker, 0, len(ids))
	for _, c := range ec.checkers {
		if idSet[c.ID()] {
			filtered = append(filtered, c)
		}
	}
	ec.checkers = filtered
	// Rebuild layers from filtered checkers
	ec.layers = [][]Checker{}
	remaining := make([]Checker, len(filtered))
	copy(remaining, filtered)
	for len(remaining) > 0 {
		layer := make([]Checker, 0)
		for _, c := range remaining {
			deps := c.Dependencies()
			ready := true
			for _, dep := range deps {
				found := false
				for _, lc := range filtered {
					if lc.ID() == dep {
						found = true
						break
					}
				}
				if found && !idSet[deps[0]] {
					// dependency is outside our filter set
					ready = true
					continue
				}
				stillPending := false
				for _, r := range remaining {
					if r.ID() == dep {
						stillPending = true
						break
					}
				}
				if stillPending {
					ready = false
					break
				}
			}
			if ready {
				layer = append(layer, c)
			}
		}
		if len(layer) == 0 {
			layer = remaining[:1]
		}
		ec.layers = append(ec.layers, layer)
		remaining = remaining[len(layer):]
	}
	return ec
}

func (ec *ExecutionContext) Progress() <-chan ProgressEvent {
	return ec.progress
}

func (ec *ExecutionContext) EventBus() core.EventBus {
	return ec.eventBus
}
