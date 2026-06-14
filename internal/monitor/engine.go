package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/linuxhealthdoctor/lhd/internal/core"
	"github.com/linuxhealthdoctor/lhd/internal/storage"
	gopsutil_cpu "github.com/shirou/gopsutil/v3/cpu"
	gopsutil_disk "github.com/shirou/gopsutil/v3/disk"
	gopsutil_mem "github.com/shirou/gopsutil/v3/mem"
)

type Engine struct {
	mu         sync.Mutex
	running    bool
	stopCh     chan struct{}
	interval   time.Duration
	store      storage.Store
	eventBus   core.EventBus
	thresholds ThresholdMap
	alerts     []Alert
}

type ThresholdMap map[string]ComponentThresholds

type ComponentThresholds struct {
	Warning  float64
	Critical float64
}

type Alert struct {
	ID        string
	Severity  core.Severity
	Message   string
	Metric    string
	Value     float64
	Threshold float64
	Timestamp time.Time
	Acknowledged bool
}

func NewEngine(store storage.Store, eventBus core.EventBus) *Engine {
	return &Engine{
		stopCh:   make(chan struct{}),
		store:    store,
		eventBus: eventBus,
		thresholds: defaultThresholds(),
	}
}

func defaultThresholds() ThresholdMap {
	return ThresholdMap{
		"cpu.usage_percent":    {Warning: 80, Critical: 95},
		"memory.usage_percent": {Warning: 80, Critical: 95},
		"disk.usage_percent":   {Warning: 80, Critical: 92},
	}
}

func (e *Engine) Start(ctx context.Context, interval time.Duration) error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return fmt.Errorf("monitor is already running")
	}
	e.running = true
	e.interval = interval
	e.mu.Unlock()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.collect(ctx)
		case <-e.stopCh:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.running {
		close(e.stopCh)
		e.running = false
	}
}

func (e *Engine) IsRunning() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.running
}

func (e *Engine) collect(ctx context.Context) {
	metrics := make(map[string]float64)

	metrics["cpu.usage_percent"] = readCPUUsage(ctx)
	metrics["memory.usage_percent"] = readMemoryUsage(ctx)
	metrics["disk.usage_percent"] = readDiskUsage(ctx)

	for name, value := range metrics {
		m := &core.Metric{
			Name:      name,
			Value:     value,
			Timestamp: time.Now(),
		}
		if err := e.store.WriteMetric(ctx, m); err != nil {
			continue
		}
		e.evaluateThreshold(name, value)
	}
}

func (e *Engine) evaluateThreshold(name string, value float64) {
	thresholds, ok := e.thresholds[name]
	if !ok {
		return
	}

	var severity core.Severity
	var threshold float64

	if value >= thresholds.Critical {
		severity = core.SeverityCritical
		threshold = thresholds.Critical
	} else if value >= thresholds.Warning {
		severity = core.SeverityWarning
		threshold = thresholds.Warning
	} else {
		return
	}

	alert := Alert{
		ID:        fmt.Sprintf("alert-%d", time.Now().UnixNano()),
		Severity:  severity,
		Message:   fmt.Sprintf("%s: %.1f (threshold: %.1f)", name, value, threshold),
		Metric:    name,
		Value:     value,
		Threshold: threshold,
		Timestamp: time.Now(),
	}

	e.mu.Lock()
	e.alerts = append(e.alerts, alert)
	e.mu.Unlock()

	e.eventBus.Publish(string(core.EventThreshold), &core.Event{
		Type:     core.EventThreshold,
		Source:   "monitor",
		Severity: severity,
		Message:  alert.Message,
		Data: map[string]interface{}{
			"metric":    name,
			"value":     value,
			"threshold": threshold,
		},
	})
}

func (e *Engine) Alerts() []Alert {
	e.mu.Lock()
	defer e.mu.Unlock()
	result := make([]Alert, len(e.alerts))
	copy(result, e.alerts)
	return result
}

func readCPUUsage(ctx context.Context) float64 {
	cpu, err := gopsutil_cpu.PercentWithContext(ctx, 0, false)
	if err != nil || len(cpu) == 0 {
		return 0.0
	}
	return cpu[0]
}

func readMemoryUsage(ctx context.Context) float64 {
	mem, err := gopsutil_mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return 0.0
	}
	return mem.UsedPercent
}

func readDiskUsage(ctx context.Context) float64 {
	partitions, err := gopsutil_disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return 0.0
	}
	total := uint64(0)
	used := uint64(0)
	for _, p := range partitions {
		usage, err := gopsutil_disk.UsageWithContext(ctx, p.Mountpoint)
		if err != nil {
			continue
		}
		total += usage.Total
		used += usage.Used
	}
	if total == 0 {
		return 0.0
	}
	return float64(used) / float64(total) * 100.0
}
