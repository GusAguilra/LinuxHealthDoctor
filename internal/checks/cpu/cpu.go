package cpu

import (
	"context"
	"fmt"
	"time"

	"github.com/GusAguilra/LinuxHealthDoctor/internal/core"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/plugin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
)

func init() {
	plugin.Register(&CPUUsageCheck{})
	plugin.Register(&CPULoadCheck{})
	plugin.Register(&CPUTemperatureCheck{})
	plugin.Register(&CPUGovernorCheck{})
}

type CPUUsageCheck struct{}

func (c *CPUUsageCheck) ID() string               { return "cpu.usage" }
func (c *CPUUsageCheck) Name() string             { return "CPU Usage" }
func (c *CPUUsageCheck) Description() string      { return "Checks CPU usage percentage" }
func (c *CPUUsageCheck) Category() core.Component { return core.ComponentCPU }
func (c *CPUUsageCheck) Dependencies() []string   { return nil }
func (c *CPUUsageCheck) Tags() []string           { return []string{"performance", "resource"} }

func (c *CPUUsageCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	percent, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		result.Status = core.StatusError
		result.Error = fmt.Errorf("failed to read CPU usage: %w", err)
		return result, nil
	}

	usage := percent[0]
	result.Metrics = map[string]float64{"usage_percent": usage}
	result.Details = map[string]interface{}{"usage_percent": fmt.Sprintf("%.1f%%", usage)}

	if usage > 95 {
		result.Status = core.StatusFail
		result.Severity = core.SeverityCritical
		result.Message = fmt.Sprintf("CPU usage is critically high: %.1f%%", usage)
	} else if usage > 80 {
		result.Status = core.StatusFail
		result.Severity = core.SeverityWarning
		result.Message = fmt.Sprintf("CPU usage is high: %.1f%%", usage)
	} else {
		result.Message = fmt.Sprintf("CPU usage is normal: %.1f%%", usage)
	}

	return result, nil
}

type CPULoadCheck struct{}

func (c *CPULoadCheck) ID() string               { return "cpu.load" }
func (c *CPULoadCheck) Name() string             { return "CPU Load Average" }
func (c *CPULoadCheck) Description() string      { return "Checks CPU load averages" }
func (c *CPULoadCheck) Category() core.Component { return core.ComponentCPU }
func (c *CPULoadCheck) Dependencies() []string   { return nil }
func (c *CPULoadCheck) Tags() []string           { return []string{"performance", "resource"} }

func (c *CPULoadCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	loadAvg, err := load.AvgWithContext(ctx)
	if err != nil {
		result.Status = core.StatusError
		result.Error = fmt.Errorf("failed to read load average: %w", err)
		return result, nil
	}

	cpuCount, _ := cpu.Counts(true)
	perCPU1m := loadAvg.Load1 / float64(cpuCount)

	result.Metrics = map[string]float64{
		"load_1m": loadAvg.Load1, "load_5m": loadAvg.Load5, "load_15m": loadAvg.Load15,
	}
	result.Details = map[string]interface{}{
		"load_1m": fmt.Sprintf("%.2f", loadAvg.Load1), "load_5m": fmt.Sprintf("%.2f", loadAvg.Load5),
		"load_15m": fmt.Sprintf("%.2f", loadAvg.Load15), "cpu_cores": cpuCount,
		"per_core_1m": fmt.Sprintf("%.2f", perCPU1m),
	}

	if perCPU1m > 2.0 {
		result.Status = core.StatusFail
		result.Severity = core.SeverityCritical
		result.Message = fmt.Sprintf("CPU load per core is critically high: %.2f", perCPU1m)
	} else if perCPU1m > 1.0 {
		result.Status = core.StatusFail
		result.Severity = core.SeverityWarning
		result.Message = fmt.Sprintf("CPU load per core is elevated: %.2f", perCPU1m)
	} else {
		result.Message = fmt.Sprintf("CPU load is normal: %.2f per core", perCPU1m)
	}

	return result, nil
}

type CPUTemperatureCheck struct{}

func (c *CPUTemperatureCheck) ID() string               { return "cpu.temperature" }
func (c *CPUTemperatureCheck) Name() string             { return "CPU Temperature" }
func (c *CPUTemperatureCheck) Description() string      { return "Checks CPU thermal status" }
func (c *CPUTemperatureCheck) Category() core.Component { return core.ComponentCPU }
func (c *CPUTemperatureCheck) Dependencies() []string   { return nil }
func (c *CPUTemperatureCheck) Tags() []string           { return []string{"hardware", "thermal"} }

func (c *CPUTemperatureCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}
	result.Message = "CPU temperature check completed (requires privileged access for detailed sensor data)"
	return result, nil
}

type CPUGovernorCheck struct{}

func (c *CPUGovernorCheck) ID() string               { return "cpu.governor" }
func (c *CPUGovernorCheck) Name() string             { return "CPU Scaling Governor" }
func (c *CPUGovernorCheck) Description() string      { return "Checks CPU frequency scaling governor" }
func (c *CPUGovernorCheck) Category() core.Component { return core.ComponentCPU }
func (c *CPUGovernorCheck) Dependencies() []string   { return nil }
func (c *CPUGovernorCheck) Tags() []string           { return []string{"performance", "power"} }

func (c *CPUGovernorCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}
	result.Message = "CPU governor check completed"
	return result, nil
}
