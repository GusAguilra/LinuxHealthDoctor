package fleet

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type FleetEngine struct {
	inventory *Inventory
	sshPool   *SSHPool
	config    EngineConfig
}

type EngineConfig struct {
	Concurrency int
	Timeout     time.Duration
	Retry       int
}

type FleetResult struct {
	Timestamp  time.Time
	Hosts      int
	Successful int
	Failed     int
	Results    []HostResult
}

type HostResult struct {
	Hostname string
	Success  bool
	Error    error
	Duration time.Duration
	Stdout   string
	Stderr   string
}

func NewFleetEngine(inventory *Inventory, sshConfig SSHConfig, engineConfig EngineConfig) *FleetEngine {
	if engineConfig.Concurrency <= 0 {
		engineConfig.Concurrency = 10
	}
	if engineConfig.Timeout <= 0 {
		engineConfig.Timeout = 30 * time.Second
	}
	if engineConfig.Retry <= 0 {
		engineConfig.Retry = 2
	}

	return &FleetEngine{
		inventory: inventory,
		sshPool:   NewSSHPool(sshConfig),
		config:    engineConfig,
	}
}

func (e *FleetEngine) Close() {
	e.sshPool.Close()
}

func (e *FleetEngine) RunCommand(ctx context.Context, command string) (*FleetResult, error) {
	return e.RunCommandOnHosts(ctx, command, e.inventory.Hosts)
}

func (e *FleetEngine) RunCommandOnHosts(ctx context.Context, command string, hosts []Host) (*FleetResult, error) {
	result := &FleetResult{
		Timestamp: time.Now(),
		Hosts:     len(hosts),
	}

	if len(hosts) == 0 {
		return result, nil
	}

	results := make([]HostResult, len(hosts))
	sem := make(chan struct{}, e.config.Concurrency)
	var wg sync.WaitGroup

	for i, host := range hosts {
		wg.Add(1)
		sem <- struct{}{}

		go func(idx int, h Host) {
			defer wg.Done()
			defer func() { <-sem }()

			select {
			case <-ctx.Done():
				results[idx] = HostResult{
					Hostname: h.Hostname,
					Success:  false,
					Error:    ctx.Err(),
				}
				return
			default:
			}

			hr := e.executeWithRetry(h, command)
			results[idx] = hr
		}(i, host)
	}

	wg.Wait()

	result.Results = results
	for _, r := range results {
		if r.Success {
			result.Successful++
		} else {
			result.Failed++
		}
	}

	return result, nil
}

func (e *FleetEngine) executeWithRetry(host Host, command string) HostResult {
	var lastErr error
	var lastResult SSHResult

	for attempt := 0; attempt <= e.config.Retry; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		lastResult = e.sshPool.ExecuteWithTimeout(host, command, e.config.Timeout)
		if lastResult.Error == nil {
			return HostResult{
				Hostname: host.Hostname,
				Success:  true,
				Duration: lastResult.Duration,
				Stdout:   lastResult.Stdout,
				Stderr:   lastResult.Stderr,
			}
		}
		lastErr = lastResult.Error
	}

	return HostResult{
		Hostname: host.Hostname,
		Success:  false,
		Error:    fmt.Errorf("after %d retries: %w", e.config.Retry, lastErr),
		Duration: lastResult.Duration,
	}
}

func (e *FleetEngine) RunCheck(ctx context.Context, checkName string, checkFn func(ctx context.Context, host Host) (bool, error)) (*FleetResult, error) {
	result := &FleetResult{
		Timestamp: time.Now(),
		Hosts:     len(e.inventory.Hosts),
	}

	if len(e.inventory.Hosts) == 0 {
		return result, nil
	}

	results := make([]HostResult, len(e.inventory.Hosts))
	sem := make(chan struct{}, e.config.Concurrency)
	var wg sync.WaitGroup

	for i, host := range e.inventory.Hosts {
		wg.Add(1)
		sem <- struct{}{}

		go func(idx int, h Host) {
			defer wg.Done()
			defer func() { <-sem }()

			start := time.Now()
			success, err := checkFn(ctx, h)
			results[idx] = HostResult{
				Hostname: h.Hostname,
				Success:  success && err == nil,
				Error:    err,
				Duration: time.Since(start),
			}
		}(i, host)
	}

	wg.Wait()

	result.Results = results
	for _, r := range results {
		if r.Success {
			result.Successful++
		} else {
			result.Failed++
		}
	}

	return result, nil
}
