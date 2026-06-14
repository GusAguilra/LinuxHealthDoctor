package disk

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/GusAguilra/LinuxHealthDoctor/internal/core"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/plugin"
	"github.com/shirou/gopsutil/v3/disk"
)

func init() {
	plugin.Register(&DiskUsageCheck{})
	plugin.Register(&DiskIOCheck{})
}

type DiskUsageCheck struct{}

func (c *DiskUsageCheck) ID() string               { return "disk.usage" }
func (c *DiskUsageCheck) Name() string             { return "Disk Usage" }
func (c *DiskUsageCheck) Description() string      { return "Checks disk partition usage percentages" }
func (c *DiskUsageCheck) Category() core.Component { return core.ComponentDisk }
func (c *DiskUsageCheck) Dependencies() []string   { return nil }
func (c *DiskUsageCheck) Tags() []string           { return []string{"storage", "resource", "disk"} }

func (c *DiskUsageCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		result.Status = core.StatusError
		result.Error = fmt.Errorf("failed to list disk partitions: %w", err)
		return result, nil
	}

	result.Metrics = make(map[string]float64)
	result.Details = map[string]interface{}{
		"partitions": []map[string]interface{}{},
	}
	partitionDetails := []map[string]interface{}{}
	hasIssues := false
	maxUsage := 0.0
	criticalMount := ""

	skipped := 0
	for _, p := range partitions {
		if strings.HasPrefix(p.Device, "/dev/loop") || p.Fstype == "squashfs" || p.Fstype == "tmpfs" {
			skipped++
			continue
		}
		usage, err := disk.UsageWithContext(ctx, p.Mountpoint)
		if err != nil {
			skipped++
			continue
		}
		usagePct := usage.UsedPercent
		partitionInfo := map[string]interface{}{
			"mountpoint":    p.Mountpoint,
			"device":        p.Device,
			"fstype":        p.Fstype,
			"usage_percent": fmt.Sprintf("%.1f%%", usagePct),
			"total_gb":      fmt.Sprintf("%.1f", float64(usage.Total)/1024/1024/1024),
			"used_gb":       fmt.Sprintf("%.1f", float64(usage.Used)/1024/1024/1024),
			"free_gb":       fmt.Sprintf("%.1f", float64(usage.Free)/1024/1024/1024),
		}
		partitionDetails = append(partitionDetails, partitionInfo)
		result.Metrics["usage_"+p.Mountpoint] = usagePct

		if usagePct > maxUsage {
			maxUsage = usagePct
			criticalMount = p.Mountpoint
		}

		switch {
		case usagePct > 95:
			hasIssues = true
			if result.Severity < core.SeverityCritical {
				result.Severity = core.SeverityCritical
			}
		case usagePct > 85:
			hasIssues = true
			if result.Severity < core.SeverityWarning {
				result.Severity = core.SeverityWarning
			}
		}
	}

	result.Details["partitions"] = partitionDetails

	if hasIssues {
		result.Status = core.StatusFail
		result.Message = fmt.Sprintf("Disk usage is critical on %s: %.1f%%", criticalMount, maxUsage)
	} else {
		result.Message = "All disk partitions are within normal usage limits"
	}
	if skipped > 0 {
		result.Message += fmt.Sprintf(" (%d partitions skipped)", skipped)
	}

	return result, nil
}

type DiskIOCheck struct{}

func (c *DiskIOCheck) ID() string               { return "disk.io" }
func (c *DiskIOCheck) Name() string             { return "Disk I/O" }
func (c *DiskIOCheck) Description() string      { return "Checks disk I/O statistics" }
func (c *DiskIOCheck) Category() core.Component { return core.ComponentDisk }
func (c *DiskIOCheck) Dependencies() []string   { return nil }
func (c *DiskIOCheck) Tags() []string           { return []string{"storage", "performance", "disk"} }

func (c *DiskIOCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	counters, err := disk.IOCountersWithContext(ctx)
	if err != nil {
		result.Status = core.StatusError
		result.Error = fmt.Errorf("failed to read disk I/O counters: %w", err)
		return result, nil
	}

	result.Metrics = make(map[string]float64)
	result.Details = map[string]interface{}{
		"devices": []map[string]interface{}{},
	}
	deviceDetails := []map[string]interface{}{}

	totalIO := uint64(0)
	for name, io := range counters {
		deviceDetails = append(deviceDetails, map[string]interface{}{
			"device":        name,
			"read_count":    io.ReadCount,
			"write_count":   io.WriteCount,
			"read_bytes":    io.ReadBytes,
			"write_bytes":   io.WriteBytes,
			"read_time_ms":  io.ReadTime,
			"write_time_ms": io.WriteTime,
		})
		result.Metrics[name+"_read_count"] = float64(io.ReadCount)
		result.Metrics[name+"_write_count"] = float64(io.WriteCount)
		result.Metrics[name+"_read_bytes"] = float64(io.ReadBytes)
		result.Metrics[name+"_write_bytes"] = float64(io.WriteBytes)
		totalIO += io.ReadCount + io.WriteCount
	}

	result.Details["devices"] = deviceDetails
	result.Message = fmt.Sprintf("Disk I/O check completed for %d devices", len(counters))

	return result, nil
}
