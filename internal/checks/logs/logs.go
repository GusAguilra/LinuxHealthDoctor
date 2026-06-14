package logs

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/linuxhealthdoctor/lhd/internal/core"
	"github.com/linuxhealthdoctor/lhd/internal/plugin"
)

func init() {
	plugin.Register(&LogsErrorsCheck{})
	plugin.Register(&LogsJournalCheck{})
}

type LogsErrorsCheck struct{}

func (c *LogsErrorsCheck) ID() string               { return "logs.errors" }
func (c *LogsErrorsCheck) Name() string             { return "Recent Log Errors" }
func (c *LogsErrorsCheck) Description() string      { return "Checks for recent error log entries" }
func (c *LogsErrorsCheck) Category() core.Component { return core.ComponentLogs }
func (c *LogsErrorsCheck) Dependencies() []string   { return nil }
func (c *LogsErrorsCheck) Tags() []string           { return []string{"logs", "errors", "monitoring"} }

func (c *LogsErrorsCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	cmd := exec.CommandContext(ctx, "journalctl", "--since", "1 hour ago", "-p", "err", "--no-pager", "-n", "100")
	output, err := cmd.Output()
	if err != nil {
		result.Status = core.StatusError
		result.Error = fmt.Errorf("failed to read journalctl errors: %w", err)
		return result, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	errorCount := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			errorCount++
		}
	}

	result.Metrics = map[string]float64{
		"errors_last_hour": float64(errorCount),
	}
	result.Details = map[string]interface{}{
		"error_count": errorCount,
		"time_range":  "1 hour",
	}

	switch {
	case errorCount > 50:
		result.Status = core.StatusFail
		result.Severity = core.SeverityCritical
		result.Message = fmt.Sprintf("Critical: %d error log entries in the last hour", errorCount)
	case errorCount > 10:
		result.Status = core.StatusFail
		result.Severity = core.SeverityWarning
		result.Message = fmt.Sprintf("Elevated error count: %d log errors in the last hour", errorCount)
	default:
		result.Message = fmt.Sprintf("%d log errors in the last hour", errorCount)
	}

	return result, nil
}

type LogsJournalCheck struct{}

func (c *LogsJournalCheck) ID() string               { return "logs.journal" }
func (c *LogsJournalCheck) Name() string             { return "Journal Health" }
func (c *LogsJournalCheck) Description() string      { return "Checks systemd journal health and disk usage" }
func (c *LogsJournalCheck) Category() core.Component { return core.ComponentLogs }
func (c *LogsJournalCheck) Dependencies() []string   { return nil }
func (c *LogsJournalCheck) Tags() []string           { return []string{"logs", "journal", "systemd"} }

func (c *LogsJournalCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	cmd := exec.CommandContext(ctx, "journalctl", "--disk-usage")
	output, err := cmd.Output()
	if err != nil {
		result.Status = core.StatusError
		result.Error = fmt.Errorf("failed to check journal disk usage: %w", err)
		return result, nil
	}

	usageLine := strings.TrimSpace(string(output))
	result.Details = map[string]interface{}{
		"disk_usage": usageLine,
	}
	result.Message = fmt.Sprintf("Journal disk usage: %s", usageLine)

	cmd = exec.CommandContext(ctx, "journalctl", "--list-boots", "--no-pager", "-n", "1")
	if _, err := cmd.Output(); err != nil {
		result.Status = core.StatusFail
		result.Severity = core.SeverityWarning
		result.Message = "Journal is not accessible"
		return result, nil
	}

	result.Details["verified"] = true

	return result, nil
}
