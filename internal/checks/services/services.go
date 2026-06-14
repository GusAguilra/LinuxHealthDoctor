package services

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/GusAguilra/LinuxHealthDoctor/internal/core"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/plugin"
)

func init() {
	plugin.Register(&ServicesFailedCheck{})
	plugin.Register(&ServicesCriticalCheck{})
}

type ServicesFailedCheck struct{}

func (c *ServicesFailedCheck) ID() string               { return "services.failed" }
func (c *ServicesFailedCheck) Name() string             { return "Failed Services" }
func (c *ServicesFailedCheck) Description() string      { return "Checks for failed systemd units" }
func (c *ServicesFailedCheck) Category() core.Component { return core.ComponentServices }
func (c *ServicesFailedCheck) Dependencies() []string   { return nil }
func (c *ServicesFailedCheck) Tags() []string           { return []string{"services", "systemd", "stability"} }

func (c *ServicesFailedCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	cmd := exec.CommandContext(ctx, "systemctl", "list-units", "--state=failed", "--no-legend", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		result.Status = core.StatusError
		result.Error = fmt.Errorf("failed to list failed services: %w", err)
		return result, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	failedUnits := []string{}
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		fields := strings.Fields(l)
		if len(fields) > 0 {
			failedUnits = append(failedUnits, fields[0])
		}
	}

	result.Metrics = map[string]float64{
		"failed_units": float64(len(failedUnits)),
	}
	result.Details = map[string]interface{}{
		"failed_units": failedUnits,
	}

	switch {
	case len(failedUnits) > 5:
		result.Status = core.StatusFail
		result.Severity = core.SeverityCritical
		result.Message = fmt.Sprintf("%d failed systemd units detected", len(failedUnits))
	case len(failedUnits) > 0:
		result.Status = core.StatusFail
		result.Severity = core.SeverityWarning
		result.Message = fmt.Sprintf("%d failed systemd units: %s", len(failedUnits), strings.Join(failedUnits, ", "))
	default:
		result.Message = "No failed systemd units"
	}

	return result, nil
}

type ServicesCriticalCheck struct{}

func (c *ServicesCriticalCheck) ID() string               { return "services.critical" }
func (c *ServicesCriticalCheck) Name() string             { return "Critical Services" }
func (c *ServicesCriticalCheck) Description() string      { return "Checks that critical services are running" }
func (c *ServicesCriticalCheck) Category() core.Component { return core.ComponentServices }
func (c *ServicesCriticalCheck) Dependencies() []string   { return nil }
func (c *ServicesCriticalCheck) Tags() []string           { return []string{"services", "systemd", "reliability"} }

func (c *ServicesCriticalCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	criticalServices := []string{
		"sshd.service",
		"systemd-journald.service",
		"systemd-logind.service",
		"dbus.service",
		"NetworkManager.service",
	}

	stopped := []string{}
	running := 0

	for _, svc := range criticalServices {
		cmd := exec.CommandContext(ctx, "systemctl", "is-active", svc)
		output, err := cmd.Output()
		if err != nil {
			stopped = append(stopped, svc)
			continue
		}
		status := strings.TrimSpace(string(output))
		if status == "active" {
			running++
		} else {
			stopped = append(stopped, svc)
		}
	}

	result.Metrics = map[string]float64{
		"critical_total":  float64(len(criticalServices)),
		"critical_running": float64(running),
		"critical_stopped": float64(len(stopped)),
	}
	result.Details = map[string]interface{}{
		"critical_services": criticalServices,
		"running":           running,
		"stopped":           stopped,
	}

	if len(stopped) > 0 {
		result.Status = core.StatusFail
		result.Severity = core.SeverityCritical
		result.Message = fmt.Sprintf("%d critical services are not running: %s", len(stopped), strings.Join(stopped, ", "))
	} else {
		result.Message = fmt.Sprintf("All %d critical services are running", len(criticalServices))
	}

	return result, nil
}
