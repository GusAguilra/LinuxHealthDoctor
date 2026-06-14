package baseline

import (
	"fmt"
	"math"
	"strings"

	"github.com/GusAguilra/LinuxHealthDoctor/internal/core"
)

type Comparator struct {
	MetricThreshold float64
	ConfigStrict    bool
}

type CompareOption func(*Comparator)

func WithMetricThreshold(threshold float64) CompareOption {
	return func(c *Comparator) {
		c.MetricThreshold = threshold
	}
}

func WithConfigStrict(strict bool) CompareOption {
	return func(c *Comparator) {
		c.ConfigStrict = strict
	}
}

func NewComparator(opts ...CompareOption) *Comparator {
	c := &Comparator{
		MetricThreshold: 0.1,
		ConfigStrict:    true,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Comparator) CompareMetrics(expected, actual MetricBaseline) (*Deviation, error) {
	if expected.Name != actual.Name {
		return nil, fmt.Errorf("metric name mismatch: %s vs %s", expected.Name, actual.Name)
	}

	delta := actual.Mean - expected.Mean
	if expected.Mean != 0 {
		delta = (actual.Mean - expected.Mean) / math.Abs(expected.Mean)
	}

	deviation := &Deviation{
		Metric:   expected.Name,
		Expected: expected.Mean,
		Actual:   actual.Mean,
		Delta:    delta,
		Severity: c.computeSeverity(delta, expected.StdDev),
	}

	return deviation, nil
}

func (c *Comparator) CompareConfigs(expected, actual ConfigBaseline) ([]string, error) {
	var diffs []string

	if expected.Path != actual.Path {
		diffs = append(diffs, fmt.Sprintf("path mismatch: %s vs %s", expected.Path, actual.Path))
	}

	if expected.Expected != actual.Expected {
		diffs = append(diffs, fmt.Sprintf("value mismatch for key %s: expected %s, got %s", expected.Key, expected.Expected, actual.Expected))
	}

	if expected.Permissions != actual.Permissions && c.ConfigStrict {
		diffs = append(diffs, fmt.Sprintf("permissions mismatch for %s: expected %o, got %o", expected.Path, expected.Permissions, actual.Permissions))
	}

	return diffs, nil
}

func (c *Comparator) CompareServices(expected, actual ServiceBaseline) ([]string, error) {
	var diffs []string

	if expected.ExpectedStatus != actual.ExpectedStatus {
		diffs = append(diffs, fmt.Sprintf("service %s status mismatch: expected %s, got %s", expected.Name, expected.ExpectedStatus, actual.ExpectedStatus))
	}

	if expected.ExpectedEnabled != actual.ExpectedEnabled {
		state := "disabled"
		if actual.ExpectedEnabled {
			state = "enabled"
		}
		diffs = append(diffs, fmt.Sprintf("service %s enablement mismatch: expected enabled=%v, got %s", expected.Name, expected.ExpectedEnabled, state))
	}

	return diffs, nil
}

func (c *Comparator) computeSeverity(delta, stdDev float64) core.Severity {
	absDelta := math.Abs(delta)

	if stdDev == 0 {
		if absDelta > c.MetricThreshold {
			return core.SeverityWarning
		}
		return core.SeverityNone
	}

	ratio := absDelta / stdDev
	switch {
	case ratio >= 3:
		return core.SeverityCritical
	case ratio >= 2:
		return core.SeverityWarning
	case ratio >= 1:
		return core.SeverityInfo
	default:
		return core.SeverityNone
	}
}

func (c *Comparator) SummarizeDeviations(deviations []Deviation) string {
	if len(deviations) == 0 {
		return "No deviations detected"
	}

	var critical, warning, info int
	for _, d := range deviations {
		switch d.Severity {
		case core.SeverityCritical:
			critical++
		case core.SeverityWarning:
			warning++
		case core.SeverityInfo:
			info++
		}
	}

	parts := make([]string, 0, 3)
	if critical > 0 {
		parts = append(parts, fmt.Sprintf("%d critical", critical))
	}
	if warning > 0 {
		parts = append(parts, fmt.Sprintf("%d warning", warning))
	}
	if info > 0 {
		parts = append(parts, fmt.Sprintf("%d info", info))
	}

	return fmt.Sprintf("%d deviations: %s", len(deviations), strings.Join(parts, ", "))
}
