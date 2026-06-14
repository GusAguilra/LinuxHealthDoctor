package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Global   GlobalConfig   `mapstructure:"global"`
	Doctor   DoctorConfig   `mapstructure:"doctor"`
	Monitor  MonitorConfig  `mapstructure:"monitor"`
	Dashboard DashboardConfig `mapstructure:"dashboard"`
	Report   ReportConfig   `mapstructure:"report"`
	Baseline BaselineConfig `mapstructure:"baseline"`
	Snapshot SnapshotConfig `mapstructure:"snapshot"`
	Fleet    FleetConfig    `mapstructure:"fleet"`
	Security SecurityConfig `mapstructure:"security"`
	Custom   CustomConfig   `mapstructure:"custom_checks"`
	Plugins  PluginConfig   `mapstructure:"plugins"`
	Knowledge KnowledgeConfig `mapstructure:"knowledge"`
}

type GlobalConfig struct {
	DataDir     string `mapstructure:"data_dir"`
	LogDir      string `mapstructure:"log_dir"`
	LogLevel    string `mapstructure:"log_level"`
	Color       string `mapstructure:"color"`
	Verbose     bool   `mapstructure:"verbose"`
	Timeout     string `mapstructure:"timeout"`
	MaxParallel int    `mapstructure:"max_parallel"`
}

type DoctorConfig struct {
	Categories       []string `mapstructure:"categories"`
	Exclude          []string `mapstructure:"exclude"`
	SeverityThreshold string  `mapstructure:"severity_threshold"`
	Fix              bool     `mapstructure:"fix"`
}

type MonitorConfig struct {
	Enabled   bool          `mapstructure:"enabled"`
	Interval  string        `mapstructure:"interval"`
	Retention string        `mapstructure:"retention"`
	Alerting  AlertConfig   `mapstructure:"alerting"`
	Thresholds ThresholdsConfig `mapstructure:"thresholds"`
}

type AlertConfig struct {
	Enabled        bool `mapstructure:"enabled"`
	DesktopNotify  bool `mapstructure:"desktop_notify"`
	Sound          bool `mapstructure:"sound"`
}

type ThresholdsConfig struct {
	CPU     CPUThresholds     `mapstructure:"cpu"`
	Memory  MemoryThresholds  `mapstructure:"memory"`
	Disk    DiskThresholds    `mapstructure:"disk"`
	Network NetworkThresholds `mapstructure:"network"`
}

type CPUThresholds struct {
	UsagePercent ThresholdValues `mapstructure:"usage_percent"`
	Load1m       ThresholdValues `mapstructure:"load_1m"`
}

type MemoryThresholds struct {
	UsagePercent    ThresholdValues `mapstructure:"usage_percent"`
	SwapUsagePercent ThresholdValues `mapstructure:"swap_usage_percent"`
}

type DiskThresholds struct {
	UsagePercent  ThresholdValues `mapstructure:"usage_percent"`
	InodePercent  ThresholdValues `mapstructure:"inode_percent"`
}

type NetworkThresholds struct {
	LatencyMs        ThresholdValues `mapstructure:"latency_ms"`
	PacketLossPercent ThresholdValues `mapstructure:"packet_loss_percent"`
}

type ThresholdValues struct {
	Warning  float64 `mapstructure:"warning"`
	Critical float64 `mapstructure:"critical"`
}

type DashboardConfig struct {
	RefreshInterval string `mapstructure:"refresh_interval"`
	DefaultView     string `mapstructure:"default_view"`
	Compact         bool   `mapstructure:"compact"`
	ShowAllChecks   bool   `mapstructure:"show_all_checks"`
}

type ReportConfig struct {
	DefaultFormat    string   `mapstructure:"default_format"`
	IncludeEvidence  bool     `mapstructure:"include_evidence"`
	IncludeRemediation bool   `mapstructure:"include_remediation"`
	TemplateDir      string   `mapstructure:"template_dir"`
}

type BaselineConfig struct {
	AutoCapture       bool    `mapstructure:"auto_capture"`
	Schedule          string  `mapstructure:"schedule"`
	Retention         int     `mapstructure:"retention"`
	DeviationThreshold DeviationConfig `mapstructure:"deviation_threshold"`
}

type DeviationConfig struct {
	Warning  float64 `mapstructure:"warning"`
	Critical float64 `mapstructure:"critical"`
}

type SnapshotConfig struct {
	Compress        bool     `mapstructure:"compress"`
	IncludeLogs     bool     `mapstructure:"include_logs"`
	MaxSize         string   `mapstructure:"max_size"`
	ExcludePatterns []string `mapstructure:"exclude_patterns"`
}

type FleetConfig struct {
	InventoryFile string      `mapstructure:"inventory_file"`
	SSH           SSHConfig   `mapstructure:"ssh"`
	Concurrency   int         `mapstructure:"concurrency"`
	Parallelism   string      `mapstructure:"parallelism"`
	Retry         int         `mapstructure:"retry"`
	BastionHost   string      `mapstructure:"bastion_host"`
}

type SSHConfig struct {
	DefaultPort          int    `mapstructure:"default_port"`
	Timeout              string `mapstructure:"timeout"`
	User                 string `mapstructure:"user"`
	KeyFile              string `mapstructure:"key_file"`
	StrictHostKeyChecking bool  `mapstructure:"strict_host_key_checking"`
}

type SecurityConfig struct {
	CheckCVEs    bool   `mapstructure:"check_cves"`
	CVEDBUpdate  int    `mapstructure:"cve_db_update"`
	CustomCVEDB  string `mapstructure:"custom_cve_db"`
	FailOnHigh   bool   `mapstructure:"fail_on_high"`
	SELinux      SELinuxConfig   `mapstructure:"selinux"`
	AppArmor     AppArmorConfig  `mapstructure:"apparmor"`
	Auditd       AuditdConfig    `mapstructure:"auditd"`
}

type SELinuxConfig struct {
	CheckEnforcing bool `mapstructure:"check_enforcing"`
}

type AppArmorConfig struct {
	CheckProfiles bool `mapstructure:"check_profiles"`
}

type AuditdConfig struct {
	CheckRunning bool `mapstructure:"check_running"`
}

type CustomConfig struct {
	Enabled  bool     `mapstructure:"enabled"`
	Dirs     []string `mapstructure:"dirs"`
	Timeout  string   `mapstructure:"timeout"`
	AllowedShebangs []string `mapstructure:"allowed_shebangs"`
}

type PluginConfig struct {
	CheckUpdates  bool   `mapstructure:"check_updates"`
	UpdateChannel string `mapstructure:"update_channel"`
}

type KnowledgeConfig struct {
	CustomDirs       []string `mapstructure:"custom_dirs"`
	SeverityMapping  map[string]string `mapstructure:"severity_mapping"`
}

func Default() *Config {
	return &Config{
		Global: GlobalConfig{
			DataDir:     "~/.local/share/lhd",
			LogDir:      "~/.local/state/lhd/logs",
			LogLevel:    "info",
			Color:       "auto",
			Verbose:     false,
			Timeout:     "30s",
			MaxParallel: 4,
		},
		Doctor: DoctorConfig{
			Categories: []string{"cpu", "memory", "disk", "network", "kernel", "services", "security", "logs", "hardware", "updates", "containers"},
			SeverityThreshold: "warning",
			Fix: false,
		},
		Monitor: MonitorConfig{
			Enabled:   true,
			Interval:  "60s",
			Retention: "720h",
			Alerting: AlertConfig{
				Enabled:       true,
				DesktopNotify: true,
				Sound:         false,
			},
			Thresholds: ThresholdsConfig{
				CPU: CPUThresholds{
					UsagePercent: ThresholdValues{Warning: 80, Critical: 95},
					Load1m:       ThresholdValues{Warning: 4.0, Critical: 8.0},
				},
				Memory: MemoryThresholds{
					UsagePercent:    ThresholdValues{Warning: 80, Critical: 95},
					SwapUsagePercent: ThresholdValues{Warning: 50, Critical: 80},
				},
				Disk: DiskThresholds{
					UsagePercent: ThresholdValues{Warning: 80, Critical: 92},
					InodePercent: ThresholdValues{Warning: 80, Critical: 92},
				},
				Network: NetworkThresholds{
					LatencyMs:        ThresholdValues{Warning: 100, Critical: 500},
					PacketLossPercent: ThresholdValues{Warning: 1.0, Critical: 5.0},
				},
			},
		},
		Dashboard: DashboardConfig{
			RefreshInterval: "5s",
			DefaultView:     "overview",
			Compact:         false,
			ShowAllChecks:   false,
		},
		Report: ReportConfig{
			DefaultFormat:    "markdown",
			IncludeEvidence:  true,
			IncludeRemediation: true,
			TemplateDir:      "~/.config/lhd/templates",
		},
		Baseline: BaselineConfig{
			AutoCapture: false,
			Schedule:    "0 4 * * 0",
			Retention:   10,
			DeviationThreshold: DeviationConfig{
				Warning:  0.1,
				Critical: 0.25,
			},
		},
		Snapshot: SnapshotConfig{
			Compress:    true,
			IncludeLogs: false,
			MaxSize:     "100MB",
			ExcludePatterns: []string{"/tmp/**", "/var/cache/**", "/proc/**", "/sys/**"},
		},
		Fleet: FleetConfig{
			InventoryFile: "~/.config/lhd/hosts.yaml",
			SSH: SSHConfig{
				DefaultPort:          22,
				Timeout:              "30s",
				User:                 "root",
				KeyFile:              "~/.ssh/id_rsa",
				StrictHostKeyChecking: true,
			},
			Concurrency: 10,
			Parallelism: "per-host",
			Retry:       2,
		},
		Security: SecurityConfig{
			CheckCVEs:   true,
			CVEDBUpdate: 86400,
			FailOnHigh:  false,
			SELinux:     SELinuxConfig{CheckEnforcing: true},
			AppArmor:    AppArmorConfig{CheckProfiles: true},
			Auditd:      AuditdConfig{CheckRunning: true},
		},
		Custom: CustomConfig{
			Enabled:  true,
			Dirs:     []string{"~/.config/lhd/checks", "/etc/lhd/checks"},
			Timeout:  "60s",
			AllowedShebangs: []string{"/bin/bash", "/bin/sh", "/usr/bin/python3"},
		},
		Plugins: PluginConfig{
			CheckUpdates:  true,
			UpdateChannel: "stable",
		},
		Knowledge: KnowledgeConfig{
			CustomDirs: []string{"~/.config/lhd/knowledge", "/etc/lhd/knowledge"},
		},
	}
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("lhd")
	v.SetConfigType("yaml")

	if path != "" {
		v.SetConfigFile(path)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot find home dir: %w", err)
		}
		v.AddConfigPath(filepath.Join(home, ".config", "lhd"))
		v.AddConfigPath("/etc/lhd")
		v.AddConfigPath(".")
	}

	v.SetEnvPrefix("LHD")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
	}

	cfg := Default()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	return cfg, nil
}
