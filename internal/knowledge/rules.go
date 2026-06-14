package knowledge

import (
	"fmt"
	"os"

	"github.com/linuxhealthdoctor/lhd/internal/core"
	"gopkg.in/yaml.v3"
)

type yamlRule struct {
	ID           string               `yaml:"id"`
	Name         string               `yaml:"name"`
	Description  string               `yaml:"description"`
	Component    string               `yaml:"component"`
	Severity     string               `yaml:"severity"`
	Conditions   []yamlCondition      `yaml:"conditions"`
	Conclusion   string               `yaml:"conclusion"`
	Certainty    float64              `yaml:"certainty"`
	Remediations []core.Remediation `yaml:"remediations"`
}

type yamlCondition struct {
	FactName string  `yaml:"fact_name"`
	Operator string  `yaml:"operator"`
	Value    float64 `yaml:"value"`
}

type knowledgeBase struct {
	Version string     `yaml:"version"`
	Domain  string     `yaml:"domain"`
	Rules   []yamlRule `yaml:"rules"`
}

func LoadRulesFromYAML(path string) ([]Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading knowledge base %s: %w", path, err)
	}

	var kb knowledgeBase
	if err := yaml.Unmarshal(data, &kb); err != nil {
		return nil, fmt.Errorf("parsing knowledge base %s: %w", path, err)
	}

	rules := make([]Rule, 0, len(kb.Rules))
	for _, yr := range kb.Rules {
		sev, err := parseSeverity(yr.Severity)
		if err != nil {
			return nil, fmt.Errorf("rule %s: %w", yr.ID, err)
		}

		comp := core.Component(yr.Component)
		if comp == "" {
			comp = core.Component(kb.Domain)
		}

		conditions := make([]Condition, len(yr.Conditions))
		for i, c := range yr.Conditions {
			conditions[i] = Condition{
				FactName: c.FactName,
				Operator: c.Operator,
				Value:    c.Value,
			}
		}

		if yr.Remediations == nil {
			yr.Remediations = []core.Remediation{}
		}

		rules = append(rules, Rule{
			ID:           yr.ID,
			Name:         yr.Name,
			Description:  yr.Description,
			Component:    comp,
			Severity:     sev,
			Conditions:   conditions,
			Conclusion:   yr.Conclusion,
			Certainty:    yr.Certainty,
			Remediations: yr.Remediations,
		})
	}

	return rules, nil
}

func parseSeverity(s string) (core.Severity, error) {
	switch s {
	case "none":
		return core.SeverityNone, nil
	case "info":
		return core.SeverityInfo, nil
	case "warning":
		return core.SeverityWarning, nil
	case "critical":
		return core.SeverityCritical, nil
	case "fatal":
		return core.SeverityFatal, nil
	default:
		return core.SeverityNone, fmt.Errorf("unknown severity: %s", s)
	}
}

func BuiltinCPURules() []Rule {
	return []Rule{
		{
			ID:          "cpu.high_usage",
			Name:        "High CPU Usage",
			Description: "CPU usage exceeds warning threshold",
			Component:   core.ComponentCPU,
			Severity:    core.SeverityWarning,
			Conditions: []Condition{
				{FactName: "cpu.usage_percent", Operator: "gt", Value: 80},
			},
			Conclusion: "CPU usage exceeds normal threshold (80%)",
			Certainty:  0.85,
			Remediations: []core.Remediation{
				{Step: 1, Action: "Identify top processes consuming CPU", Command: "ps aux --sort=-%cpu | head -10", Priority: 1, Risk: "low"},
				{Step: 2, Action: "Check for runaway processes", Command: "top -b -n1 | head -20", Priority: 2, Risk: "low"},
			},
		},
		{
			ID:          "cpu.critical_usage",
			Name:        "Critical CPU Usage",
			Description: "CPU usage critically high",
			Component:   core.ComponentCPU,
			Severity:    core.SeverityCritical,
			Conditions: []Condition{
				{FactName: "cpu.usage_percent", Operator: "gt", Value: 95},
			},
			Conclusion: "CPU usage is critically high (95%+)",
			Certainty:  0.95,
			Remediations: []core.Remediation{
				{Step: 1, Action: "Immediately identify and investigate top CPU processes", Command: "ps aux --sort=-%cpu | head -10", Priority: 1, Risk: "low"},
				{Step: 2, Action: "Check for fork bombs or runaways", Command: "ps -eo pid,ppid,cmd,%cpu --sort=-%cpu | head -20", Priority: 2, Risk: "low"},
				{Step: 3, Action: "Consider restarting misbehaving services", Command: "systemctl list-units --state-running | head -20", Priority: 3, Risk: "medium"},
			},
		},
		{
			ID:          "cpu.high_load",
			Name:        "High System Load Average",
			Description: "System load average is elevated",
			Component:   core.ComponentCPU,
			Severity:    core.SeverityWarning,
			Conditions: []Condition{
				{FactName: "cpu.load_1m", Operator: "gt", Value: 4.0},
			},
			Conclusion: "System load average is elevated, system may be under heavy demand",
			Certainty:  0.80,
			Remediations: []core.Remediation{
				{Step: 1, Action: "Check load averages and running processes", Command: "uptime && ps -eo pid,comm,%cpu,%mem --sort=-%cpu | head -15", Priority: 1, Risk: "low"},
				{Step: 2, Action: "Check I/O wait which may indicate disk bottleneck", Command: "iostat -xz 1 3", Priority: 2, Risk: "low"},
			},
		},
		{
			ID:          "cpu.load_spike",
			Name:        "Recent CPU Load Spike",
			Description: "1-minute load significantly higher than 5-minute load",
			Component:   core.ComponentCPU,
			Severity:    core.SeverityWarning,
			Conditions: []Condition{
				{FactName: "cpu.load_1m", Operator: "gt", Value: 2.0},
				{FactName: "cpu.load_5m", Operator: "gt", Value: 0.5},
			},
			Conclusion: "Recent CPU load spike detected - 1m load significantly exceeds 5m average",
			Certainty:  0.75,
			Remediations: []core.Remediation{
				{Step: 1, Action: "Identify process causing recent spike", Command: "ps aux --sort=-%cpu | head -10", Priority: 1, Risk: "low"},
				{Step: 2, Action: "Check system logs for recent activity", Command: "journalctl --since '5 minutes ago' --no-pager | tail -50", Priority: 2, Risk: "low"},
			},
		},
	}
}

func BuiltinMemoryRules() []Rule {
	return []Rule{
		{
			ID:          "memory.high_usage",
			Name:        "High Memory Usage",
			Description: "Memory usage exceeds warning threshold",
			Component:   core.ComponentMemory,
			Severity:    core.SeverityWarning,
			Conditions: []Condition{
				{FactName: "memory.usage_percent", Operator: "gt", Value: 80},
			},
			Conclusion: "Memory usage exceeds normal threshold (80%)",
			Certainty:  0.85,
			Remediations: []core.Remediation{
				{Step: 1, Action: "Identify top memory-consuming processes", Command: "ps aux --sort=-%mem | head -10", Priority: 1, Risk: "low"},
				{Step: 2, Action: "Check for memory leaks in applications", Command: "smem -t -p | head -15", Priority: 2, Risk: "low"},
			},
		},
		{
			ID:          "memory.critical_usage",
			Name:        "Critical Memory Usage",
			Description: "Memory usage critically high",
			Component:   core.ComponentMemory,
			Severity:    core.SeverityCritical,
			Conditions: []Condition{
				{FactName: "memory.usage_percent", Operator: "gt", Value: 95},
			},
			Conclusion: "Memory usage is critically high (95%+) - OOM risk imminent",
			Certainty:  0.95,
			Remediations: []core.Remediation{
				{Step: 1, Action: "Identify largest memory consumers immediately", Command: "ps aux --sort=-%mem | head -10", Priority: 1, Risk: "low"},
				{Step: 2, Action: "Check swap usage and OOM status", Command: "swapon --show && dmesg | grep -i 'oom\\|out of memory' | tail -10", Priority: 2, Risk: "low"},
				{Step: 3, Action: "Consider clearing caches or restarting memory-heavy services", Command: "sync && echo 3 > /proc/sys/vm/drop_caches", Priority: 3, Risk: "medium"},
			},
		},
		{
			ID:          "memory.swap_high",
			Name:        "High Swap Usage",
			Description: "Swap usage indicates memory pressure",
			Component:   core.ComponentMemory,
			Severity:    core.SeverityWarning,
			Conditions: []Condition{
				{FactName: "memory.swap_usage_percent", Operator: "gt", Value: 50},
			},
			Conclusion: "High swap usage indicates significant memory pressure",
			Certainty:  0.80,
			Remediations: []core.Remediation{
				{Step: 1, Action: "Check which processes are using swap", Command: "for file in /proc/*/status; do awk '/VmSwap|Name/{printf $2 \" \" $3}END{ print \"\"}' $file 2>/dev/null; done | sort -k2 -n -r | head -10", Priority: 1, Risk: "low"},
				{Step: 2, Action: "Review memory allocation and consider adding RAM", Command: "free -h && vmstat -s", Priority: 2, Risk: "low"},
			},
		},
	}
}

func BuiltinDiskRules() []Rule {
	return []Rule{
		{
			ID:          "disk.space_low",
			Name:        "Disk Space Low",
			Description: "Disk usage exceeds warning threshold",
			Component:   core.ComponentDisk,
			Severity:    core.SeverityWarning,
			Conditions: []Condition{
				{FactName: "disk.usage_percent", Operator: "gt", Value: 80},
			},
			Conclusion: "Disk usage exceeds normal threshold (80%)",
			Certainty:  0.85,
			Remediations: []core.Remediation{
				{Step: 1, Action: "Identify largest files and directories", Command: "du -sh /* 2>/dev/null | sort -hr | head -10", Priority: 1, Risk: "low"},
				{Step: 2, Action: "Check for old log files", Command: "find /var/log -type f -name '*.log' -size +100M -exec ls -lh {} \\; 2>/dev/null", Priority: 2, Risk: "low"},
				{Step: 3, Action: "Clean package cache", Command: "apt-get clean 2>/dev/null || dnf clean all 2>/dev/null || pacman -Sc --noconfirm 2>/dev/null", Priority: 3, Risk: "low"},
			},
		},
		{
			ID:          "disk.space_critical",
			Name:        "Disk Space Critical",
			Description: "Disk usage critically high",
			Component:   core.ComponentDisk,
			Severity:    core.SeverityCritical,
			Conditions: []Condition{
				{FactName: "disk.usage_percent", Operator: "gt", Value: 92},
			},
			Conclusion: "Disk usage is critically high (92%+) - system may become unstable",
			Certainty:  0.95,
			Remediations: []core.Remediation{
				{Step: 1, Action: "Immediately free disk space", Command: "du -sh /* 2>/dev/null | sort -hr | head -10", Priority: 1, Risk: "low"},
				{Step: 2, Action: "Clean journal logs", Command: "journalctl --vacuum-size=200M", Priority: 2, Risk: "low"},
				{Step: 3, Action: "Remove old unused packages", Command: "apt autoremove --purge -y 2>/dev/null || dnf autoremove -y 2>/dev/null || pacman -Rns $(pacman -Qdtq) --noconfirm 2>/dev/null", Priority: 3, Risk: "medium"},
			},
		},
		{
			ID:          "disk.inode_exhaustion",
			Name:        "Inode Exhaustion Risk",
			Description: "Inode usage is critically high",
			Component:   core.ComponentDisk,
			Severity:    core.SeverityCritical,
			Conditions: []Condition{
				{FactName: "disk.inode_percent", Operator: "gt", Value: 85},
			},
			Conclusion: "Inode usage is critically high - risk of filesystem exhaustion",
			Certainty:  0.80,
			Remediations: []core.Remediation{
				{Step: 1, Action: "Find directories with excessive small files", Command: "for d in /*; do echo \"$d: $(find $d -xdev -type f 2>/dev/null | wc -l)\"; done 2>/dev/null | sort -t: -k2 -rn | head -10", Priority: 1, Risk: "low"},
				{Step: 2, Action: "Check for mail spool or temp file accumulation", Command: "ls -la /var/spool/mail/ 2>/dev/null && ls -la /tmp/ 2>/dev/null | wc -l", Priority: 2, Risk: "low"},
			},
		},
	}
}

func BuiltinNetworkRules() []Rule {
	return []Rule{
		{
			ID:          "network.high_latency",
			Name:        "High Network Latency",
			Description: "Network latency exceeds warning threshold",
			Component:   core.ComponentNetwork,
			Severity:    core.SeverityWarning,
			Conditions: []Condition{
				{FactName: "network.latency_ms", Operator: "gt", Value: 100},
			},
			Conclusion: "Network latency exceeds normal threshold (100ms)",
			Certainty:  0.75,
			Remediations: []core.Remediation{
				{Step: 1, Action: "Test connectivity and latency to common endpoints", Command: "ping -c 5 8.8.8.8", Priority: 1, Risk: "low"},
				{Step: 2, Action: "Check for network interface errors", Command: "netstat -i || ip -s link", Priority: 2, Risk: "low"},
			},
		},
		{
			ID:          "network.packet_loss",
			Name:        "Network Packet Loss Detected",
			Description: "Packet loss exceeds threshold",
			Component:   core.ComponentNetwork,
			Severity:    core.SeverityCritical,
			Conditions: []Condition{
				{FactName: "network.packet_loss_percent", Operator: "gt", Value: 1.0},
			},
			Conclusion: "Network packet loss detected (>1%) which may indicate faulty hardware or congestion",
			Certainty:  0.85,
			Remediations: []core.Remediation{
				{Step: 1, Action: "Run extended ping test to determine loss pattern", Command: "ping -c 100 -i 0.2 8.8.8.8 | grep loss", Priority: 1, Risk: "low"},
				{Step: 2, Action: "Check interface errors and duplex mismatches", Command: "ethtool $(ip route get 8.8.8.8 | awk '{print $5}') 2>/dev/null", Priority: 2, Risk: "low"},
				{Step: 3, Action: "Review system logs for network driver errors", Command: "dmesg | grep -i 'eth\\|net\\|link\\|duplex' | tail -20", Priority: 3, Risk: "low"},
			},
		},
	}
}
