package security

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/linuxhealthdoctor/lhd/internal/core"
	"github.com/linuxhealthdoctor/lhd/internal/plugin"
)

func init() {
	plugin.Register(&SecurityPortsCheck{})
	plugin.Register(&SecuritySUIDCheck{})
	plugin.Register(&SecuritySSHCheck{})
}

type SecurityPortsCheck struct{}

func (c *SecurityPortsCheck) ID() string               { return "security.ports" }
func (c *SecurityPortsCheck) Name() string             { return "Open Ports" }
func (c *SecurityPortsCheck) Description() string      { return "Checks for listening ports and unexpected services" }
func (c *SecurityPortsCheck) Category() core.Component { return core.ComponentSecurity }
func (c *SecurityPortsCheck) Dependencies() []string   { return nil }
func (c *SecurityPortsCheck) Tags() []string           { return []string{"security", "network", "ports"} }

func (c *SecurityPortsCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	cmd := exec.CommandContext(ctx, "ss", "-tlnp")
	output, err := cmd.Output()
	if err != nil {
		result.Status = core.StatusError
		result.Error = fmt.Errorf("failed to list listening ports: %w", err)
		return result, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	ports := []map[string]string{}
	highPorts := 0

	for _, l := range lines[1:] {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		fields := strings.Fields(l)
		if len(fields) < 4 {
			continue
		}
		addr := fields[3]
		parts := strings.Split(addr, ":")
		portStr := parts[len(parts)-1]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}
		entry := map[string]string{
			"address": addr,
			"port":    portStr,
		}
		if len(fields) > 4 {
			entry["process"] = fields[len(fields)-1]
		}
		ports = append(ports, entry)
		if port > 1024 {
			highPorts++
		}
	}

	result.Metrics = map[string]float64{
		"listening_ports": float64(len(ports)),
		"high_ports":      float64(highPorts),
	}
	result.Details = map[string]interface{}{
		"listening_ports": len(ports),
		"ports":           ports,
	}

	if len(ports) > 50 {
		result.Status = core.StatusFail
		result.Severity = core.SeverityWarning
		result.Message = fmt.Sprintf("High number of listening ports: %d", len(ports))
	} else {
		result.Message = fmt.Sprintf("%d listening ports detected", len(ports))
	}

	return result, nil
}

type SecuritySUIDCheck struct{}

func (c *SecuritySUIDCheck) ID() string               { return "security.suid" }
func (c *SecuritySUIDCheck) Name() string             { return "SUID Files" }
func (c *SecuritySUIDCheck) Description() string      { return "Checks for SUID binaries that could be security risks" }
func (c *SecuritySUIDCheck) Category() core.Component { return core.ComponentSecurity }
func (c *SecuritySUIDCheck) Dependencies() []string   { return nil }
func (c *SecuritySUIDCheck) Tags() []string           { return []string{"security", "suid", "binaries"} }

func (c *SecuritySUIDCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	cmd := exec.CommandContext(ctx, "find", "/usr", "/bin", "/sbin",
		"-type", "f", "-perm", "-4000",
		"!", "-name", "*.sh",
		"!", "-path", "*/snap/*",
	)
	output, err := cmd.Output()
	if err != nil {
		result.Status = core.StatusError
		result.Error = fmt.Errorf("failed to search for SUID files: %w", err)
		return result, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	suidFiles := []string{}
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			suidFiles = append(suidFiles, l)
		}
	}

	result.Metrics = map[string]float64{
		"suid_files": float64(len(suidFiles)),
	}
	result.Details = map[string]interface{}{
		"suid_count": len(suidFiles),
		"suid_files": suidFiles,
	}

	if len(suidFiles) > 100 {
		result.Status = core.StatusFail
		result.Severity = core.SeverityWarning
		result.Message = fmt.Sprintf("High number of SUID binaries: %d", len(suidFiles))
	} else {
		result.Message = fmt.Sprintf("%d SUID binaries found", len(suidFiles))
	}

	return result, nil
}

type SecuritySSHCheck struct{}

func (c *SecuritySSHCheck) ID() string               { return "security.ssh" }
func (c *SecuritySSHCheck) Name() string             { return "SSH Configuration" }
func (c *SecuritySSHCheck) Description() string      { return "Checks SSH daemon configuration for security settings" }
func (c *SecuritySSHCheck) Category() core.Component { return core.ComponentSecurity }
func (c *SecuritySSHCheck) Dependencies() []string   { return nil }
func (c *SecuritySSHCheck) Tags() []string           { return []string{"security", "ssh", "configuration"} }

func (c *SecuritySSHCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*core.CheckResult, error) {
	result := &core.CheckResult{ID: c.ID(), Timestamp: time.Now(), Status: core.StatusPass, Category: c.Category()}

	checks := map[string]string{
		"PermitRootLogin": "no",
		"PasswordAuthentication": "no",
		"Protocol": "2",
		"X11Forwarding": "no",
	}

	cmd := exec.CommandContext(ctx, "sshd", "-T")
	output, err := cmd.Output()
	if err != nil {
		result.Status = core.StatusError
		result.Error = fmt.Errorf("failed to read SSH configuration: %w", err)
		return result, nil
	}

	configLines := strings.Split(string(output), "\n")
	configMap := make(map[string]string)
	for _, l := range configLines {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		parts := strings.Fields(l)
		if len(parts) >= 2 {
			configMap[parts[0]] = parts[1]
		}
	}

	issues := []string{}
	for key, expected := range checks {
		if val, ok := configMap[key]; ok {
			if !strings.EqualFold(val, expected) {
				issues = append(issues, fmt.Sprintf("%s=%s (expected %s)", key, val, expected))
			}
		}
	}

	result.Metrics = map[string]float64{
		"checks_performed": float64(len(checks)),
		"issues_found":     float64(len(issues)),
	}
	result.Details = map[string]interface{}{
		"checks_performed": len(checks),
		"issues":           issues,
	}

	if len(issues) > 0 {
		result.Status = core.StatusFail
		result.Severity = core.SeverityWarning
		result.Message = fmt.Sprintf("SSH configuration has %d security issues", len(issues))
	} else {
		result.Message = "SSH configuration security checks passed"
	}

	return result, nil
}
