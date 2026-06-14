package network

import (
	"context"
	"fmt"
	stdnet "net"
	"os/exec"
	"time"

	"github.com/GusAguilra/LinuxHealthDoctor/internal/core"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/plugin"
	"github.com/shirou/gopsutil/v3/net"
)

func init() {
	plugin.Register(&NetworkConnectivityCheck{})
	plugin.Register(&NetworkInterfacesCheck{})
	plugin.Register(&NetworkDNSCheck{})
}

type NetworkConnectivityCheck struct{}

func (c *NetworkConnectivityCheck) ID() string               { return "network.connectivity" }
func (c *NetworkConnectivityCheck) Name() string             { return "Network Connectivity" }
func (c *NetworkConnectivityCheck) Description() string      { return "Checks basic network connectivity" }
func (c *NetworkConnectivityCheck) Category() core.Component { return core.ComponentNetwork }
func (c *NetworkConnectivityCheck) Dependencies() []string   { return nil }
func (c *NetworkConnectivityCheck) Tags() []string           { return []string{"network", "connectivity", "reliability"} }

func (c *NetworkConnectivityCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	targets := []string{"8.8.8.8", "1.1.1.1"}
	reachable := 0

	for _, target := range targets {
		cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", "2", target)
		if err := cmd.Run(); err == nil {
			reachable++
		}
	}

	result.Metrics = map[string]float64{
		"targets_tested": float64(len(targets)),
		"targets_reachable": float64(reachable),
	}
	result.Details = map[string]interface{}{
		"targets_tested":   len(targets),
		"targets_reachable": reachable,
	}

	switch {
	case reachable == 0:
		result.Status = core.StatusFail
		result.Severity = core.SeverityCritical
		result.Message = "No network targets reachable"
	case reachable < len(targets):
		result.Status = core.StatusFail
		result.Severity = core.SeverityWarning
		result.Message = fmt.Sprintf("Only %d/%d network targets reachable", reachable, len(targets))
	default:
		result.Message = fmt.Sprintf("All %d network targets reachable", reachable)
	}

	return result, nil
}

type NetworkInterfacesCheck struct{}

func (c *NetworkInterfacesCheck) ID() string               { return "network.interfaces" }
func (c *NetworkInterfacesCheck) Name() string             { return "Network Interfaces" }
func (c *NetworkInterfacesCheck) Description() string      { return "Checks network interface errors and stats" }
func (c *NetworkInterfacesCheck) Category() core.Component { return core.ComponentNetwork }
func (c *NetworkInterfacesCheck) Dependencies() []string   { return nil }
func (c *NetworkInterfacesCheck) Tags() []string           { return []string{"network", "interfaces", "errors"} }

func (c *NetworkInterfacesCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	counters, err := net.IOCountersWithContext(ctx, true)
	if err != nil {
		result.Status = core.StatusError
		result.Error = fmt.Errorf("failed to read network interface counters: %w", err)
		return result, nil
	}

	result.Metrics = make(map[string]float64)
	result.Details = map[string]interface{}{
		"interfaces": []map[string]interface{}{},
	}
	ifaceDetails := []map[string]interface{}{}
	totalErrors := uint64(0)

	for _, ctr := range counters {
		errCount := ctr.Errin + ctr.Errout
		totalErrors += errCount
		ifaceDetails = append(ifaceDetails, map[string]interface{}{
			"name":       ctr.Name,
			"bytes_sent": ctr.BytesSent,
			"bytes_recv": ctr.BytesRecv,
			"errors_in":  ctr.Errin,
			"errors_out": ctr.Errout,
			"dropped_in": ctr.Dropin,
			"dropped_out": ctr.Dropout,
		})
		result.Metrics[ctr.Name+"_errors"] = float64(errCount)
		result.Metrics[ctr.Name+"_bytes_sent"] = float64(ctr.BytesSent)
		result.Metrics[ctr.Name+"_bytes_recv"] = float64(ctr.BytesRecv)
	}

	result.Details["interfaces"] = ifaceDetails

	if totalErrors > 1000 {
		result.Status = core.StatusFail
		result.Severity = core.SeverityWarning
		result.Message = fmt.Sprintf("High network error count detected: %d total errors", totalErrors)
	} else {
		result.Message = fmt.Sprintf("Network interfaces look healthy (%d interfaces, %d errors)", len(counters), totalErrors)
	}

	return result, nil
}

type NetworkDNSCheck struct{}

func (c *NetworkDNSCheck) ID() string               { return "network.dns" }
func (c *NetworkDNSCheck) Name() string             { return "DNS Resolution" }
func (c *NetworkDNSCheck) Description() string      { return "Checks DNS resolution is working" }
func (c *NetworkDNSCheck) Category() core.Component { return core.ComponentNetwork }
func (c *NetworkDNSCheck) Dependencies() []string   { return nil }
func (c *NetworkDNSCheck) Tags() []string           { return []string{"network", "dns", "connectivity"} }

func (c *NetworkDNSCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	hosts := []string{"google.com", "cloudflare.com"}
	resolved := 0

	for _, host := range hosts {
		ips, err := stdnet.DefaultResolver.LookupHost(ctx, host)
		if err == nil && len(ips) > 0 {
			resolved++
		}
	}

	result.Metrics = map[string]float64{
		"hosts_tested":     float64(len(hosts)),
		"hosts_resolved":   float64(resolved),
	}
	result.Details = map[string]interface{}{
		"hosts_tested":   hosts,
		"hosts_resolved": resolved,
	}

	switch {
	case resolved == 0:
		result.Status = core.StatusFail
		result.Severity = core.SeverityCritical
		result.Message = "DNS resolution failed for all test hosts"
	case resolved < len(hosts):
		result.Status = core.StatusFail
		result.Severity = core.SeverityWarning
		result.Message = fmt.Sprintf("DNS resolution partially failed: %d/%d resolved", resolved, len(hosts))
	default:
		result.Message = "DNS resolution working correctly"
	}

	return result, nil
}
