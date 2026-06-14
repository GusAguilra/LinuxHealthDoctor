package baseline

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/GusAguilra/LinuxHealthDoctor/internal/core"
)

type Engine struct {
	store Store
}

type Store interface {
	SaveBaseline(ctx context.Context, b *Baseline) error
	LoadBaseline(ctx context.Context, id string) (*Baseline, error)
	ListBaselines(ctx context.Context) ([]*Baseline, error)
	DeleteBaseline(ctx context.Context, id string) error
}

func NewEngine(store Store) *Engine {
	return &Engine{store: store}
}

type Baseline struct {
	ID          string
	Name        string
	Description string
	Timestamp   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Distro      string
	Kernel      string
	Hostname    string
	Metrics     map[string]MetricBaseline
	Configs     map[string]ConfigBaseline
	Services    map[string]ServiceBaseline
	Version     int
	Tags        []string
}

type MetricBaseline struct {
	Name    string
	Unit    string
	Mean    float64
	Median  float64
	P95     float64
	P99     float64
	StdDev  float64
	Min     float64
	Max     float64
	Samples int
}

type ConfigBaseline struct {
	Path        string
	Key         string
	Expected    string
	Permissions os.FileMode
}

type ServiceBaseline struct {
	Name            string
	ExpectedStatus  string
	ExpectedEnabled bool
}

type Deviation struct {
	Metric   string
	Expected float64
	Actual   float64
	Delta    float64
	Severity core.Severity
}

type ComparisonResult struct {
	BaselineID   string
	Timestamp    time.Time
	Deviations   []Deviation
	TotalMetrics int
	PassedChecks int
	HealthScore  float64
	Summary      string
}

func (e *Engine) Capture(ctx context.Context) (*Baseline, error) {
	baseline := &Baseline{
		ID:        fmt.Sprintf("bl-%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Version:   1,
		Metrics:   make(map[string]MetricBaseline),
		Configs:   make(map[string]ConfigBaseline),
		Services:  make(map[string]ServiceBaseline),
	}

	return baseline, nil
}

func (e *Engine) Compare(ctx context.Context, base *Baseline, current *core.AggregatedResult) (*ComparisonResult, error) {
	result := &ComparisonResult{
		BaselineID:   base.ID,
		Timestamp:    time.Now(),
		Deviations:   make([]Deviation, 0),
		TotalMetrics: len(base.Metrics),
	}

	for name, expected := range base.Metrics {
		for _, results := range current.Results {
			for _, r := range results {
				if val, ok := r.Metrics[name]; ok {
					delta := val - expected.Mean
					deviation := Deviation{
						Metric:   name,
						Expected: expected.Mean,
						Actual:   val,
						Delta:    delta,
						Severity: core.SeverityInfo,
					}
					if deviation.Delta > expected.StdDev*2 {
						deviation.Severity = core.SeverityWarning
					}
					if deviation.Delta > expected.StdDev*3 {
						deviation.Severity = core.SeverityCritical
					}
					result.Deviations = append(result.Deviations, deviation)
					break
				}
			}
		}
	}

	result.PassedChecks = result.TotalMetrics - len(result.Deviations)
	if result.TotalMetrics > 0 {
		result.HealthScore = float64(result.PassedChecks) / float64(result.TotalMetrics) * 100
	}
	result.Summary = fmt.Sprintf("Compared %d metrics: %d passed, %d deviations", result.TotalMetrics, result.PassedChecks, len(result.Deviations))

	return result, nil
}
