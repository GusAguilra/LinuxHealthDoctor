package kernel

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
	plugin.Register(&KernelVersionCheck{})
	plugin.Register(&KernelDmesgCheck{})
	plugin.Register(&KernelSysctlCheck{})
}

type KernelVersionCheck struct{}

func (c *KernelVersionCheck) ID() string               { return "kernel.version" }
func (c *KernelVersionCheck) Name() string             { return "Kernel Version" }
func (c *KernelVersionCheck) Description() string      { return "Checks kernel version meets minimum requirements" }
func (c *KernelVersionCheck) Category() core.Component { return core.ComponentKernel }
func (c *KernelVersionCheck) Dependencies() []string   { return nil }
func (c *KernelVersionCheck) Tags() []string           { return []string{"kernel", "version", "security"} }

func (c *KernelVersionCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	cmd := exec.CommandContext(ctx, "uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		result.Status = core.StatusError
		result.Error = fmt.Errorf("failed to get kernel version: %w", err)
		return result, nil
	}

	version := strings.TrimSpace(string(output))
	result.Metrics = map[string]float64{}
	result.Details = map[string]interface{}{
		"kernel_version": version,
	}
	result.Message = fmt.Sprintf("Kernel version: %s", version)

	return result, nil
}

type KernelDmesgCheck struct{}

func (c *KernelDmesgCheck) ID() string               { return "kernel.dmesg" }
func (c *KernelDmesgCheck) Name() string             { return "Kernel Dmesg" }
func (c *KernelDmesgCheck) Description() string      { return "Scans kernel ring buffer for errors" }
func (c *KernelDmesgCheck) Category() core.Component { return core.ComponentKernel }
func (c *KernelDmesgCheck) Dependencies() []string   { return nil }
func (c *KernelDmesgCheck) Tags() []string           { return []string{"kernel", "errors", "stability"} }

func (c *KernelDmesgCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	cmd := exec.CommandContext(ctx, "dmesg", "--level=err,warn")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.CommandContext(ctx, "dmesg")
		output, err = cmd.Output()
		if err != nil {
			result.Status = core.StatusError
			result.Error = fmt.Errorf("failed to read dmesg: %w", err)
			return result, nil
		}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	nonEmpty := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			nonEmpty++
		}
	}

	if len(lines) > 100 {
		lines = lines[len(lines)-100:]
	}

	result.Metrics = map[string]float64{
		"error_lines": float64(nonEmpty),
	}
	result.Details = map[string]interface{}{
		"error_count": nonEmpty,
		"log_sample":  []string{},
	}

	errorStrings := []string{"error", "panic", "bug", "oops", "failed"}
	criticalErrors := 0
	for _, l := range lines {
		lower := strings.ToLower(l)
		for _, e := range errorStrings {
			if strings.Contains(lower, e) {
				criticalErrors++
				break
			}
		}
	}

	switch {
	case criticalErrors > 10:
		result.Status = core.StatusFail
		result.Severity = core.SeverityCritical
		result.Message = fmt.Sprintf("Kernel ring buffer has %d critical errors", criticalErrors)
	case nonEmpty > 50:
		result.Status = core.StatusFail
		result.Severity = core.SeverityWarning
		result.Message = fmt.Sprintf("Kernel ring buffer has %d warnings/errors", nonEmpty)
	default:
		result.Message = fmt.Sprintf("Kernel ring buffer is clean (%d non-error messages)", nonEmpty)
	}

	if nonEmpty > 0 && nonEmpty <= 10 {
		result.Details["log_sample"] = lines[:nonEmpty]
	} else if nonEmpty > 10 {
		result.Details["log_sample"] = lines[:10]
	}

	return result, nil
}

type KernelSysctlCheck struct{}

func (c *KernelSysctlCheck) ID() string               { return "kernel.sysctl" }
func (c *KernelSysctlCheck) Name() string             { return "Sysctl Anomalies" }
func (c *KernelSysctlCheck) Description() string      { return "Checks for sysctl anomalies and insecure settings" }
func (c *KernelSysctlCheck) Category() core.Component { return core.ComponentKernel }
func (c *KernelSysctlCheck) Dependencies() []string   { return nil }
func (c *KernelSysctlCheck) Tags() []string           { return []string{"kernel", "security", "configuration"} }

func (c *KernelSysctlCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	checks := map[string]string{
		"net.ipv4.ip_forward":      "0",
		"net.ipv4.conf.all.rp_filter": "1",
		"kernel.dmesg_restrict":    "1",
	}

	anomalies := []string{}
	failedParams := 0
	for param, expected := range checks {
		cmd := exec.CommandContext(ctx, "sysctl", "-n", param)
		output, err := cmd.Output()
		if err != nil {
			failedParams++
			continue
		}
		val := strings.TrimSpace(string(output))
		if val != expected {
			anomalies = append(anomalies, fmt.Sprintf("%s=%s (expected %s)", param, val, expected))
		}
	}

	result.Metrics = map[string]float64{
		"checks_performed": float64(len(checks)),
		"anomalies_found":  float64(len(anomalies)),
	}
	result.Details = map[string]interface{}{
		"parameters_checked": len(checks),
		"anomalies":          anomalies,
		"sysctl_errors":      failedParams,
	}

	if len(anomalies) > 0 {
		result.Status = core.StatusFail
		result.Severity = core.SeverityWarning
		result.Message = fmt.Sprintf("Found %d sysctl setting anomalies", len(anomalies))
	} else {
		result.Message = "All checked sysctl parameters are within expected values"
	}
	if failedParams > 0 {
		result.Message += fmt.Sprintf(" (%d parameters could not be read)", failedParams)
	}

	return result, nil
}
