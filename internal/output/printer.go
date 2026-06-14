package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/GusAguilra/LinuxHealthDoctor/internal/core"
)

type Printer struct {
	Color   bool
	Verbose bool
}

func NewPrinter(color bool) *Printer {
	return &Printer{Color: color}
}

func (p *Printer) PrintResult(result *core.AggregatedResult) {
	fmt.Println(p.FormatResult(result))
}

func (p *Printer) FormatResult(result *core.AggregatedResult) string {
	var b strings.Builder

	b.WriteString("\n=== Linux Health Doctor ===\n")
	b.WriteString(fmt.Sprintf("Health Score: %.1f/100\n", result.HealthScore))
	b.WriteString(fmt.Sprintf("Checks: %d total | %d passed | %d failed | %d errors | %d skipped\n",
		result.TotalChecks, result.PassedChecks, result.FailedChecks,
		result.ErrorChecks, result.SkippedChecks))
	b.WriteString(fmt.Sprintf("Duration: %s\n", result.Duration.Round(time.Millisecond)))
	b.WriteString("\n")

	for _, component := range core.AllComponents() {
		results, ok := result.Results[component]
		if !ok || len(results) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("--- %s ---\n", strings.ToUpper(string(component))))
		for _, r := range results {
			status := formatStatus(r.Status)
			sev := formatSeverity(r.Severity)
			b.WriteString(fmt.Sprintf("  %s %s %s\n", status, sev, r.Message))
			if p.Verbose && len(r.Metrics) > 0 {
				for k, v := range r.Metrics {
					b.WriteString(fmt.Sprintf("    %s: %.2f\n", k, v))
				}
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

func formatStatus(s core.CheckStatus) string {
	switch s {
	case core.StatusPass:
		return "PASS"
	case core.StatusFail:
		return "FAIL"
	case core.StatusError:
		return "ERROR"
	case core.StatusSkip:
		return "SKIP"
	default:
		return "UNKN"
	}
}

func formatSeverity(s core.Severity) string {
	switch s {
	case core.SeverityCritical:
		return "CRITICAL"
	case core.SeverityWarning:
		return "WARNING"
	case core.SeverityInfo:
		return "INFO"
	default:
		return ""
	}
}
