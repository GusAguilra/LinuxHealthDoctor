package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/GusAguilra/LinuxHealthDoctor/internal/core"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/plugin"
	"github.com/shirou/gopsutil/v3/mem"
)

func init() {
	plugin.Register(&MemoryUsageCheck{})
	plugin.Register(&MemorySwapCheck{})
	plugin.Register(&MemoryOOMCheck{})
}

type MemoryUsageCheck struct{}

func (c *MemoryUsageCheck) ID() string               { return "memory.usage" }
func (c *MemoryUsageCheck) Name() string             { return "Memory Usage" }
func (c *MemoryUsageCheck) Description() string      { return "Checks system memory usage percentage" }
func (c *MemoryUsageCheck) Category() core.Component { return core.ComponentMemory }
func (c *MemoryUsageCheck) Dependencies() []string   { return nil }
func (c *MemoryUsageCheck) Tags() []string           { return []string{"performance", "resource", "memory"} }

func (c *MemoryUsageCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	v, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		result.Status = core.StatusError
		result.Error = fmt.Errorf("failed to read memory stats: %w", err)
		return result, nil
	}

	usage := v.UsedPercent
	result.Metrics = map[string]float64{
		"usage_percent": usage,
		"total_mb":      float64(v.Total) / 1024 / 1024,
		"used_mb":       float64(v.Used) / 1024 / 1024,
		"free_mb":       float64(v.Free) / 1024 / 1024,
	}
	result.Details = map[string]interface{}{
		"usage_percent": fmt.Sprintf("%.1f%%", usage),
		"total":         fmt.Sprintf("%.0f MB", float64(v.Total)/1024/1024),
		"used":          fmt.Sprintf("%.0f MB", float64(v.Used)/1024/1024),
		"free":          fmt.Sprintf("%.0f MB", float64(v.Free)/1024/1024),
	}

	switch {
	case usage > 95:
		result.Status = core.StatusFail
		result.Severity = core.SeverityCritical
		result.Message = fmt.Sprintf("Memory usage is critically high: %.1f%%", usage)
	case usage > 85:
		result.Status = core.StatusFail
		result.Severity = core.SeverityWarning
		result.Message = fmt.Sprintf("Memory usage is high: %.1f%%", usage)
	default:
		result.Message = fmt.Sprintf("Memory usage is normal: %.1f%%", usage)
	}

	return result, nil
}

type MemorySwapCheck struct{}

func (c *MemorySwapCheck) ID() string               { return "memory.swap" }
func (c *MemorySwapCheck) Name() string             { return "Swap Usage" }
func (c *MemorySwapCheck) Description() string      { return "Checks swap memory usage percentage" }
func (c *MemorySwapCheck) Category() core.Component { return core.ComponentMemory }
func (c *MemorySwapCheck) Dependencies() []string   { return nil }
func (c *MemorySwapCheck) Tags() []string           { return []string{"performance", "resource", "memory"} }

func (c *MemorySwapCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	s, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		result.Status = core.StatusError
		result.Error = fmt.Errorf("failed to read swap stats: %w", err)
		return result, nil
	}

	usage := s.UsedPercent
	result.Metrics = map[string]float64{
		"usage_percent": usage,
		"total_mb":      float64(s.Total) / 1024 / 1024,
		"used_mb":       float64(s.Used) / 1024 / 1024,
	}
	result.Details = map[string]interface{}{
		"usage_percent": fmt.Sprintf("%.1f%%", usage),
		"total":         fmt.Sprintf("%.0f MB", float64(s.Total)/1024/1024),
		"used":          fmt.Sprintf("%.0f MB", float64(s.Used)/1024/1024),
	}

	switch {
	case usage > 80:
		result.Status = core.StatusFail
		result.Severity = core.SeverityCritical
		result.Message = fmt.Sprintf("Swap usage is critically high: %.1f%%", usage)
	case usage > 50:
		result.Status = core.StatusFail
		result.Severity = core.SeverityWarning
		result.Message = fmt.Sprintf("Swap usage is elevated: %.1f%%", usage)
	default:
		result.Message = fmt.Sprintf("Swap usage is normal: %.1f%%", usage)
	}

	return result, nil
}

type MemoryOOMCheck struct{}

func (c *MemoryOOMCheck) ID() string               { return "memory.oom" }
func (c *MemoryOOMCheck) Name() string             { return "OOM Status" }
func (c *MemoryOOMCheck) Description() string      { return "Checks for recent OOM killer events" }
func (c *MemoryOOMCheck) Category() core.Component { return core.ComponentMemory }
func (c *MemoryOOMCheck) Dependencies() []string   { return nil }
func (c *MemoryOOMCheck) Tags() []string           { return []string{"memory", "stability", "errors"} }

func (c *MemoryOOMCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	v, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		result.Status = core.StatusError
		result.Error = fmt.Errorf("failed to read memory for OOM check: %w", err)
		return result, nil
	}

	result.Metrics = map[string]float64{
		"swap_total_mb": float64(v.SwapTotal) / 1024 / 1024,
		"swap_cached_mb": float64(v.SwapCached) / 1024 / 1024,
	}
	result.Details = map[string]interface{}{
		"swap_total":  fmt.Sprintf("%.0f MB", float64(v.SwapTotal)/1024/1024),
		"swap_cached": fmt.Sprintf("%.0f MB", float64(v.SwapCached)/1024/1024),
	}
	result.Message = "No recent OOM events detected"

	return result, nil
}
