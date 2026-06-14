package knowledge

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/GusAguilra/LinuxHealthDoctor/internal/baseline"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/core"
)

type Fact struct {
	ID        string
	Type      string
	Component core.Component
	Name      string
	Value     float64
	Severity  core.Severity
	Timestamp time.Time
	Source    string
}

type Rule struct {
	ID           string
	Name         string
	Description  string
	Component    core.Component
	Severity     core.Severity
	Conditions   []Condition
	Conclusion   string
	Certainty    float64
	Remediations []core.Remediation
}

type Condition struct {
	FactName string
	Operator string
	Value    float64
}

type AnalysisResult struct {
	ID              string
	Timestamp       time.Time
	OverallSeverity core.Severity
	Conclusions     []Conclusion
	RootCause       *Conclusion
	Chain           []Conclusion
}

type Conclusion struct {
	ID           string
	Title        string
	Description  string
	Severity     core.Severity
	Component    core.Component
	Certainty    float64
	Evidence     []string
	IsRootCause  bool
	Remediations []core.Remediation
}

type Engine struct {
	mu       sync.RWMutex
	facts    []Fact
	factByID map[string]*Fact
	rules    []Rule
	ruleByID map[string]*Rule
}

func NewEngine() *Engine {
	return &Engine{
		facts:    make([]Fact, 0),
		factByID: make(map[string]*Fact),
		rules:    make([]Rule, 0),
		ruleByID: make(map[string]*Rule),
	}
}

func (e *Engine) AddRule(rule Rule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = append(e.rules, rule)
	e.ruleByID[rule.ID] = &e.rules[len(e.rules)-1]
}

func (e *Engine) AddRules(rules []Rule) {
	for i := range rules {
		e.AddRule(rules[i])
	}
}

func (e *Engine) Rules() []Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()
	r := make([]Rule, len(e.rules))
	copy(r, e.rules)
	return r
}

func (e *Engine) Facts() []Fact {
	e.mu.RLock()
	defer e.mu.RUnlock()
	f := make([]Fact, len(e.facts))
	copy(f, e.facts)
	return f
}

func (e *Engine) IngestFacts(ctx context.Context, results []*core.CheckResult) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, res := range results {
		if res == nil {
			continue
		}

		for name, val := range res.Metrics {
			fact := Fact{
				ID:        fmt.Sprintf("%s.%s", res.ID, name),
				Type:      "metric",
				Component: res.Category,
				Name:      fmt.Sprintf("%s.%s", res.Category, name),
				Value:     val,
				Severity:  res.Severity,
				Timestamp: res.Timestamp,
				Source:    res.ID,
			}
			e.facts = append(e.facts, fact)
			e.factByID[fact.ID] = &e.facts[len(e.facts)-1]
		}

		switch res.Status {
		case core.StatusFail:
			fact := Fact{
				ID:        fmt.Sprintf("%s.status", res.ID),
				Type:      "state",
				Component: res.Category,
				Name:      fmt.Sprintf("%s.status", res.Category),
				Value:     0,
				Severity:  res.Severity,
				Timestamp: res.Timestamp,
				Source:    res.ID,
			}
			e.facts = append(e.facts, fact)
			e.factByID[fact.ID] = &e.facts[len(e.facts)-1]
		case core.StatusPass:
			fact := Fact{
				ID:        fmt.Sprintf("%s.status", res.ID),
				Type:      "state",
				Component: res.Category,
				Name:      fmt.Sprintf("%s.status", res.Category),
				Value:     1,
				Severity:  core.SeverityNone,
				Timestamp: res.Timestamp,
				Source:    res.ID,
			}
			e.facts = append(e.facts, fact)
			e.factByID[fact.ID] = &e.facts[len(e.facts)-1]
		}

		if res.Message != "" {
			detailVal := 0.0
			if res.Status == core.StatusPass {
				detailVal = 1.0
			}
			fact := Fact{
				ID:        fmt.Sprintf("%s.detail", res.ID),
				Type:      "event",
				Component: res.Category,
				Name:      fmt.Sprintf("%s.detail.%s", res.Category, res.ID),
				Value:     detailVal,
				Severity:  res.Severity,
				Timestamp: res.Timestamp,
				Source:    res.ID,
			}
			e.facts = append(e.facts, fact)
			e.factByID[fact.ID] = &e.facts[len(e.facts)-1]
		}
	}
	return nil
}

func (e *Engine) Analyze(ctx context.Context) (*AnalysisResult, error) {
	e.mu.RLock()
	facts := make([]Fact, len(e.facts))
	copy(facts, e.facts)
	rules := make([]Rule, len(e.rules))
	copy(rules, e.rules)
	e.mu.RUnlock()

	activeFacts := e.prepareFactMap(facts)

	conclusions := e.forwardChain(rules, activeFacts)

	grouped := e.groupByComponent(conclusions)
	var overall core.Severity
	for _, cc := range grouped {
		for _, c := range cc {
			if c.Severity > overall {
				overall = c.Severity
			}
		}
	}

	rootCause := e.findRootCause(conclusions, grouped)

	chain := e.buildCausalChain(conclusions, rootCause)

	return &AnalysisResult{
		ID:              fmt.Sprintf("rca-%d", time.Now().UnixNano()),
		Timestamp:       time.Now(),
		OverallSeverity: overall,
		Conclusions:     conclusions,
		RootCause:       rootCause,
		Chain:           chain,
	}, nil
}

func (e *Engine) AnalyzeWithBaseline(ctx context.Context, base *baseline.Baseline) (*AnalysisResult, error) {
	e.mu.RLock()
	facts := make([]Fact, len(e.facts))
	copy(facts, e.facts)
	rules := make([]Rule, len(e.rules))
	copy(rules, e.rules)
	e.mu.RUnlock()

	if base != nil && base.Metrics != nil {
		for name, mb := range base.Metrics {
			facts = append(facts, Fact{
				ID:        fmt.Sprintf("baseline.%s", name),
				Type:      "metric",
				Component: e.factComponent(name),
				Name:      name,
				Value:     mb.Mean,
				Severity:  core.SeverityInfo,
				Timestamp: base.Timestamp,
				Source:    "baseline",
			})
		}
	}

	activeFacts := e.prepareFactMap(facts)

	baselineRules := e.generateBaselineRules(base, facts)
	allRules := append(rules, baselineRules...)

	conclusions := e.forwardChain(allRules, activeFacts)

	grouped := e.groupByComponent(conclusions)
	var overall core.Severity
	for _, cc := range grouped {
		for _, c := range cc {
			if c.Severity > overall {
				overall = c.Severity
			}
		}
	}

	rootCause := e.findRootCause(conclusions, grouped)
	chain := e.buildCausalChain(conclusions, rootCause)

	return &AnalysisResult{
		ID:              fmt.Sprintf("rca-%d", time.Now().UnixNano()),
		Timestamp:       time.Now(),
		OverallSeverity: overall,
		Conclusions:     conclusions,
		RootCause:       rootCause,
		Chain:           chain,
	}, nil
}

func (e *Engine) prepareFactMap(facts []Fact) map[string]float64 {
	m := make(map[string]float64, len(facts))
	for _, f := range facts {
		m[f.Name] = f.Value
	}
	return m
}

func (e *Engine) forwardChain(rules []Rule, facts map[string]float64) []Conclusion {
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Severity > rules[j].Severity
	})

	var conclusions []Conclusion
	seen := make(map[string]bool)

	for _, rule := range rules {
		if matched, evidence := e.evaluateConditions(rule.Conditions, facts); matched {
			if seen[rule.ID] {
				continue
			}
			seen[rule.ID] = true

			certainty := rule.Certainty
			if certainty <= 0 {
				certainty = 0.5
			}

			conclusions = append(conclusions, Conclusion{
				ID:           rule.ID,
				Title:        rule.Name,
				Description:  rule.Conclusion,
				Severity:     rule.Severity,
				Component:    rule.Component,
				Certainty:    certainty,
				Evidence:     evidence,
				Remediations: rule.Remediations,
			})
		}
	}

	return conclusions
}

func (e *Engine) evaluateConditions(conditions []Condition, facts map[string]float64) (bool, []string) {
	if len(conditions) == 0 {
		return false, nil
	}

	var evidence []string
	for _, cond := range conditions {
		val, ok := facts[cond.FactName]
		if !ok {
			return false, nil
		}

		matched := false
		switch cond.Operator {
		case "gt":
			matched = val > cond.Value
		case "gte":
			matched = val >= cond.Value
		case "lt":
			matched = val < cond.Value
		case "lte":
			matched = val <= cond.Value
		case "eq":
			matched = val == cond.Value
		case "neq":
			matched = val != cond.Value
		default:
			return false, nil
		}

		if !matched {
			return false, nil
		}

		evidence = append(evidence, fmt.Sprintf("%s %s %v (actual: %v)", cond.FactName, cond.Operator, cond.Value, val))
	}

	return true, evidence
}

func (e *Engine) groupByComponent(conclusions []Conclusion) map[core.Component][]Conclusion {
	grouped := make(map[core.Component][]Conclusion)
	for _, c := range conclusions {
		grouped[c.Component] = append(grouped[c.Component], c)
	}
	return grouped
}

func (e *Engine) findRootCause(conclusions []Conclusion, grouped map[core.Component][]Conclusion) *Conclusion {
	if len(conclusions) == 0 {
		return nil
	}

	var best *Conclusion
	for i := range conclusions {
		conclusions[i].IsRootCause = false
	}

	for i := range conclusions {
		c := &conclusions[i]
		if len(c.Evidence) == 0 {
			continue
		}

		if best == nil {
			best = c
			continue
		}

		if c.Severity > best.Severity {
			best = c
		} else if c.Severity == best.Severity && c.Certainty > best.Certainty {
			best = c
		}
	}

	if best != nil {
		best.IsRootCause = true
	}
	return best
}

func (e *Engine) buildCausalChain(conclusions []Conclusion, rootCause *Conclusion) []Conclusion {
	if rootCause == nil || len(conclusions) == 0 {
		return conclusions
	}

	byComponent := e.groupByComponent(conclusions)

	var chain []Conclusion
	chain = append(chain, *rootCause)

	for comp, cc := range byComponent {
		if comp == rootCause.Component {
			continue
		}
		for _, c := range cc {
			if c.Severity >= rootCause.Severity || math.Abs(c.Certainty-rootCause.Certainty) < 0.15 {
				chain = append(chain, c)
			}
		}
	}

	sort.Slice(chain, func(i, j int) bool {
		if chain[i].Severity != chain[j].Severity {
			return chain[i].Severity > chain[j].Severity
		}
		return chain[i].Certainty > chain[j].Certainty
	})

	return chain
}

func (e *Engine) generateBaselineRules(base *baseline.Baseline, facts []Fact) []Rule {
	if base == nil {
		return nil
	}

	var rules []Rule
	for _, f := range facts {
		mb, ok := base.Metrics[f.Name]
		if !ok || mb.Mean == 0 {
			continue
		}

		deviation := math.Abs(f.Value-mb.Mean) / mb.Mean
		if deviation > 0.25 {
			var sev core.Severity
			if deviation > 0.5 {
				sev = core.SeverityCritical
			} else {
				sev = core.SeverityWarning
			}

			factName := f.Name
			thresholdVal := mb.Mean * 1.25
			rules = append(rules, Rule{
				ID:          fmt.Sprintf("baseline.deviation.%s", f.Name),
				Name:        fmt.Sprintf("Baseline Deviation: %s", f.Name),
				Description: fmt.Sprintf("Value deviates >%.0f%% from baseline", deviation*100),
				Component:   f.Component,
				Severity:    sev,
				Conditions: []Condition{
					{
						FactName: factName,
						Operator: "gt",
						Value:    thresholdVal,
					},
				},
				Conclusion: fmt.Sprintf("%s has deviated %.1f%% from baseline mean %.2f (current: %.2f)", f.Name, deviation*100, mb.Mean, f.Value),
				Certainty:  0.7,
			})
		}
	}
	return rules
}

func (e *Engine) factComponent(name string) core.Component {
	for _, comp := range core.AllComponents() {
		prefix := string(comp) + "."
		if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
			return comp
		}
	}
	return core.Component("unknown")
}
