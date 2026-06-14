package baseline

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/linuxhealthdoctor/lhd/internal/core"
)

type Profile struct {
	Name        string
	Description string
	Distro      string
	Kernel      string
	Tags        []string
}

func NewProfile(name, description string) *Profile {
	return &Profile{
		Name:        name,
		Description: description,
	}
}

func (p *Profile) Compute(ctx *core.AggregatedResult) (*Baseline, error) {
	if ctx == nil {
		return nil, fmt.Errorf("aggregated result is nil")
	}

	baseline := &Baseline{
		ID:          fmt.Sprintf("bl-%d", time.Now().UnixNano()),
		Name:        p.Name,
		Description: p.Description,
		Timestamp:   time.Now(),
		Distro:      p.Distro,
		Kernel:      p.Kernel,
		Tags:        append([]string{}, p.Tags...),
		Version:     1,
		Metrics:     make(map[string]MetricBaseline),
		Configs:     make(map[string]ConfigBaseline),
		Services:    make(map[string]ServiceBaseline),
	}

	for component, results := range ctx.Results {
		for _, r := range results {
			for metricName, values := range p.groupMetrics(r.Metrics) {
				baseline.Metrics[metricName] = p.computeMetricBaseline(metricName, "", values)
			}

			if config, ok := r.Details["config_path"]; ok {
				baseline.Configs[metricKey(component, r.ID)] = ConfigBaseline{
					Path:        fmt.Sprintf("%v", config),
					Key:         metricNameFromDetails(r.Details),
					Expected:    fmt.Sprintf("%v", r.Details["expected"]),
					Permissions: 0644,
				}
			}

			if component == core.ComponentServices {
				baseline.Services[r.ID] = ServiceBaseline{
					Name:            r.ID,
					ExpectedStatus:  "running",
					ExpectedEnabled: true,
				}
			}
		}
	}

	return baseline, nil
}

func (p *Profile) groupMetrics(metrics map[string]float64) map[string][]float64 {
	grouped := make(map[string][]float64)
	for name, val := range metrics {
		grouped[name] = append(grouped[name], val)
	}
	return grouped
}

func (p *Profile) computeMetricBaseline(name, unit string, values []float64) MetricBaseline {
	if len(values) == 0 {
		return MetricBaseline{Name: name, Unit: unit}
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	mb := MetricBaseline{
		Name:    name,
		Unit:    unit,
		Min:     sorted[0],
		Max:     sorted[len(sorted)-1],
		Samples: len(sorted),
	}

	sum := 0.0
	for _, v := range sorted {
		sum += v
	}
	mb.Mean = sum / float64(len(sorted))

	if len(sorted)%2 == 0 {
		mb.Median = (sorted[len(sorted)/2-1] + sorted[len(sorted)/2]) / 2
	} else {
		mb.Median = sorted[len(sorted)/2]
	}

	p95Idx := int(math.Ceil(0.95*float64(len(sorted))) - 1)
	if p95Idx < 0 {
		p95Idx = 0
	}
	mb.P95 = sorted[p95Idx]

	p99Idx := int(math.Ceil(0.99*float64(len(sorted))) - 1)
	if p99Idx < 0 {
		p99Idx = 0
	}
	mb.P99 = sorted[p99Idx]

	variance := 0.0
	for _, v := range sorted {
		diff := v - mb.Mean
		variance += diff * diff
	}
	variance /= float64(len(sorted))
	mb.StdDev = math.Sqrt(variance)

	return mb
}

func metricKey(component core.Component, id string) string {
	return fmt.Sprintf("%s:%s", component, id)
}

func metricNameFromDetails(details map[string]interface{}) string {
	if name, ok := details["key"]; ok {
		return fmt.Sprintf("%v", name)
	}
	return "unknown"
}
