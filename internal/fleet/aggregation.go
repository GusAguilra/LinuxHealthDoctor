package fleet

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type FleetSummary struct {
	TotalHosts      int
	SuccessfulHosts int
	FailedHosts     int
	SuccessRate     float64
	TotalDuration   time.Duration
	AverageDuration time.Duration
	SlowestHost     string
	SlowestDuration time.Duration
	FastestHost     string
	FastestDuration time.Duration
	Errors          []HostError
}

type HostError struct {
	Hostname string
	Error    string
}

func AggregateResults(result *FleetResult) *FleetSummary {
	if result == nil || len(result.Results) == 0 {
		return &FleetSummary{}
	}

	summary := &FleetSummary{
		TotalHosts:      result.Hosts,
		SuccessfulHosts: result.Successful,
		FailedHosts:     result.Failed,
	}

	if result.Hosts > 0 {
		summary.SuccessRate = float64(result.Successful) / float64(result.Hosts) * 100.0
	}

	slowest := time.Duration(0)
	fastest := time.Duration(1<<63 - 1)
	var totalDuration time.Duration

	for _, r := range result.Results {
		if r.Duration > slowest {
			slowest = r.Duration
			summary.SlowestHost = r.Hostname
			summary.SlowestDuration = r.Duration
		}
		if r.Duration < fastest && r.Duration > 0 {
			fastest = r.Duration
			summary.FastestHost = r.Hostname
			summary.FastestDuration = r.Duration
		}
		totalDuration += r.Duration

		if !r.Success && r.Error != nil {
			summary.Errors = append(summary.Errors, HostError{
				Hostname: r.Hostname,
				Error:    r.Error.Error(),
			})
		}
	}

	if len(result.Results) > 0 {
		summary.AverageDuration = totalDuration / time.Duration(len(result.Results))
	}
	summary.TotalDuration = totalDuration

	if summary.FastestDuration == time.Duration(1<<63-1) {
		summary.FastestDuration = 0
	}

	return summary
}

func (s *FleetSummary) String() string {
	var b strings.Builder
	b.WriteString("Fleet Summary\n")
	b.WriteString(strings.Repeat("=", 40))
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("  Total Hosts:     %d\n", s.TotalHosts))
	b.WriteString(fmt.Sprintf("  Successful:      %d\n", s.SuccessfulHosts))
	b.WriteString(fmt.Sprintf("  Failed:          %d\n", s.FailedHosts))
	b.WriteString(fmt.Sprintf("  Success Rate:    %.1f%%\n", s.SuccessRate))
	b.WriteString(fmt.Sprintf("  Total Duration:  %v\n", s.TotalDuration))
	b.WriteString(fmt.Sprintf("  Average:         %v\n", s.AverageDuration))

	if s.SlowestHost != "" {
		b.WriteString(fmt.Sprintf("  Slowest:         %s (%v)\n", s.SlowestHost, s.SlowestDuration))
	}
	if s.FastestHost != "" {
		b.WriteString(fmt.Sprintf("  Fastest:         %s (%v)\n", s.FastestHost, s.FastestDuration))
	}

	if len(s.Errors) > 0 {
		b.WriteString("\nErrors:\n")
		b.WriteString(strings.Repeat("-", 40))
		b.WriteString("\n")
		for _, err := range s.Errors {
			b.WriteString(fmt.Sprintf("  %s: %s\n", err.Hostname, err.Error))
		}
	}

	return b.String()
}

func (s *FleetSummary) TableRows() [][]string {
	var rows [][]string
	rows = append(rows, []string{"Total", fmt.Sprintf("%d", s.TotalHosts)})
	rows = append(rows, []string{"Successful", fmt.Sprintf("%d", s.SuccessfulHosts)})
	rows = append(rows, []string{"Failed", fmt.Sprintf("%d", s.FailedHosts)})
	rows = append(rows, []string{"Rate", fmt.Sprintf("%.1f%%", s.SuccessRate)})
	rows = append(rows, []string{"Duration", s.TotalDuration.String()})
	rows = append(rows, []string{"Average", s.AverageDuration.String()})
	if s.SlowestHost != "" {
		rows = append(rows, []string{"Slowest", fmt.Sprintf("%s (%v)", s.SlowestHost, s.SlowestDuration)})
	}
	if s.FastestHost != "" {
		rows = append(rows, []string{"Fastest", fmt.Sprintf("%s (%v)", s.FastestHost, s.FastestDuration)})
	}
	return rows
}

func GroupResultsByStatus(results []HostResult) (successful []HostResult, failed []HostResult) {
	for _, r := range results {
		if r.Success {
			successful = append(successful, r)
		} else {
			failed = append(failed, r)
		}
	}
	return
}

func SortByDuration(results []HostResult, ascending bool) {
	sort.Slice(results, func(i, j int) bool {
		if ascending {
			return results[i].Duration < results[j].Duration
		}
		return results[i].Duration > results[j].Duration
	})
}

func FilterErrors(results []HostResult) []HostError {
	var errors []HostError
	for _, r := range results {
		if !r.Success && r.Error != nil {
			errors = append(errors, HostError{
				Hostname: r.Hostname,
				Error:    r.Error.Error(),
			})
		}
	}
	return errors
}
