package report

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type ANSITerminalFormatter struct {
	Color bool
}

func (f *ANSITerminalFormatter) Format(ctx context.Context, data *ReportData, opts *FormatOptions) ([]byte, error) {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("=== %s ===\n", data.Title))
	b.WriteString(fmt.Sprintf("Generated: %s\n", data.GeneratedAt.Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("Host: %s | Distro: %s | Kernel: %s\n", data.Hostname, data.Distro, data.Kernel))

	if data.HealthScore != nil {
		b.WriteString(fmt.Sprintf("\nHealth Score: %.1f/100\n", data.HealthScore.Overall))
		b.WriteString(fmt.Sprintf("  Passed: %d | Failed: %d | Total: %d\n",
			data.HealthScore.Passed, data.HealthScore.Failed, data.HealthScore.Total))
	}

	if data.Results != nil {
		b.WriteString("\n--- Check Results ---\n")
		for _, r := range data.Results.AllResults {
			status := r.Status.String()
			sev := r.Severity.String()
			b.WriteString(fmt.Sprintf("[%s] [%s] %s: %s\n", status, sev, r.Category, r.Message))
		}
	}

	return []byte(b.String()), nil
}

func (f *ANSITerminalFormatter) Extension() string { return ".txt" }
func (f *ANSITerminalFormatter) MIMEType() string  { return "text/plain" }

type JSONFormatter struct {
	Pretty bool
}

func (f *JSONFormatter) Format(ctx context.Context, data *ReportData, opts *FormatOptions) ([]byte, error) {
	var b []byte
	var err error
	if f.Pretty {
		b, err = json.MarshalIndent(data, "", "  ")
	} else {
		b, err = json.Marshal(data)
	}
	if err != nil {
		return nil, fmt.Errorf("json marshal: %w", err)
	}
	return b, nil
}

func (f *JSONFormatter) Extension() string { return ".json" }
func (f *JSONFormatter) MIMEType() string  { return "application/json" }
