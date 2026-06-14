# Linux Health Doctor (lhd) — Architecture Document

**Version:** 1.0.0  
**License:** MIT  
**Language:** Go  
**CLI:** `lhd`  
**Tagline:** Your Linux system's primary care physician.

> **Note:** This document describes the *intended* architecture. Several modules described here (hardware/containers/updates checks, custom check scripts, PDF export, full fleet management) are planned but not yet implemented. The actual codebase contains working implementations for the core health checks, TUI dashboard, knowledge engine, and SQLite storage.

---

## Table of Contents

1. [Complete Architecture](#1-complete-architecture)
2. [Repository Structure](#2-repository-structure)
3. [Package Structure](#3-package-structure)
4. [Internal Modules](#4-internal-modules)
5. [Interfaces](#5-interfaces)
6. [Configuration Design](#6-configuration-design)
7. [CLI Command Tree](#7-cli-command-tree)
8. [Plugin Architecture](#8-plugin-architecture)
9. [Distribution Abstraction Layer](#9-distribution-abstraction-layer)
10. [Root Cause Analysis Architecture](#10-root-cause-analysis-architecture)
11. [Reporting Architecture](#11-reporting-architecture)
12. [Dashboard Architecture](#12-dashboard-architecture)
13. [Security Architecture](#13-security-architecture)
14. [Fleet Architecture](#14-fleet-architecture)
15. [Snapshot Architecture](#15-snapshot-architecture)
16. [Baseline Architecture](#16-baseline-architecture)
17. [Knowledge Engine Architecture](#17-knowledge-engine-architecture)
18. [Dependency Selection](#18-dependency-selection)
19. [Development Roadmap](#19-development-roadmap)
20. [Release Roadmap](#20-release-roadmap)

---

## 1. Complete Architecture

### 1.1 Philosophy

Linux Health Doctor is a **local-first**, **offline-first**, **zero-telemetry** health diagnostics and root-cause analysis platform for Linux systems. It treats your Linux system like a patient — running systematic checks, establishing baselines of "healthy" behavior, detecting anomalies, performing root-cause analysis, and providing remediation guidance.

Every architectural decision prioritizes:
- **Privacy** — zero data leaves the machine
- **Autonomy** — no cloud, no external AI, no telemetry
- **Composability** — every component is a distinct module with a well-defined interface
- **Extensibility** — plugins, custom checks, custom knowledge bases
- **Portability** — runs on any Linux distribution without modification

### 1.2 High-Level System Context

```
┌─────────────────────────────────────────────────────────────┐
│                        User (Terminal)                       │
└──────────────────────────┬──────────────────────────────────┘
                           │ lhd doctor | monitor | dashboard ...
                           ▼
┌──────────────────────────────────────────────────────────────┐
│                        CLI Layer (Cobra)                     │
│  ┌─────────┬──────────┬──────────┬─────────┬──────────────┐ │
│  │ doctor  │ diagnose  │ monitor  │dashboard│   report     │ │
│  ├─────────┼──────────┼──────────┼─────────┼──────────────┤ │
│  │baseline │ snapshot │  fleet   │ config  │   version    │ │
│  └─────────┴──────────┴──────────┴─────────┴──────────────┘ │
└──────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────┐
│                    Orchestration Layer                        │
│  ┌──────────────────────────────────────────────────────────┐│
│  │  Scheduler | Pipeline | Event Bus | Output Formatter    ││
│  └──────────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────────┘
                           │
               ┌───────────┼───────────┐
               ▼           ▼           ▼
┌─────────────────┐ ┌───────────┐ ┌──────────────────────┐
│  Plugin System   │ │Knowledge  │ │   Distribution       │
│  (Health Checks) │ │ Engine    │ │   Abstraction Layer  │
└─────────────────┘ └───────────┘ └──────────────────────┘
       │                  │                  │
       ▼                  ▼                  ▼
┌──────────────────────────────────────────────────────────────┐
│                    Storage Layer                              │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌───────────────┐  │
│  │ SQLite   │ │  BoltDB   │ │ YAML     │ │   JSON/CSV    │  │
│  │(metadata)│ │(timeseries│ │(config)  │ │   (reports)   │  │
│  │          │ │ data)    │ │          │ │               │  │
│  └──────────┘ └──────────┘ └──────────┘ └───────────────┘  │
└──────────────────────────────────────────────────────────────┘
```

### 1.3 Architectural Decisions Record

| ID | Decision | Rationale |
|---|---|---|
| ADR-001 | Go as implementation language | Single binary, cross-compilation, excellent stdlib for systems programming, static typing, great concurrency for parallel checks |
| ADR-002 | Compiled-in plugins (no WASM/Python) | Simplicity, security, single-binary distribution, no runtime dependencies |
| ADR-003 | SQLite for relational/persistent data | Local-first, zero-config, battle-tested, embedded, ACID compliant |
| ADR-004 | BoltDB for time-series monitoring data | Purpose-built for time-series, fast reads/writes, embedded, no server |
| ADR-005 | YAML for human-editable configuration | Readable, widely understood, supports comments |
| ADR-006 | Cobra for CLI | De facto standard Go CLI framework, subcommand support, autocomplete |
| ADR-007 | Bubble Tea for TUI dashboard | Charmbracelet ecosystem, Elm-architecture, excellent Go TUI framework |
| ADR-008 | No WASM/runtime plugin loading | Security surface reduction, distribution simplicity, no versioning nightmares |
| ADR-009 | Rule-based knowledge engine (no ML/AI) | Deterministic, auditable, explainable, offline-capable, no model dependencies |
| ADR-010 | Event bus for internal communication | Decouples check execution from data collection, enables reactive monitoring |
| ADR-011 | Distribution abstraction as Go interfaces | Clean separation, easy to add new distros, testable with mocks |
| ADR-012 | All checks are plugins with standard interface | Uniform execution model, composable pipelines, shared infrastructure |

---

## 2. Repository Structure

```
lhd/
├── .github/
│   └── workflows/
│       ├── ci.yml                   # CI pipeline: lint, test, build
│       └── release.yml              # GoReleaser automated releases
│
├── cmd/
│   └── lhd/
│       └── main.go                # Entry point
│
├── internal/
│   ├── cli/                       # CLI command implementations
│   │   ├── root.go                # Root command
│   │   ├── doctor.go              # lhd doctor
│   │   ├── diagnose.go            # lhd diagnose
│   │   ├── monitor.go             # lhd monitor
│   │   ├── dashboard.go           # lhd dashboard
│   │   ├── report.go              # lhd report
│   │   ├── baseline.go            # lhd baseline
│   │   ├── snapshot.go            # lhd snapshot
│   │   ├── fleet.go               # lhd fleet
│   │   ├── config.go              # lhd config
│   │   └── completion.go          # Shell completion
│   │
│   ├── core/                      # Core domain types
│   │   ├── types.go               # Core data types
│   │   ├── errors.go              # Domain errors
│   │   ├── constants.go           # Constants
│   │   ├── event.go               # Event bus types
│   │   └── json.go                # JSON helpers
│   │
│   ├── plugin/                    # Plugin system
│   │   ├── interface.go           # Checker interface
│   │   ├── registry.go            # Plugin registry
│   │   └── context.go             # Check context
│   │
│   ├── checks/                    # Built-in health check plugins
│   │   ├── cpu/
│   │   │   └── cpu.go              # CPU checks
│   │   ├── memory/
│   │   │   └── memory.go           # Memory checks
│   │   ├── disk/
│   │   │   └── disk.go             # Disk checks
│   │   ├── network/
│   │   │   └── network.go          # Network checks
│   │   ├── kernel/
│   │   │   └── kernel.go           # Kernel checks
│   │   ├── services/
│   │   │   └── services.go         # Service checks
│   │   ├── security/
│   │   │   └── security.go         # Security checks
│   │   └── logs/
│   │       └── logs.go             # Log analysis checks
│   │
│   ├── distro/                    # Distribution abstraction layer
│   │   ├── interface.go           # Distro interface
│   │   ├── common.go              # Common implementations
│   │   ├── detection.go           # Auto-detection logic
│   │   ├── debian.go              # Debian/Ubuntu
│   │   ├── rhel.go                # RHEL/CentOS/Fedora
│   │   ├── arch.go                # Arch Linux
│   │   ├── suse.go                # openSUSE/SLES
│   │   ├── gentoo.go              # Gentoo
│   │   ├── alpine.go              # Alpine
│   │   └── nixos.go               # NixOS
│   │
│   ├── knowledge/                 # Knowledge engine (in progress)
│   │   ├── engine.go              # Inference engine
│   │   ├── rules.go               # Rule definitions
│   │   └── builtin/               # Built-in knowledge bases
│   │       └── cpu.yaml
│   │       ├── security.yaml
│   │       └── performance.yaml
│   │
│   ├── monitor/                   # Monitoring engine (prototype)
│   │   └── engine.go              # Monitor orchestration
│   │
│   ├── dashboard/                 # TUI Dashboard
│   │   ├── model.go               # Bubble Tea model
│   │   ├── view.go                # View rendering
│   │   ├── update.go              # Update handlers
│   │   ├── components/            # TUI components
│   │   │   ├── gauge.go           # Health gauge
│   │   │   ├── table.go           # Data table
│   │   │   ├── chart.go           # Sparkline chart
│   │   │   └── statusbar.go       # Status bar
│   │   └── colors.go              # Color scheme
│   │
│   ├── report/                    # Reporting engine
│   │   ├── engine.go              # Report generation
│   │   ├── formatter.go           # Output formatting (ANSI, JSON)
│   │   └── formats/               # (planned: additional formats)
│   │   └── templates/             # (planned: report templates)
│   │
│   ├── baseline/                  # Baseline engine (in progress)
│   │   ├── engine.go              # Baseline capture/compare
│   │   ├── profile.go             # System profile definition
│   │   ├── comparator.go          # Comparison logic
│   │   ├── deviation.go           # Deviation analysis
│   │   └── migration.go           # Baseline version migration
│   │
│   ├── snapshot/                  # Snapshot engine
│   │   ├── engine.go              # Snapshot capture
│   │   ├── manifest.go            # Snapshot manifest
│   │   ├── differ.go              # Snapshot diffing
│   │   ├── restore.go             # Snapshot restore (selective)
│   │   └── compress.go            # Compression strategy
│   │
│   ├── fleet/                     # Fleet management (SSH-based, in progress)
│   │   ├── engine.go              # Fleet orchestration
│   │   ├── ssh.go                 # SSH connection management
│   │   ├── inventory.go           # Fleet inventory
│   │   └── aggregation.go         # Result aggregation
│   │
│   ├── storage/                   # Storage layer
│   │   ├── store.go               # Storage interface
│   │   ├── sqlite.go              # SQLite implementation (primary)
│   │   ├── boltdb.go              # BoltDB implementation (stub)
│   │   └── repository/            # Data access repositories
│   │       └── check_result.go
│   │
│   ├── config/                    # Configuration management
│   │   └── config.go              # Config struct, defaults & Viper loading
│   │
│   ├── output/                    # Output formatting
│   │   ├── printer.go             # Print interface
│   │   ├── table.go               # Table formatter
│   │   ├── color.go               # Color utilities
│   │   └── progress.go            # Progress indicators
│   │
│   └── scheduler/                 # Job scheduling
│       ├── scheduler.go           # Cron-like scheduler
│       ├── job.go                 # Job definition
│       └── cron.go                # Cron expression parser
│
├── pkg/                           # Public reusable packages
│   ├── version/
│   │   └── version.go             # Version info (set at build time)
│   └── lhd/                       # Public library interface
│       └── lhd.go                 # Library entry point (embed LHD)
│
├── testdata/                      # Test fixtures
│   ├── configs/
│   │   ├── valid.yaml
│   │   └── invalid.yaml
│   ├── snapshots/
│   └── knowledge/
│
├── docs/                          # Documentation
│   ├── getting-started.md
│   ├── commands.md
│   ├── configuration.md
│   ├── plugins.md
│   ├── knowledge-base.md
│   ├── fleet.md
│   └── architecture.md            # Symlink or pointer to root
│
├── examples/                      # Example configs and scripts
│   ├── lhd.yaml
│   └── custom-check.sh
│
├── scripts/                       # Build & dev scripts
│   ├── build.sh
│   ├── test.sh
│   ├── lint.sh
│   └── release.sh
│
├── Makefile
├── go.mod
├── go.sum
├── goreleaser.yaml                # GoReleaser configuration
├── LICENSE                        # MIT License
├── README.md
├── CONTRIBUTING.md
├── SECURITY.md
└── ARCHITECTURE.md                # This file (symlinked in docs/)
```

---

## 3. Package Structure

The Go package structure follows Go's standard project layout conventions:

```
lhd (module: github.com/linuxhealthdoctor/lhd)
├── cmd/lhd/main.go         → package main
├── internal/cli/            → package cli
├── internal/core/           → package core
├── internal/plugin/         → package plugin
├── internal/checks/cpu/     → package cpu
├── internal/checks/memory/  → package memory
├── internal/checks/disk/    → package disk
├── internal/checks/...      → package each
├── internal/distro/         → package distro
├── internal/knowledge/      → package knowledge
├── internal/monitor/        → package monitor
├── internal/dashboard/      → package dashboard
├── internal/report/         → package report
├── internal/baseline/       → package baseline
├── internal/snapshot/       → package snapshot
├── internal/fleet/          → package fleet
├── internal/storage/        → package storage
├── internal/config/         → package config
├── internal/output/         → package output
├── internal/scheduler/      → package scheduler
└── pkg/version/             → package version
```

**Visibility Rule:** `internal/` ensures these packages cannot be imported by external consumers. The public API surface is `pkg/lhd/` for embedding LHD as a library.

---

## 4. Internal Modules

### 4.1 `core` — Domain Types & Shared Primitives

Foundation types used everywhere:

```go
// Severity levels (ordered)
type Severity int
const (
    SeverityNone Severity = iota
    SeverityInfo
    SeverityWarning
    SeverityCritical
    SeverityFatal
)

// CheckStatus for health checks
type CheckStatus int
const (
    StatusUnknown CheckStatus = iota
    StatusPass
    StatusFail
    StatusError
    StatusSkip
)

// SystemComponent categories
type Component string
const (
    ComponentCPU       Component = "cpu"
    ComponentMemory    Component = "memory"
    ComponentDisk      Component = "disk"
    ComponentNetwork   Component = "network"
    ComponentKernel    Component = "kernel"
    ComponentServices  Component = "services"
    ComponentSecurity  Component = "security"
    ComponentLogs      Component = "logs"
    ComponentHardware  Component = "hardware"
    ComponentUpdates   Component = "updates"
    ComponentContainers Component = "containers"
)

// Timestamp convenience
type Timestamp = time.Time
```

### 4.2 `plugin` — Plugin System

Manages registration and execution of health check plugins.

### 4.3 `checks` — Built-in Helm Check Implementations

Each sub-package is a self-contained health check category implementing the `plugin.Checker` interface.

### 4.4 `distro` — Distribution Abstraction Layer

Provides distro-specific implementations for package management, service management, file system paths, and system configuration.

### 4.5 `knowledge` — Knowledge Engine & RCA

Contains the rule-based inference engine, fact database, conclusion generation, and remediation mapping.

### 4.6 `monitor` — Continuous Monitoring Engine

Handles periodic data collection, threshold evaluation, alert generation, and time-series storage.

### 4.7 `dashboard` — Bubble Tea TUI

Elm-architecture TUI with real-time health display, system metrics, and interactive navigation.

### 4.8 `report` — Reporting Engine

Multi-format report generation from check results, baselines, or snapshots.

### 4.9 `baseline` — Baseline Management

Captures system state as a "healthy reference" and compares current state against it.

### 4.10 `snapshot` — System Snapshot

Captures full or partial system state for backup, comparison, or rollback reference.

### 4.11 `fleet` — Fleet Management

SSH-based remote execution, inventory management, aggregate reporting across multiple hosts.

### 4.12 `storage` — Persistence Layer

Abstract storage interface with SQLite and BoltDB implementations.

### 4.13 `config` — Configuration

Viper-based configuration loading with YAML, environment variable, and CLI flag support.

### 4.14 `output` — Terminal Output

Shared output formatting, colorization, progress bars, and table rendering.

### 4.15 `scheduler` — Job Scheduler

Cron-like scheduler for periodic task execution (monitoring, baseline comparison).

---

## 5. Interfaces

### 5.1 Plugin Checker Interface

```go
// Checker is the interface every health check plugin implements.
type Checker interface {
    // ID returns a unique identifier for this check.
    ID() string

    // Name returns a human-readable name.
    Name() string

    // Description returns what this check does.
    Description() string

    // Category returns the system component category.
    Category() core.Component

    // Check executes the health check and returns results.
    Check(ctx context.Context, req *CheckRequest) (*CheckResult, error)

    // Dependencies returns IDs of checks that must run first.
    Dependencies() []string

    // Tags returns metadata tags for filtering.
    Tags() []string
}

type CheckRequest struct {
    Baseline  *baseline.Profile  // Optional baseline for comparison
    Threshold *Threshold         // Custom thresholds
    Options   map[string]interface{}
}

type CheckResult struct {
    ID          string
    Status      core.CheckStatus
    Severity    core.Severity
    Message     string
    Details     map[string]interface{}
    Metrics     map[string]float64
    Timestamp   time.Time
    Duration    time.Duration
    Error       error
    Remediation []Remediation
    Evidence    []Evidence
}

type Remediation struct {
    Step        int
    Action      string
    Command     string   // Shell command to execute (user-verified only)
    Priority    int      // 1 = most urgent
    Reference   string   // Link to docs
    Risk        string   // "low", "medium", "high"
    AutoFixable bool     // Can lhd attempt fix (requires --fix flag)
}

type Evidence struct {
    Source    string   // e.g., /proc/cpuinfo, /var/log/syslog
    Raw       string   // Raw evidence snippet
    Interpret string   // Interpreted meaning
}
```

### 5.2 Distribution Interface

```go
type Distro interface {
    // ID returns distro identifier (e.g., "ubuntu", "arch")
    ID() string

    // Name returns full distro name
    Name() string

    // Version returns distro version string
    Version() string

    // PackageManager returns the package manager info
    PackageManager() PackageManager

    // ServiceManager returns the init system info
    ServiceManager() ServiceManager

    // FilePaths returns distro-specific file paths
    FilePaths() FilePaths

    // KernelInfo returns kernel-specific information
    KernelInfo() KernelInfo

    // SecurityInfo returns security subsystem info (SELinux, AppArmor, etc.)
    SecurityInfo() SecurityInfo

    // Compatible returns true if this distro matches the given ID
    Compatible(id string) bool
}

type PackageManager struct {
    Name    string   // "apt", "dnf", "pacman", "zypper", "emerge", "apk"
    Binary  string   // binary path
    CheckCmd string  // command to check updates
    InstallCmd string // command to install
    UpdateCmd string  // command to update DB
    ListCmd  string   // command to list installed
}

type ServiceManager struct {
    Name   string   // "systemd", "openrc", "runit", "s6"
    Binary string
}

type FilePaths struct {
    FSTab        string
    DefaultLogDir string
    ReposDir     string
    SSHConfig    string
}
```

### 5.3 Storage Interface

```go
type Store interface {
    // CheckResults
    SaveCheckResult(ctx context.Context, result *plugin.CheckResult) error
    QueryCheckResults(ctx context.Context, filter ResultFilter) ([]*plugin.CheckResult, error)
    LatestCheckResult(ctx context.Context, checkID string) (*plugin.CheckResult, error)

    // Baselines
    SaveBaseline(ctx context.Context, baseline *baseline.Baseline) error
    GetBaseline(ctx context.Context, id string) (*baseline.Baseline, error)
    ListBaselines(ctx context.Context) ([]*baseline.Baseline, error)
    DeleteBaseline(ctx context.Context, id string) error

    // Snapshots
    SaveSnapshot(ctx context.Context, snapshot *snapshot.Snapshot) error
    GetSnapshot(ctx context.Context, id string) (*snapshot.Snapshot, error)
    ListSnapshots(ctx context.Context) ([]*snapshot.Snapshot, error)
    DeleteSnapshot(ctx context.Context, id string) error

    // Metrics (BoltDB-backed)
    WriteMetric(ctx context.Context, m *Metric) error
    QueryMetrics(ctx context.Context, name string, from, to time.Time) ([]*Metric, error)
    LatestMetric(ctx context.Context, name string) (*Metric, error)

    // Events
    SaveEvent(ctx context.Context, event *core.Event) error
    QueryEvents(ctx context.Context, filter EventFilter) ([]*core.Event, error)

    // Health
    Health(ctx context.Context) error
    Close() error
}
```

### 5.4 Knowledge Engine Interface

```go
type KnowledgeEngine interface {
    // IngestFacts processes check results into facts
    IngestFacts(ctx context.Context, results []*plugin.CheckResult) error

    // Analyze performs rule-based inference on current facts
    Analyze(ctx context.Context) (*AnalysisResult, error)

    // AnalyzeWithBaseline compares against a known-good state
    AnalyzeWithBaseline(ctx context.Context, base *baseline.Baseline) (*AnalysisResult, error)

    // GetRemediations returns remediation steps for a conclusion
    GetRemediations(ctx context.Context, conclusionID string) ([]*plugin.Remediation, error)

    // LoadKnowledgeBase loads rules from YAML
    LoadKnowledgeBase(ctx context.Context, path string) error

    // AddRule adds a runtime rule
    AddRule(rule Rule) error
}

type AnalysisResult struct {
    ID              string
    Timestamp       time.Time
    OverallSeverity core.Severity
    Conclusions     []Conclusion
    RootCause       *Conclusion      // Primary root cause
    Chain           []Conclusion     // Causal chain
    Recommendations []string
    Remediations    []*plugin.Remediation
}

type Conclusion struct {
    ID          string
    Title       string
    Description string
    Severity    core.Severity
    Component   core.Component
    Certainty   float64          // 0.0 - 1.0
    Evidence    []string         // Fact IDs supporting this
    Related     []string         // Related conclusion IDs
    IsRootCause bool
}
```

### 5.5 Event Bus Interface

```go
type EventBus interface {
    Publish(topic string, event *core.Event)
    Subscribe(topic string, handler EventHandler) Subscription
    SubscribeAny(handler EventHandler) Subscription
    Unsubscribe(sub Subscription)
    Close()
}

type EventHandler func(ctx context.Context, topic string, event *core.Event)

type core.Event struct {
    ID        string
    Type      EventType
    Source    string
    Severity  core.Severity
    Message   string
    Data      map[string]interface{}
    Timestamp time.Time
}

type EventType string
const (
    EventCheckStart    EventType = "check.start"
    EventCheckComplete EventType = "check.complete"
    EventCheckFail     EventType = "check.fail"
    EventAlert         EventType = "alert"
    EventThreshold     EventType = "threshold.breach"
    EventSnapshot      EventType = "snapshot"
    EventBaseline      EventType = "baseline"
    EventError         EventType = "error"
)
```

### 5.6 Output/Formatter Interface

```go
type Formatter interface {
    // Format renders check results to the specified format
    Format(ctx context.Context, results []*plugin.CheckResult, opts *FormatOptions) ([]byte, error)

    // Extension returns the file extension for this format
    Extension() string

    // MIMEType returns the content type
    MIMEType() string
}

type FormatOptions struct {
    Color     bool
    Verbose   bool
    Template  string   // Custom template path
    Include   []string // Sections to include
    Exclude   []string // Sections to exclude
    SortBy    string
    Ascending bool
}
```

---

## 6. Configuration Design

### 6.1 Configuration Sources (priority order)

1. CLI flags (highest priority)
2. Environment variables (`LHD_*`)
3. Config file (`~/.config/lhd/lhd.yaml`) or (`/etc/lhd/lhd.yaml`)
4. Defaults (lowest priority)

### 6.2 Configuration File Format

```yaml
# ~/.config/lhd/lhd.yaml
version: "1.0"

# Global settings
global:
  data_dir: ~/.local/share/lhd
  log_dir: ~/.local/state/lhd/logs
  log_level: info           # debug, info, warn, error
  color: auto               # auto, always, never
  verbose: false
  timeout: 30s              # default check timeout
  max_parallel: 4           # max parallel check execution

# Doctor command configuration
doctor:
  categories:
    - cpu
    - memory
    - disk
    - network
    - kernel
    - services
    - security
    - logs
    - hardware
    - updates
    - containers
  exclude:
    - network.dns_slow     # exclude specific checks
  severity_threshold: warning  # minimum severity to report
  fix: false               # automatically apply auto-fixable remediations

# Monitoring configuration
monitor:
  enabled: true
  interval: 60s             # collection interval
  retention: 720h           # 30 days metric retention
  alerting:
    enabled: true
    desktop_notify: true    # via notify-send/dunst
    sound: false            # audible alerts
  thresholds:
    cpu:
      usage_percent:
        warning: 80
        critical: 95
      load_1m:
        warning: 4.0
        critical: 8.0
    memory:
      usage_percent:
        warning: 80
        critical: 95
      swap_usage_percent:
        warning: 50
        critical: 80
    disk:
      usage_percent:
        warning: 80
        critical: 92
      inode_percent:
        warning: 80
        critical: 92
    network:
      latency_ms:
        warning: 100
        critical: 500
      packet_loss_percent:
        warning: 1.0
        critical: 5.0

# Dashboard configuration
dashboard:
  refresh_interval: 5s
  default_view: overview     # overview, checks, monitor, logs
  compact: false            # compact mode for small terminals
  show_all_checks: false    # show all checks, not just relevant

# Reporting configuration
report:
  default_format: markdown  # json, yaml, csv, html, markdown, ansi
  include_evidence: true
  include_remediation: true
  template_dir: ~/.config/lhd/templates

# Baseline configuration
baseline:
  auto_capture: false       # auto-capture baseline on schedule
  schedule: "0 4 * * 0"    # weekly on Sunday 4am (cron)
  retention: 10             # keep last N baselines
  deviation_threshold:
    warning: 0.1            # 10% deviation triggers warning
    critical: 0.25          # 25% deviation triggers critical

# Snapshot configuration
snapshot:
  compress: true            # gzip compress snapshot data
  include_logs: false       # include truncated logs
  max_size: 100MB           # max uncompressed snapshot size
  exclude_patterns:
    - /tmp/**
    - /var/cache/**
    - /proc/**
    - /sys/**

# Fleet configuration
fleet:
  inventory_file: ~/.config/lhd/hosts.yaml
  ssh:
    default_port: 22
    timeout: 30s
    user: root
    key_file: ~/.ssh/id_rsa
    strict_host_key_checking: true
  concurrency: 10
  parallelism: per-host     # per-host, per-check
  retry: 2                  # retry failed connections
  bastion_host: ""          # optional jump host

# Security configuration
security:
  check_cves: true          # check for known CVEs (via installed packages)
  cve_db_update: 86400      # update CVE mapping daily (if online)
  custom_cve_db: ""         # path to custom CVE database
  fail_on_high: false       # exit non-zero on high severity findings
  selinux:
    check_enforcing: true
  apparmor:
    check_profiles: true
  auditd:
    check_running: true

# Custom checks
custom_checks:
  enabled: true
  dirs:
    - ~/.config/lhd/checks
    - /etc/lhd/checks
  timeout: 60s
  allowed_shebangs:
    - /bin/bash
    - /bin/sh
    - /usr/bin/python3

# Plugin management
plugins:
  check_updates: true       # check if newer check implementations available
  update_channel: stable    # stable, beta

# Knowledge base
knowledge:
  custom_dirs:
    - ~/.config/lhd/knowledge
    - /etc/lhd/knowledge
  severity_mapping:
    custom: "user_defined"  # allow custom severity levels
```

### 6.3 Environment Variables

All config keys map to `LHD_*` environment variables:

```bash
LHD_GLOBAL_DATA_DIR=~/.local/share/lhd
LHD_GLOBAL_LOG_LEVEL=debug
LHD_DOCTOR_FIX=true
LHD_MONITOR_INTERVAL=30s
LHD_FLEET_CONCURRENCY=20
```

### 6.4 Config Validation

Configuration is validated at startup. Validation rules:
- All durations must be valid Go `time.Duration` strings
- Threshold values must be in valid ranges (0-100 for percentages)
- Cron expressions are validated before use
- File paths are checked for existence where required
- Unknown keys trigger warnings (not errors) for forward compatibility

---

## 7. CLI Command Tree

```
lhd
├── doctor
│   ├── lhd doctor                       # Run all health checks
│   ├── lhd doctor --category cpu        # Run specific category
│   ├── lhd doctor --check cpu.load      # Run specific check
│   ├── lhd doctor --severity critical   # Only critical checks
│   ├── lhd doctor --fix                 # Apply auto-fixes
│   ├── lhd doctor --format json         # Output format
│   ├── lhd doctor --output report.json  # Save to file
│   ├── lhd doctor --baseline healthy    # Compare against baseline
│   └── lhd doctor --watch              # Watch mode (re-run)
│
├── diagnose
│   ├── lhd diagnose                     # Run diagnosis (doctor + RCA)
│   ├── lhd diagnose --interactive       # Interactive troubleshooting
│   ├── lhd diagnose --symptom "high cpu" # Symptom-guided diagnosis
│   └── lhd diagnose --from-report report.json
│
├── monitor
│   ├── lhd monitor start                # Start daemon monitoring
│   ├── lhd monitor stop                 # Stop monitoring daemon
│   ├── lhd monitor status               # Monitor daemon status
│   ├── lhd monitor logs                 # View monitor logs
│   ├── lhd monitor alerts               # View active alerts
│   ├── lhd monitor thresholds           # Show current thresholds
│   └── lhd monitor metrics              # Show collected metrics
│
├── dashboard
│   ├── lhd dashboard                    # Open interactive TUI
│   ├── lhd dashboard --read-only        # Read-only mode
│   └── lhd dashboard --compact          # Compact mode
│
├── report
│   ├── lhd report generate              # Generate report from last run
│   ├── lhd report generate --from baseline-2024-01-01
│   ├── lhd report generate --from snapshot-2024-01-01
│   ├── lhd report compare baseline1 baseline2
│   ├── lhd report compare snapshot1 snapshot2
│   ├── lhd report list                  # List generated reports
│   ├── lhd report view <id>             # View a report
│   └── lhd report export <id> --format pdf
│
├── baseline
│   ├── lhd baseline capture             # Capture current state as baseline
│   ├── lhd baseline capture --name "post-update"
│   ├── lhd baseline list                # List baselines
│   ├── lhd baseline show <id>           # Show baseline details
│   ├── lhd baseline compare <id>        # Compare against baseline
│   ├── lhd baseline diff <id1> <id2>    # Diff two baselines
│   ├── lhd baseline delete <id>         # Delete baseline
│   └── lhd baseline schedule            # Schedule auto-capture
│
├── snapshot
│   ├── lhd snapshot create              # Create system snapshot
│   ├── lhd snapshot create --name "pre-update"
│   ├── lhd snapshot list                # List snapshots
│   ├── lhd snapshot show <id>           # Show snapshot details
│   ├── lhd snapshot diff <id1> <id2>    # Diff two snapshots
│   ├── lhd snapshot compare <id>        # Compare against snapshot
│   └── lhd snapshot delete <id>         # Delete snapshot
│
├── fleet
│   ├── lhd fleet init                   # Initialize fleet inventory
│   ├── lhd fleet add hostname          # Add host to inventory
│   ├── lhd fleet remove hostname       # Remove host
│   ├── lhd fleet list                   # List fleet hosts
│   ├── lhd fleet doctor                 # Run doctor across fleet
│   ├── lhd fleet diagnose               # Run diagnose across fleet
│   ├── lhd fleet report                 # Generate aggregate report
│   ├── lhd fleet baseline               # Capture baseline on fleet
│   ├── lhd fleet snapshot               # Capture snapshot on fleet
│   ├── lhd fleet ping                   # Test connectivity
│   ├── lhd fleet sync                   # Sync config to fleet
│   └── lhd fleet status                 # Fleet health overview
│
├── config
│   ├── lhd config init                  # Create default config
│   ├── lhd config show                  # Show current config
│   ├── lhd config set <key> <value>     # Set config value
│   ├── lhd config get <key>             # Get config value
│   ├── lhd config validate              # Validate config file
│   ├── lhd config path                  # Show config file path
│   └── lhd config reset                 # Reset to defaults
│
├── version
│   └── lhd version                      # Show version info
│
└── completion
    ├── lhd completion bash              # Generate bash completion
    ├── lhd completion zsh               # Generate zsh completion
    └── lhd completion fish              # Generate fish completion
```

---

## 8. Plugin Architecture

### 8.1 Design

Plugins are **compiled-in Go packages** implementing the `plugin.Checker` interface. This decision prioritizes:

- **Single binary distribution** — no separate plugin files to manage
- **Type safety** — compile-time interface checking
- **Performance** — no IPC/runtime loading overhead
- **Security** — no arbitrary code execution, no WASM sandbox
- **Simplicity** — `go build` produces a complete binary

### 8.2 Plugin Lifecycle

```
Registration → Discovery → Ordering → Execution → Aggregation
```

1. **Registration**: Each plugin registers itself via `init()` in its package, calling `plugin.Register()`
2. **Discovery**: The registry maintains a map of `category → []Checker`
3. **Ordering**: DAG-based topological sort resolves dependencies
4. **Execution**: Parallel execution respecting dependency order and `max_parallel`
5. **Aggregation**: Results collected, deduplicated, and formatted

### 8.3 Plugin Registry

```go
var global = &Registry{
    checkers: make(map[string]Checker),
    categories: make(map[core.Component][]Checker),
}

func Register(c Checker) { global.Register(c) }
func Get(id string) Checker { return global.Get(id) }
func List(opts ...ListOption) []Checker { return global.List(opts...) }

type Registry struct {
    mu         sync.RWMutex
    checkers   map[string]Checker
    categories map[core.Component][]Checker
}

func (r *Registry) Register(c Checker) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.checkers[c.ID()] = c
    r.categories[c.Category()] = append(r.categories[c.Category()], c)
}

func (r *Registry) ResolveExecutionPlan(categories []core.Component) ([][]Checker, error) {
    // 1. Filter checkers by requested categories
    // 2. Build dependency graph
    // 3. Topological sort
    // 4. Return execution layers (parallel within layers, sequential across)
    // Uses Kahn's algorithm for DAG resolution
}
```

### 8.4 Custom User Checks

While built-in checks are compiled in, users can add custom shell script checks:

```
~/.config/lhd/checks/
├── 10-disk-health.sh
├── 20-custom-app-check.sh
└── 30-database-check.sh
```

These are executed as subprocesses with a strict sandbox:
- Timeout enforced (configurable, default 60s)
- Output parsed from stdout as JSON (`{ "status": "pass|fail", "message": "...", "metrics": {} }`)
- Only shebangs from an allowlist are permitted
- No network access (via `unshare` if available or documented restriction)
- Run as the current user (not root)

### 8.5 Built-in Check Categories

| Category | Checks |
|---|---|
| `cpu` | Usage, load average, temperature, frequency scaling, governor, throttling, vulnerabilities |
| `memory` | RAM usage, swap usage, OOM status, NUMA balancing, hugepages, memory pressure |
| `disk` | Disk usage, inode usage, I/O stats, S.M.A.R.T health, filesystem errors, mount options |
| `network` | Connectivity, DNS resolution, latency, packet loss, interface errors, socket limits, conntrack |
| `kernel` | Kernel version, sysctl anomalies, module issues, dmesg errors, entropy, lockdown status |
| `services` | Failed units, stopped critical services, socket activation, timer failures |
| `security` | SELinux/AppArmor status, auditd, open ports, SUID files, world-writable files, passwords |
| `logs` | Recent errors, rate-limited logging, journald health, log rotation status |
| `hardware` | CPU/memory errors (EDAC), NVMe health, disk temperature, power supply, fan speed |
| `updates` | Available package updates, security updates, snap/flatpak updates, kernel livepatch |
| `containers` | Docker/Podman daemon status, container health, runtime issues, image vulnerabilities |

---

## 9. Distribution Abstraction Layer

### 9.1 Architecture

The DAL provides a uniform interface to distro-specific functionality. It is essential for a tool that claims to work on "any Linux."

```
┌──────────────────────────────────────────┐
│              Distro Interface            │
├──────────────────────────────────────────┤
│  PackageManager() ServiceManager()       │
│  FilePaths() SecurityInfo()             │
│  Compatible()                            │
└──────────────────────────────────────────┘
          ▲              ▲              ▲
          │              │              │
┌─────────┴──┐   ┌──────┴──────┐   ┌───┴──────────┐
│  Debian     │   │   RHEL      │   │   Arch        │
│  - apt      │   │   - dnf     │   │   - pacman    │
│  - systemd  │   │   - systemd │   │   - systemd   │
│  - /etc/apt │   │   - /etc/yum│   │   - /etc/pacman│
│  - AppArmor │   │   - SELinux │   │   - none      │
└─────────────┘   └─────────────┘   └──────────────┘
          ▲              ▲              ▲
          │              │              │
┌─────────┴──┐   ┌──────┴──────┐   ┌───┴──────────┐
│  Alpine     │   │  Gentoo     │   │   NixOS       │
│  - apk      │   │  - emerge   │   │   - nix-env   │
│  - openrc   │   │  - systemd  │   │   - systemd   │
│  - /etc/apk │   │  - /etc/...│   │   - /nix      │
└─────────────┘   └─────────────┘   └──────────────┘
```

### 9.2 Detection Logic

```go
func Detect() (Distro, error) {
    // 1. Check /etc/os-release (ID=, ID_LIKE=)
    // 2. Check /etc/lsb-release
    // 3. Check /etc/arch-release (exists = Arch)
    // 4. Check /etc/gentoo-release
    // 5. Check /etc/alpine-release
    // 6. Check /etc/SuSE-release
    // 7. Fallback: uname -a parsing
    // 8. Return UnknownDistro as last resort
}
```

### 9.3 Service Manager Detection

```go
func detectServiceManager() ServiceManager {
    // Check for: systemctl, rc-service, runsvdir, s6-svscan
    // Test in order of prevalence: systemd > openrc > runit > s6
    // Cache result for process lifetime
}
```

### 9.4 Distro-Specific Override Points

| Operation | Default Implementation | Distro-Specific |
|---|---|---|
| Package list | Parse dpkg/apt | rpm -qa, pacman -Q, etc. |
| Service status | systemctl status | rc-service status, sv status |
| Log location | /var/log | Arch: /var/log, Gentoo: /var/log |
| Repo list | /etc/apt/sources.list | /etc/yum.repos.d/, /etc/pacman.d/ |
| Security modules | Check both SELinux & AppArmor | Prioritize distro default |
| Filesystem paths | /etc/fstab | NixOS: /etc/nixos/configuration.nix |

---

## 10. Root Cause Analysis Architecture

### 10.1 Philosophy

RCA in LHD is a **rule-based inference system** — deterministic, auditable, and explainable. It does not use machine learning, neural networks, or external AI services. Every conclusion can be traced back to specific rules and evidence.

### 10.2 RCA Pipeline

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────────┐
│  Doctor   │───▶│  Facts   │───▶│   Rule   │───▶│  Conclusions │
│  Results  │    │  Engine  │    │  Engine  │    │  & RCA       │
└──────────┘    └──────────┘    └──────────┘    └──────────────┘
                      │               │                │
                      ▼               ▼                ▼
               ┌──────────┐    ┌──────────┐    ┌──────────────┐
               │  Baseline│    │  Domain  │    │  Remediation  │
               │  Facts   │    │  Rules   │    │  Engine       │
               └──────────┘    └──────────┘    └──────────────┘
```

### 10.3 Fact Model

```go
type Fact struct {
    ID          string
    Type        string      // "metric", "state", "event", "config"
    Component   core.Component
    Name        string      // e.g., "cpu.usage_percent"
    Value       interface{} // Must be comparable
    Threshold   float64     // Baseline/threshold value
    Severity    core.Severity
    Timestamp   time.Time
    Source      string      // Check ID or data source
    Confidence  float64     // 0.0 - 1.0
}
```

### 10.4 Rule Model

```go
type Rule struct {
    ID          string
    Name        string
    Description string
    Component   core.Component
    Severity    core.Severity
    Condition   Condition     // When this condition is true...
    Conclusion  string        // ...this conclusion is reached
    Certainty   float64       // Confidence in this rule (0.0 - 1.0)
    Weight      int           // Rule priority (higher = more important)
}

type Condition struct {
    Operator    string      // "and", "or", "not"
    Conditions  []Condition // Nested conditions (for compound logic)
    FactCheck   *FactCheck  // Leaf condition: check a fact
}

type FactCheck struct {
    FactName      string
    Operator    string    // "eq", "neq", "gt", "gte", "lt", "lte", "contains", "regex"
    Value       interface{}
    Threshold   *float64  // Optional threshold (for "deviation" operator)
}
```

### 10.5 RCA Inference Algorithm

```go
func (e *Engine) Analyze(ctx context.Context, facts []Fact) (*AnalysisResult, error) {
    // Phase 1: Evaluate all rules against facts
    // Phase 2: Generate conclusions from matching rules
    // Phase 3: Build causal chains (conclusion A → conclusion B)
    // Phase 4: Identify root cause (earliest node in causal chain with highest severity)
    // Phase 5: Map remediations to conclusions
    // Phase 6: Score overall severity

    // Returns AnalysisResult with:
    // - Overall severity
    // - All conclusions
    // - Root cause identification
    // - Causal chain
    // - Remediations
}
```

### 10.6 Example Rules (Knowledge Base)

```yaml
# builtin/performance.yaml
rules:
  - id: perf.cpu.high_load
    name: "High CPU Load"
    description: "System CPU load average exceeds thresholds"
    component: cpu
    severity: warning
    condition:
      operator: "or"
      conditions:
        - fact_check:
            fact_name: "cpu.load_1m"
            operator: "gt"
            value: 4.0
        - fact_check:
            fact_name: "cpu.load_5m"
            operator: "gt"
            value: 3.0
    conclusion: "CPU load is elevated, indicating system may be under heavy load"
    certainty: 0.85
    weight: 10

  - id: perf.cpu.high_usage_degraded
    name: "Sustained High CPU Usage"
    description: "CPU usage > 90% for extended period"
    component: cpu
    severity: critical
    condition:
      operator: "and"
      conditions:
        - fact_check:
            fact_name: "cpu.usage_percent"
            operator: "gt"
            value: 90.0
        - fact_check:
            fact_name: "cpu.load_1m"
            operator: "gt"
            threshold: 0.8
            # Uses baseline: if load_1m > 80% of baseline value
    conclusion: "Sustained high CPU usage with degraded responsiveness"
    certainty: 0.9
    weight: 20

  - id: causality.cpu_memory
    name: "CPU-Memory Correlation"
    description: "High CPU caused by memory pressure (swap thrashing)"
    component: cpu
    severity: critical
    condition:
      operator: "and"
      conditions:
        - fact_check:
            fact_name: "cpu.usage_percent"
            operator: "gt"
            value: 80.0
        - fact_check:
            fact_name: "memory.swap_io_percent"
            operator: "gt"
            value: 50.0
        - fact_check:
            fact_name: "memory.usage_percent"
            operator: "gt"
            value: 90.0
    conclusion: "High CPU is caused by memory pressure — kswapd is thrashing"
    certainty: 0.95
    weight: 30
```

### 10.7 Causal Chain Construction

```
Fact: cpu.usage_percent = 95%  ──────┐
                                     ├──▶ Conclusion: High CPU usage
Fact: memory.usage_percent = 92% ────┤         │
                                     │         │
Fact: memory.swap_io = high ─────────┘         │
                                                ▼
                                     ┌─────────────────────┐
                                     │ Root Cause: Memory   │
                                     │ Pressure → Swap      │
                                     │ Thrashing → High CPU │
                                     │ (kswapd)             │
                                     └─────────────────────┘
```

The engine identifies the root cause as the conclusion with the most incoming causal edges and the earliest position in the dependency chain.

---

## 11. Reporting Architecture

### 11.1 Design

Reports are generated from check results, baselines, or snapshots. The reporting engine uses a pipeline pattern:

```
Data Source → Template Engine → Formatter → Output
```

### 11.2 Data Sources

| Source | Description |
|---|---|
| `doctor` results | Latest or historical check results |
| `diagnose` results | Doctor results + RCA conclusions |
| `baseline` | Baseline profile data |
| `snapshot` | Snapshot manifest and data |
| `monitor` | Time-series metrics over a window |

### 11.3 Report Types

| Report Type | Content | Best Format |
|---|---|---|
| **Health Report** | Check results, severity overview, pass/fail counts | Markdown, HTML |
| **Diagnosis Report** | Health + RCA conclusions, root cause, causal chain | HTML, JSON |
| **Comparison Report** | Diff between two baselines/snapshots | Markdown, HTML |
| **Trend Report** | Metric trends over time (from monitor data) | HTML, CSV |
| **Fleet Report** | Aggregate health across fleet hosts | HTML, JSON |
| **Executive Summary** | High-level health score, top issues | Markdown, ANSI |
| **Compliance Report** | Check results against compliance policies | PDF, HTML |

### 11.4 Formatters

```go
type FormatRegistry struct {
    formatters map[string]Formatter
}

var formats = &FormatRegistry{
    formatters: map[string]Formatter{
        "json":     &JSONFormatter{pretty: true},
        "yaml":     &YAMLFormatter{},
        "csv":      &CSVFormatter{},
        "html":     &HTMLFormatter{},
        "markdown": &MarkdownFormatter{},
        "ansi":     &ANSITerminalFormatter{},
        "pdf":      &PDFFormatter{},
    },
}
```

### 11.5 Template System

Reports use Go's `text/template` and `html/template` for customizable output:

```go
type TemplateEngine struct {
    builtin embed.FS       // Built-in templates via //go:embed
    custom  map[string]*template.Template
}

func (t *TemplateEngine) Render(name string, data *ReportData) ([]byte, error) {
    // 1. Try custom template from config directory
    // 2. Fall back to built-in template
    // 3. Execute with data
    // 4. Return rendered bytes
}
```

### 11.6 Health Score Calculation

```go
type HealthScore struct {
    Overall     float64   // 0.0 (critical) - 100.0 (perfect)
    Category    map[core.Component]float64
    Trend       string    // "improving", "stable", "degrading"
    ScoreChange float64   // Change from last report
}

func CalculateHealthScore(results []*plugin.CheckResult) *HealthScore {
    // Each check result has a weighted score contribution
    // Critical failures reduce score more than warnings
    // Pass = full points, Warning = 50% points, Fail = 0 points
    // Final score = (sum of weighted scores / max possible score) * 100
}
```

---

## 12. Dashboard Architecture

### 12.1 Design

The dashboard is a **Bubble Tea TUI application** following the Elm architecture:

```
┌──────────────────────────────────────────────────────────────────┐
│  Model (State) → Update (Msg) → View (Render) → Model (loop)    │
└──────────────────────────────────────────────────────────────────┘
```

### 12.2 Model

```go
type Model struct {
    // Navigation
    activeTab    Tab
    previousTab  Tab

    // System info
    hostname    string
    distro      string
    uptime      time.Duration

    // Health data
    lastDoctorRun *plugin.AggregatedResult
    healthScore   *HealthScore
    activeIssues  []*plugin.CheckResult

    // Monitoring data (real-time)
    metrics     map[string]*MetricBuffer  // Circular buffers
    cpuChart    *Sparkline
    memChart    *Sparkline
    diskChart   *Sparkline
    netChart    *Sparkline

    // Alerts
    alerts      []*Alert

    // TUI state
    width       int
    height      int
    ready       bool
    loading     bool
    error       error
    keys        keymap

    // Sub-models (for tabs)
    overview    OverviewModel
    checks      ChecksModel
    monitor     MonitorModel
    logs        LogsModel
    settings    SettingsModel

    // Tickers
    refreshTick *time.Ticker
    metricsTick *time.Ticker

    // Styles
    styles      Styles
    help        help.Model
    spinner     spinner.Model
    table       table.Model
}

type Tab int
const (
    TabOverview Tab = iota
    TabChecks
    TabMonitor
    TabLogs
    TabSettings
)
```

### 12.3 Views

| View | Content |
|---|---|
| **Overview** | Hostname, distro, uptime, health score gauge, top issues, quick stats |
| **Checks** | Check results table with status, severity, message — filterable by category |
| **Monitor** | Real-time sparkline charts for CPU, memory, disk, network — threshold indicators |
| **Logs** | Recent system log entries with severity coloring, search/filter |
| **Settings** | Configuration summary, data paths, version info |

### 12.4 Key Bindings

```
Tab / Shift+Tab    : Navigate tabs
↑ ↓ k j           : Scroll / Navigate lists
Enter             : Select / Drill down
/                 : Search
r                 : Refresh
q / Ctrl+C        : Quit
?                 : Toggle help
Esc               : Back / Close
```

### 12.5 Color Scheme

```go
type Styles struct {
    // Severity colors
    Critical    lipgloss.Color
    Warning     lipgloss.Color
    Info        lipgloss.Color
    Pass        lipgloss.Color
    Unknown     lipgloss.Color

    // UI
    Background  lipgloss.Color
    Foreground  lipgloss.Color
    Border      lipgloss.Color
    Highlight   lipgloss.Color
    Dimmed      lipgloss.Color
}
```

### 12.6 Rendering Loop

```go
func (m Model) Init() tea.Cmd {
    return tea.Batch(
        m.spinner.Tick,
        m.refreshTick,         // Refresh health data every 5s
        m.metricsTick,         // Refresh metrics every 2s
        m.initialLoad(),       // Initial data fetch
    )
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.ready = true
    case tea.KeyMsg:
        return m.handleKeyMsg(msg)
    case RefreshMsg:
        return m.refreshData()
    case MetricsMsg:
        return m.refreshMetrics()
    case error:
        m.error = msg
        return m, nil
    }
    return m, nil
}

func (m Model) View() string {
    if !m.ready {
        return m.spinner.View() + " Loading..."
    }

    header := m.renderHeader()
    tabs := m.renderTabs()
    content := m.renderActiveTab()
    footer := m.renderFooter()
    help := m.help.View()

    return lipgloss.JoinVertical(lipgloss.Top,
        header,
        tabs,
        content,
        footer,
        help,
    )
}
```

---

## 13. Security Architecture

### 13.1 Security Principles

1. **Least privilege** — LHD runs at user privilege level by default; `sudo` only when necessary
2. **Defense in depth** — Multiple layers of validation, sandboxing, and restriction
3. **No network telemetry** — Zero data leaves the machine (by design and enforcement)
4. **Deterministic operation** — Same inputs always produce same outputs (reproducible diagnostics)
5. **Transparent operation** — Every action is logged and explainable

### 13.2 Privilege Model

```
┌──────────────────────────────────────────────────────────────┐
│                    lhd Process                                │
├──────────────────────────────────────────────────────────────┤
│  Normal Mode (default)                                       │
│  - Read /proc, /sys (world-readable)                         │
│  - Read logs (via journalctl user, /var/log accessible)      │
│  - Read config files                                         │
│  - Run shell commands for custom checks                      │
│  - Write to ~/.local/share/lhd/                              │
├──────────────────────────────────────────────────────────────┤
│  Privileged Mode (sudo lhd)                                  │
│  - Full system access                                        │
│  - S.M.A.R.T disk health                                     │
│  - Kernel audit log                                          │
│  - All service statuses                                      │
│  - Security module checks (SELinux, AppArmor)                │
│  - Package management operations                             │
└──────────────────────────────────────────────────────────────┘
```

### 13.3 Security Checks

| Check | What It Tests | Requires Privilege |
|---|---|---|
| Open ports | Unintended listening services | No (netstat/ss) |
| SUID files | Unexpected SUID binaries | Yes |
| World-writable | Dangerous file permissions | Yes (system dirs) |
| Password policy | Weak password configuration | Yes |
| SSH config | SSH hardening (PermitRootLogin, etc.) | No |
| Kernel hardening | KASLR, SMEP, SMAP, KPTI status | No |
| SELinux/AppArmor | Enforcement mode, loaded profiles | Yes |
| Auditd | Audit daemon running, rules loaded | Yes |
| Firewall | iptables/nftables active, rules | Yes (read) |
| CVE scan | Known vulnerabilities in installed packages | No |
| Failed logins | Recent failed authentication attempts | No |
| User audit | Stale users, empty passwords, uid 0 | Yes |

### 13.4 Custom Check Sandboxing

```go
type SandboxConfig struct {
    Timeout       time.Duration
    AllowedPaths  []string    // Read-only paths
    DeniedPaths   []string    // Explicitly denied
    AllowedEnv    []string    // Allowlisted environment variables (PATH, HOME)
    NetworkAccess bool        // Default: false
    MaxOutputSize int64       // Default: 1MB
    AllowedUsers  []string    // Only allow running as these users
}
```

### 13.5 Secure Storage

- SQLite database permissions: `0600`
- BoltDB file permissions: `0600`
- Config file permissions: `0600` (can contain SSH keys for fleet)
- No secrets in process environment
- SSH agent forwarding for fleet (no stored keys on disk)

### 13.6 Supply Chain Security

- `go.mod` with pinned dependencies (no `latest`)
- `go.sum` for checksum verification
- Reproducible builds via Go's buildid
- Signed releases via GoReleaser + Cosign
- SBOM generation per release
- Dependabot/Security alerts enabled

---

## 14. Fleet Architecture

### 14.1 Design

Fleet management treats remote hosts as extension of local diagnostics. Communication is exclusively via SSH — no agents, no daemons, no cloud.

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Control Node                                  │
│  lhd fleet doctor --hosts all                                        │
└──────┬──────────┬──────────┬──────────┬──────────┬──────────┬───────┘
       │          │          │          │          │          │
       ▼          ▼          ▼          ▼          ▼          ▼
    ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐
    │Host 1│  │Host 2│  │Host 3│  │Host 4│  │Host 5│  │Host N│
    │LHD   │  │LHD   │  │LHD   │  │LHD   │  │LHD   │  │LHD   │
    └──────┘  └──────┘  └──────┘  └──────┘  └──────┘  └──────┘
```

### 14.2 Execution Modes

| Mode | Description | Use Case |
|---|---|---|
| **SSH Push** | Control node SSHes into each host, runs lhd, collects results | Ad-hoc diagnostics |
| **SSH Pull** (future) | Hosts push results to a central collector (requires daemon) | Regular reporting |
| **File-based** (future) | Hosts write results to shared filesystem | Air-gapped environments |

### 14.3 Inventory Format

```yaml
# ~/.config/lhd/hosts.yaml
version: "1"
defaults:
  user: root
  port: 22
  timeout: 30s
  parallelism: per-host

hosts:
  - hostname: web-01
    address: 10.0.1.10
    alias: "Production Web Server 1"
    tags:
      - production
      - web
      - critical
    labels:
      region: us-east-1
      role: webserver
    ssh:
      user: admin
      port: 2222
    checks:
      exclude:
        - containers.*     # No containers on this host

  - hostname: db-01
    address: 10.0.1.20
    tags:
      - production
      - database
      - critical
    checks:
      categories:
        - cpu
        - memory
        - disk
        - kernel

  - hostname: dev-*
    tags:
      - development
    # Glob pattern matching — all hosts matching dev-*
```

### 14.4 Parallel Execution Engine

```go
type FleetExecutor struct {
    semaphore   chan struct{}    // Concurrency limiter
    sshPool     *ssh.Pool       // Connection pool
    results     sync.Map        // Thread-safe result aggregation
}

func (e *FleetExecutor) Execute(ctx context.Context, plan *ExecutionPlan) (*FleetResult, error) {
    results := make(chan *HostResult, len(plan.Hosts))
    ctx, cancel := context.WithTimeout(ctx, plan.Timeout)
    defer cancel()

    for _, host := range plan.Hosts {
        e.semaphore <- struct{}{}  // Acquire slot
        go func(h Host) {
            defer func() { <-e.semaphore }()
            result := e.executeOnHost(ctx, h, plan.Command)
            results <- result
        }(host)
    }

    return e.aggregateResults(ctx, results, len(plan.Hosts))
}
```

### 14.5 Result Aggregation

```go
type FleetResult struct {
    Timestamp       time.Time
    TotalHosts      int
    SuccessfulHosts int
    FailedHosts     int
    Results         []*HostResult
    HealthScore     *HealthScore
    CategorySummary map[core.Component]*CategorySummary
    TopIssues       []*plugin.CheckResult
    Duration        time.Duration
}

type HostResult struct {
    Hostname  string
    Address   string
    Success   bool
    Error     error
    Duration  time.Duration
    Health    *plugin.AggregatedResult
    Baseline  *baseline.Baseline
    Snapshot  *snapshot.Snapshot
    Distro    string
    Version   string
}
```

### 14.6 Security Considerations

- SSH keys stored in config file with `0600` permissions
- Support for SSH agent (recommended over stored keys)
- Strict host key checking enabled by default
- Known hosts management with `lhd fleet init`
- Bastion/jump host support for segmented networks
- Connection logging for audit trail

---

## 15. Snapshot Architecture

### 15.1 Design

A snapshot captures the system state at a point in time for comparison, rollback reference, or migration planning.

### 15.2 Snapshot Contents

```go
type Snapshot struct {
    ID          string
    Name        string
    Description string
    Timestamp   time.Time
    Distro      string
    Kernel      string
    Hostname    string

    Categories map[core.Component]*CategoryData
    Manifest   *Manifest
    Size       int64
    Compressed bool
    Checksum   string
}

type CategoryData struct {
    Component   core.Component
    Data        map[string]interface{}   // Structured data
    Raw         map[string][]byte        // Raw file contents (controlled)
    FileCount   int
    TotalSize   int64
}

type Manifest struct {
    Files        []FileEntry
    Packages     []PackageEntry
    Services     []ServiceEntry
    Network      *NetworkState
    Processes    []ProcessEntry
    Mounts       []MountEntry
    KernelParams map[string]string
    Sysctl       map[string]string
    Env          map[string]string
}

type FileEntry struct {
    Path        string
    Size        int64
    Mode        os.FileMode
    UID         int
    GID         int
    SHA256      string
    Symlink     string  // If symlink
    Content     []byte  // Only for text config files (controlled)
}

type PackageEntry struct {
    Name       string
    Version    string
    Repository string
    Size       int64
    Installed  time.Time
}

type ServiceEntry struct {
    Name      string
    Status    string  // active, inactive, enabled, disabled
    LoadState string
    Uptime    time.Duration
}
```

### 15.3 Snapshot Storage

Snapshots are stored in the data directory:

```
~/.local/share/lhd/
├── snapshots/
│   ├── 2024-01-15T10-30-00-pre-upgrade/
│   │   ├── manifest.json        # Snapshot metadata
│   │   ├── categories/
│   │   │   ├── cpu.json
│   │   │   ├── memory.json
│   │   │   ├── disk.json
│   │   │   ├── network.json
│   │   │   ├── kernel.json
│   │   │   ├── services.json
│   │   │   ├── packages.json
│   │   │   ├── files.json.gz    # Compressed file contents
│   │   │   └── proc.json
│   │   └── checksums.txt
│   └── ...
└──
```

### 15.4 Diff Engine

```go
type SnapshotDiff struct {
    ID              string
    Left            string
    Right           string
    Timestamp       time.Time
    Changes         []Change
    Summary         DiffSummary

    // Per-category
    PackageDiff     *PackageDiff
    FileDiff        *FileDiff
    ServiceDiff     *ServiceDiff
    ConfigDiff      *ConfigDiff
    KernelDiff      *KernelDiff
}

type Change struct {
    Type        ChangeType   // added, removed, modified, unchanged
    Category    core.Component
    Path        string       // For files
    Key         string       // For key-value data
    OldValue    interface{}
    NewValue    interface{}
    Severity    core.Severity
}

type ChangeType string
const (
    ChangeAdded     ChangeType = "added"
    ChangeRemoved   ChangeType = "removed"
    ChangeModified  ChangeType = "modified"
    ChangeUnchanged ChangeType = "unchanged"
)
```

---

## 16. Baseline Architecture

### 16.1 Design

A baseline represents a "healthy" reference state for the system. Unlike snapshots (comprehensive captures), baselines are focused on metrics and configurations that define normal operation.

### 16.2 Baseline Profile

```go
type Baseline struct {
    ID          string
    Name        string
    Description string
    Timestamp   time.Time
    ExpiresAt   *time.Time           // Optional expiration

    // System identity
    Distro      string
    Kernel      string
    Hostname    string

    // Metric baselines (statistical)
    Metrics     map[string]*MetricBaseline

    // Configuration baselines
    Config      map[string]*ConfigBaseline

    // Service baselines
    Services    map[string]*ServiceBaseline

    // Performance baselines
    Performance *PerformanceBaseline

    // Security baselines
    Security    *SecurityBaseline

    // Metadata
    Version     int
    ParentID    string               // For baseline hierarchy
    Tags        []string
}

type MetricBaseline struct {
    Name        string
    Unit        string
    Mean        float64
    Median      float64
    P95         float64
    P99         float64
    StdDev      float64
    Min         float64
    Max         float64
    Samples     int
}

type ConfigBaseline struct {
    Path        string
    Key         string
    Expected    string
    Regex       string          // Alternative: regex match
    MustExist   bool
    Permissions os.FileMode
    Owner       string
}

type ServiceBaseline struct {
    Name        string
    ExpectedStatus string     // active, inactive
    ExpectedEnabled bool
    MinUptime   time.Duration
}

type PerformanceBaseline struct {
    BootTime    time.Duration
    DiskReadAhead int
    Swappiness  int
    FSBlockSize int
}
```

### 16.3 Comparison

```go
func (e *Engine) Compare(ctx context.Context, base *Baseline, current *plugin.AggregatedResult) (*ComparisonResult, error) {
    deviations := []Deviation{}

    // For each metric in baseline, compare with current value
    for name, baselineMetric := range base.Metrics {
        currentValue, ok := current.Metrics[name]
        if !ok {
            deviations = append(deviations, Deviation{
                Metric: name,
                Type:   DeviationMissing,
                Severity: SeverityWarning,
            })
            continue
        }

        // Calculate deviation percentage
        deviation := math.Abs(currentValue - baselineMetric.Mean) / baselineMetric.Mean
        severity := SeverityInfo
        if deviation > 0.25 { severity = SeverityCritical
        } else if deviation > 0.10 { severity = SeverityWarning }

        deviations = append(deviations, Deviation{
            Metric:      name,
            Type:        DeviationThreshold,
            Expected:    baselineMetric.Mean,
            Actual:      currentValue,
            Deviation:   deviation,
            Severity:    severity,
        })
    }

    return &ComparisonResult{
        Baseline:     base,
        Timestamp:    time.Now(),
        Deviations:   deviations,
        TotalChecks:  len(base.Metrics),
        PassedChecks: countPassed(deviations),
        HealthScore:  calculateScore(deviations),
        Summary:      generateSummary(deviations),
    }, nil
}
```

### 16.4 Baseline Lifecycle

```
Capture → Store → Compare → Notify → Update/Expire
   │        │        │          │          │
   │        │        │          │          └── Auto-expire old baselines
   │        │        │          └────────────── Alert on significant deviation
   │        │        └───────────────────────── Compare doctor results against baseline
   │        └────────────────────────────────── Store with versioning
   └─────────────────────────────────────────── Capture system metrics
```

---

## 17. Knowledge Engine Architecture

### 17.1 Design

The knowledge engine is a forward-chaining rule-based inference system. It is the core intelligence of LHD, transforming raw check results into meaningful diagnoses, root causes, and remediation plans.

### 17.2 Components

```
┌────────────────────────────────────────────────────────────────┐
│                     Knowledge Engine                            │
├────────────────────────────────────────────────────────────────┤
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌───────────────┐  │
│  │  Fact    │  │  Rule    │  │  Inference                     │  │
│  │  Manager │  │  Engine  │  │  Engine   │  │  Explanation  │  │
│  └──────────┘  └──────────┘  └──────────┘  │  Generator    │  │
│       │             │             │          └───────────────┘  │
│       ▼             ▼             ▼                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                     │
│  │  Fact DB  │  │  Rule DB │  │  Conclusion                    │
│  │ (in memory)│  │ (YAML)  │  │  Store   │                     │
│  └──────────┘  └──────────┘  └──────────┘                     │
└────────────────────────────────────────────────────────────────┘
```

### 17.3 Rule Categories

| Category | Purpose | Example |
|---|---|---|
| **Threshold Rules** | Compare metrics against fixed thresholds | "CPU > 90% → Warning" |
| **Baseline Rules** | Compare metrics against baseline | "Memory > 120% of baseline → Warning" |
| **Correlation Rules** | Identify multi-metric patterns | "High CPU + High I/O → Possible swap thrashing" |
| **Temporal Rules** | Detect trends over time | "Disk usage increased 10% in 1 hour → Investigate" |
| **Boolean Rules** | Check binary system states | "SELinux disabled → Security warning" |
| **Anomaly Rules** | Detect statistical outliers | "Connection count > 3σ from mean → Possible attack" |

### 17.4 Inference Algorithm

```go
func (e *Engine) Infer(ctx context.Context) (*AnalysisResult, error) {
    // Phase 1: Forward chaining
    //   - Iterate rules in priority order
    //   - Evaluate conditions against facts
    //   - Generate new facts from conclusions
    //   - Continue until no new facts (fixed point)

    // Phase 2: Conflict resolution
    //   - Multiple rules may conclude same thing
    //   - Use certainty + weight to resolve

    // Phase 3: Causal graph construction
    //   - Build DAG of conclusions
    //   - Edge if one conclusion's facts are prerequisites for another

    // Phase 4: Root cause identification
    //   - Source nodes in causal DAG with highest severity
    //   - Path from root cause to symptoms

    // Phase 5: Remediation mapping
    //   - Map conclusions to remediation steps
    //   - Order by priority and risk
    //   - Mark auto-fixable remediations
}
```

### 17.5 Knowledge Base Format (YAML)

```yaml
# builtin/cpu.yaml
version: "1.0"
domain: cpu
description: "CPU health knowledge base"

facts:
  - name: cpu.cores
    type: metric
    source: /proc/cpuinfo

  - name: cpu.usage_percent
    type: metric
    source: /proc/stat
    unit: percent

  - name: cpu.load_1m
    type: metric
    source: /proc/loadavg
    unit: load

  - name: cpu.temperature
    type: metric
    source: sensors
    unit: celsius

  - name: cpu.governor
    type: state
    source: /sys/devices/system/cpu/cpu0/cpufreq/scaling_governor

  - name: cpu.throttling
    type: state
    source: /sys/devices/system/cpu/cpu0/thermal_throttle/core_throttle_count

rules:
  - id: cpu.overheated
    name: "CPU Overheating"
    severity: critical
    conditions:
      all:
        - fact: cpu.temperature
          operator: gt
          value: 85
        - fact: cpu.throttling
          operator: gt
          value: 0
    conclusions:
      - id: cpu.overheated.conclusion
        title: "CPU is overheating and throttling"
        description: "CPU temperature exceeds 85°C with active thermal throttling"
        certainty: 0.95
        remediations:
          - step: 1
            action: "Check CPU cooler mounting and thermal paste"
            command: "sensors && dmesg | grep -i thermal"
            risk: low
          - step: 2
            action: "Check system fan operation"
            command: "pwmconfig && fancontrol"
            risk: low
          - step: 3
            action: "Reduce CPU frequency to prevent damage"
            command: "cpupower frequency-set --governor powersave"
            risk: medium
            auto_fixable: true

  - id: cpu.high_temp_no_throttle
    name: "High CPU Temperature (Pre-throttle)"
    severity: warning
    conditions:
      all:
        - fact: cpu.temperature
          operator: gt
          value: 75
        - fact: cpu.throttling
          operator: eq
          value: 0
    conclusions:
      - id: cpu.high_temp_no_throttle.conclusion
        title: "CPU temperature is elevated"
        description: "CPU temperature exceeds 75°C but has not yet begun throttling"
        certainty: 0.80
        remediations:
          - step: 1
            action: "Improve system cooling before throttling begins"
            command: "sensors"
            risk: low

  - id: cpu.load_anomaly
    name: "Abnormal CPU Load Pattern"
    severity: warning
    conditions:
      all:
        - fact: cpu.load_1m
          operator: gt
          value: "baseline * 2"    # Reference to baseline
        - fact: cpu.load_5m
          operator: lt
          fact_ref: cpu.load_1m    # Recent load higher than 5m avg
    conclusions:
      - id: cpu.load_anomaly.conclusion
        title: "Recent CPU load spike detected"
        description: "1-minute load average is more than double the baseline and rising"
        certainty: 0.75
        remediations:
          - step: 1
            action: "Identify top CPU-consuming processes"
            command: "ps aux --sort=-%cpu | head -20"
            risk: low
          - step: 2
            action: "Check system logs for recent activity"
            command: "journalctl --since '5 minutes ago' --no-pager | tail -50"
            risk: low
```

### 17.6 Meta-Knowledge

The engine includes meta-rules that improve diagnostic quality:

```yaml
# meta-rules about the diagnostic process itself
rules:
  - id: meta.insufficient_data
    name: "Insufficient Data for Diagnosis"
    severity: info
    conditions:
      all:
        - fact: _meta.checks_executed
          operator: lt
          value: 3
        - fact: _meta.missing_baseline
          operator: eq
          value: true
    conclusions:
      - id: meta.insufficient_data.conclusion
        title: "More data needed for accurate diagnosis"
        description: "Run with --all flag and consider capturing a baseline first"
        certainty: 1.0
```

### 17.7 Hot-Reload

The knowledge engine supports hot-reloading YAML rule files:

```go
// Watch directory for changes and reload knowledge base
// Uses fsnotify for file system events
func (e *Engine) Watch(ctx context.Context, dir string) error {
    watcher, _ := fsnotify.NewWatcher()
    defer watcher.Close()

    for {
        select {
        case event := <-watcher.Events:
            if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
                e.LoadKnowledgeBaseFile(event.Name)
                e.logger.Info("Knowledge base reloaded", "file", event.Name)
            }
        case err := <-watcher.Errors:
            return err
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}
```

---

## 18. Dependency Selection

### 18.1 Core Dependencies

| Package | Purpose | Justification | Version |
|---|---|---|---|
| `github.com/spf13/cobra` | CLI framework | De facto standard, subcommand support, autocomplete | v1.8+ |
| `github.com/spf13/viper` | Configuration | YAML/env/flag unification, 12-factor support | v1.18+ |
| `github.com/charmbracelet/bubbletea` | TUI framework | Elm-architecture, excellent Go support, active maintainer | v0.25+ |
| `github.com/charmbracelet/lipgloss` | TUI styling | CSS-like styling for terminal, composable | v0.10+ |
| `github.com/charmbracelet/bubbles` | TUI components | Table, spinner, progress, textinput, help, viewport | v0.18+ |
| `github.com/charmbracelet/glamour` | Markdown rendering | For report previews in terminal | v0.6+ |
| `modernc.org/sqlite` | SQL database | Pure Go SQLite (no CGo), embeddable, ACID | v1.28+ |
| `go.etcd.io/bbolt` | Key-value store | Embedded, time-series optimized, battle-tested in etcd | v1.3+ |
| `golang.org/x/crypto` | SSH for fleet | Standard Go SSH library, no external deps | Latest |
| `gopkg.in/yaml.v3` | YAML parsing | De facto standard, well-maintained | v3.0+ |
| `github.com/robfig/cron/v3` | Cron scheduling | For scheduled baselines, monitoring | v3.0+ |
| `github.com/shirou/gopsutil/v3` | System metrics | Cross-platform system metrics, well-maintained | v3.24+ |
| `github.com/jaypipes/ghw` | Hardware info | Detailed hardware introspection | v0.12+ |
| `github.com/olekukonez/tablewriter` | Terminal tables | Rich table formatting with alignment, colors | v0.0+ |
| `github.com/go-git/go-git/v5` | Git operations | For knowledge base versioning (future) | v5.11+ |

### 18.2 Why These Dependencies

| Decision | Alternative | Rationale |
|---|---|---|
| `modernc.org/sqlite` over `mattn/go-sqlite3` | CGo-based SQLite | Pure Go = no C toolchain requirement, cross-compilation friendly |
| `go.etcd.io/bbolt` over Badger | Badger is more complex | BoltDB simplicity, proven in etcd, good for time-series patterns |
| `shirou/gopsutil` over direct `/proc` parsing | Direct parsing is fragile | gopsutil handles edge cases, cross-platform, well-tested |
| `jaypipes/ghw` over `dmidecode` parsing | ghw is pure Go | No external command dependency, structured output |
| No logging framework (std log/slog) | logrus, zap, zerolog | Go 1.21+ `slog` is sufficient, reduces dependency surface |

### 18.3 Dependency Management

- Dependencies pinned to minor versions (e.g., `v1.28.x`)
- `go.mod` committed to repository
- `go.sum` for integrity verification
- Regular `go mod tidy` in CI
- Dependabot configured for automated updates
- Vendoring considered for enterprise air-gapped deployments

### 18.4 Zero External Dependencies (Philosophical)

LHD has **zero runtime external dependencies**:
- No system package requirements beyond a standard Linux userspace
- No database server, no runtime, no interpreter
- Runs on minimal/docker/embedded Linux systems
- Single static binary with `CGO_ENABLED=0`

---

## 19. Development Roadmap

### Phase 1: Foundation (Weeks 1-4)

```
Week 1: Project skeleton
├── Go module initialization
├── Repository structure
├── Makefile, CI pipeline
├── Cobra CLI skeleton (all commands)
├── Core types and interfaces
└── Configuration system (Viper)

Week 2: Plugin system + Distribution layer
├── Plugin registry and execution engine
├── Distro detection and interface
├── Debian, RHEL, Arch implementations
├── Topological sort for check ordering
└── Parallel execution with semaphore

Week 3: CPU + Memory + Disk checks
├── CPU: usage, load, temperature, governor
├── Memory: usage, swap, OOM, pressure
├── Disk: usage, inodes, I/O, S.M.A.R.T
├── Output: terminal table + color
├── Progress indicators
└── Unit tests for all checks

Week 4: Network + Kernel + Services checks
├── Network: connectivity, DNS, latency, ports
├── Kernel: version, sysctl, dmesg, entropy
├── Services: systemd units, socket activation
├── Integration tests
├── CLI flag processing
└── Beta: lhd doctor working end-to-end
```

### Phase 2: Intelligence (Weeks 5-8)

```
Week 5: Knowledge engine
├── Fact management
├── Rule engine (forward chaining)
├── YAML knowledge base loading
├── Built-in rules (CPU, memory, disk)
├── Conclusion generation
└── Unit tests for inference engine

Week 6: Root cause analysis
├── Causal chain construction
├── Root cause identification
├── Severity scoring
├── Remediation mapping
├── lhd diagnose command
└── Integration tests with baseline

Week 7: Baseline engine
├── Baseline profile definition
├── Baseline capture
├── Baseline storage (SQLite)
├── Baseline comparison
├── Deviation analysis
├── lhd baseline command
└── lhd doctor --baseline comparison

Week 8: Snapshot engine
├── Snapshot capture (manifest + data)
├── Compression (gzip)
├── Snapshot storage
├── Snapshot diff engine
├── lhd snapshot command
└── Snapshot + baseline interoperability
```

### Phase 3: Monitoring & Reporting (Weeks 9-12)

```
Week 9: Storage layer
├── SQLite store implementation
├── BoltDB metrics store
├── Repository pattern
├── Database migrations
├── Data retention policies
└── Performance benchmarking

Week 10: Monitoring engine
├── Continuous data collection
├── Sampler and collector
├── Threshold evaluation
├── Alert generation
├── Desktop notifications
├── lhd monitor command
└── Daemon mode

Week 11: Reporting engine
├── Multi-format output (JSON, YAML, CSV)
├── Markdown and HTML templates
├── PDF generation (via templates)
├── Template engine (customizable)
├── Report comparison feature
├── lhd report command
└── Health scoring

Week 12: Testing + Documentation
├── Comprehensive test suite
├── Integration tests
├── Documentation site
├── Man pages
├── Shell completions
└── Example configurations
```

### Phase 4: Dashboard & Fleet (Weeks 13-16)

```
Week 13: TUI Dashboard (Part 1)
├── Bubble Tea model structure
├── Tab navigation
├── Overview tab (health gauge, stats)
├── Check tab (result table)
├── Color scheme and styling
└── Keyboard bindings

Week 14: TUI Dashboard (Part 2)
├── Monitor tab (sparkline charts)
├── Logs tab (journal viewer)
├── Settings tab
├── Real-time refresh
├── Help modal
└── lhd dashboard command

Week 15: Fleet management
├── SSH connection management
├── Inventory YAML format
├── Parallel fleet execution
├── Result aggregation
├── Fleet reporting
├── lhd fleet command
└── Security — known hosts, key management

Week 16: Security checks + Hardening
├── Full security check suite
├── CVE scanning
├── Custom check sandboxing
├── Secure config handling
├── Penetration testing
├── Security review
└── Supply chain security setup
```

### Phase 5: Polish & Release (Weeks 17-20)

```
Week 17: Performance optimization
├── Check execution profiling
├── Memory optimization
├── Startup time reduction
├── Large fleet benchmarking
├── SQLite query optimization
└── BoltDB compaction

Week 18: Edge cases + Error handling
├── Graceful degradation on missing data
├── Permission error handling
├── Network timeout handling
├── Signal handling (SIGINT, SIGTERM)
├── Race condition testing
└── Fuzz testing

Week 19: User experience
├── Progressive disclosure (simple by default)
├── Better error messages
├── Suggestion mode (what to do next)
├── Color-blind friendly palette
├── Accessibility review
└── User testing feedback

Week 20: Release preparation
├── GoReleaser configuration
├── Release signing (Cosign)
├── SBOM generation
├── Homebrew formula
├── Docker image
├── AUR package
├── Nix package
└── v1.0.0 release
```

---

## 20. Release Roadmap

### Version Scheme

Follows [Semantic Versioning](https://semver.org/):
- **Major**: Breaking CLI/subcommand changes, data format changes, storage schema changes
- **Minor**: New features, new checks, new output formats, backward-compatible additions
- **Patch**: Bug fixes, performance improvements, security patches

### Release Cadence

| Version | Timeline | Focus |
|---|---|---|
| **v0.1.0-alpha** | Week 4 | Core CLI skeleton, config, plugin system, CPU/memory/disk checks — `lhd doctor` works |
| **v0.2.0-alpha** | Week 6 | All built-in checks, distro abstraction, basic output formatting |
| **v0.3.0-alpha** | Week 8 | Knowledge engine, RCA, `lhd diagnose` command with root cause identification |
| **v0.4.0-alpha** | Week 9 | Baseline engine — `lhd baseline capture/compare/list` |
| **v0.5.0-alpha** | Week 10 | Snapshot engine — `lhd snapshot create/diff/list` |
| **v0.6.0-beta** | Week 11 | Storage layer, monitoring engine — `lhd monitor start/stop/status` |
| **v0.7.0-beta** | Week 12 | Reporting — `lhd report generate` with JSON/YAML/Markdown/HTML |
| **v0.8.0-beta** | Week 14 | TUI Dashboard — `lhd dashboard` with full interactive interface |
| **v0.9.0-beta** | Week 15 | Fleet management — `lhd fleet` (SSH-based multi-host diagnostics) |
| **v0.10.0-beta** | Week 16 | Security checks, CVE scanning, release hardening, pentest fixes |
| **v1.0.0-rc.1** | Week 18 | Feature freeze, comprehensive testing, documentation complete |
| **v1.0.0-rc.2** | Week 19 | Bug fixes from rc.1, performance optimization, UX polish |
| **v1.0.0** | Week 20 | **First stable release** — Production-ready, fully documented |
| **v1.1.0** | Week 24 | Custom check improvements, more knowledge base rules, NixOS support |
| **v1.2.0** | Quarter 2 | Plugin SDK for external Go plugins (gRPC-based), more fleet features |
| **v2.0.0** | Future | Possible architecture evolution based on community feedback |

### Release Artifacts

| Artifact | Platform | Method |
|---|---|---|
| Linux amd64 (static binary) | All | GoReleaser |
| Linux arm64 (static binary) | Raspberry Pi, ARM servers | GoReleaser |
| Linux 386 (static binary) | Legacy systems | GoReleaser |
| Homebrew formula | macOS/Linux | Homebrew tap |
| Docker image | Docker Hub, GHCR | Docker multi-arch |
| AUR package | Arch Linux | AUR |
| Nix package | NixOS | Nixpkgs |
| DEB package | Debian/Ubuntu | GoReleaser (nfpm) |
| RPM package | RHEL/Fedora | GoReleaser (nfpm) |
| APK package | Alpine | GoReleaser (nfpm) |
| Source tarball | All | GitHub releases |
| SBOM | All | GoReleaser (sBOM) |
| Signed checksums | All | GoReleaser (cosign) |

### Distribution Channels

```yaml
# goreleaser.yaml (structure)
version: 2
project_name: lhd
builds:
  - env: [CGO_ENABLED=0]
    goos: [linux]
    goarch: [amd64, arm64, "386"]
    ldflags:
      - -s -w -X github.com/linuxhealthdoctor/lhd/pkg/version.Version={{.Version}}
archives:
  - format: tar.gz
    files: [LICENSE, README.md]
checksum:
  name_template: 'checksums.txt'
snapshot:
  name_template: "{{ .Version }}-next"
changelog:
  sort: asc
nfpms:
  - package_name: lhd
    homepage: https://github.com/linuxhealthdoctor/lhd
    maintainer: Linux Health Doctor Team
    description: Linux system health diagnostics and root-cause analysis
    license: MIT
    formats: [deb, rpm, apk]
brews:
  - tap: linuxhealthdoctor/homebrew-tap
    folder: Formula
dockers:
  - image_templates: ["ghcr.io/linuxhealthdoctor/lhd:{{ .Version }}"]
sboms:
  - artifacts: archive
signs:
  - artifacts: all
    args: ["--output", "${signature}", "dist/${artifact}"]
```

### Long-Term Vision

```
v1.0 ─── Foundation: Complete single-host diagnostics
    │
    ├── v1.x ─── Enhancement: More checks, rules, formats
    │   ├── More built-in knowledge base rules
    │   ├── Custom check SDK improvements
    │   ├── Community-contributed check packages
    │   └── Performance profiling integration
    │
    ├── v2.x ─── Scale: Fleet and enterprise features
    │   ├── Centralized fleet management server
    │   ├── RBAC for fleet operations
    │   ├── Compliance framework (CIS, SOC2, HIPAA mappings)
    │   ├── Alertmanager/PagerDuty integration
    │   └── Webhook notifications
    │
    └── v3.x ─── Intelligence: Advanced diagnostics
        ├── Anomaly detection (statistical, not ML)
        ├── Predictive failure analysis
        ├── Capacity planning recommendations
        ├── Automated remediation (opt-in)
        └── Integration with Ansible, Puppet, Salt
```

---

## Appendix A: Error Handling Strategy

```go
// Domain error types
var (
    ErrCheckFailed     = errors.New("check execution failed")
    ErrCheckTimeout    = errors.New("check timed out")
    ErrPermission      = errors.New("insufficient permissions")
    ErrConfigNotFound  = errors.New("configuration not found")
    ErrInvalidConfig   = errors.New("invalid configuration")
    ErrNoBaseline      = errors.New("no baseline found")
    ErrNoSnapshot      = errors.New("no snapshot found")
    ErrFleetConnect    = errors.New("fleet connection failed")
    ErrFleetTimeout    = errors.New("fleet operation timed out")
    ErrPluginPanic     = errors.New("plugin panicked")
)

// Error wrapping pattern
func (c *CPUSCheck) Check(ctx context.Context, req *plugin.CheckRequest) (*plugin.CheckResult, error) {
    result := &plugin.CheckResult{
        ID:        c.ID(),
        Timestamp: time.Now(),
    }

    value, err := readCPULoad()
    if err != nil {
        result.Status = core.StatusError
        result.Error = fmt.Errorf("%w: reading CPU load: %w", ErrCheckFailed, err)
        return result, nil  // Return result with error, don't propagate
    }

    // ... check logic ...
    return result, nil
}
```

## Appendix B: Testing Strategy

| Level | Scope | Tools |
|---|---|---|
| **Unit tests** | Individual check logic, rule evaluation, baseline comparison | `go test`, table-driven tests |
| **Integration tests** | Plugin system with real distro detection, actual /proc reading | `go test` with build tags, Docker-based test fixtures |
| **Snapshot tests** | Output formatting, report generation, TUI rendering | `go test`, golden files |
| **Fleet tests** | SSH connection, parallel execution, result aggregation | SSH server mock (gossh), Docker compose |
| **End-to-end** | Full command pipeline from CLI to output | `go test`, testscript |
| **Fuzz tests** | Config parsing, rule condition evaluation, YAML loading | `go test -fuzz` |
| **Performance** | Check execution speed, memory usage, fleet scale | `go test -bench`, pprof |
| **Security** | Custom check sandboxing, privilege escalation paths | `go test`, custom security harness |

## Appendix C: Data Flow Diagrams

### Doctor Command Data Flow

```
User: lhd doctor
  │
  ▼
CLI (Cobra) parses flags
  │ --category, --severity, --format, --output, --fix
  ▼
Config Loader (Viper)
  │ Merges: defaults → config file → env vars → CLI flags
  ▼
Distro Detection
  │ Identify: Ubuntu 24.04, systemd, apt
  ▼
Plugin Registry
  │ Filter: requested categories, enabled checks
  │ Order: dependency-based topological sort
  ▼
Execution Pipeline
  │ Layer 1: independent checks (parallel, N goroutines)
  │ Layer 2: dependent checks (after deps complete)
  │ Progress: spinner/progress bar
  ▼
Result Aggregation
  │ Collect: all CheckResults
  │ Score: CalculateHealthScore
  ▼
Output Formatter
  │ Format: ANSI terminal / JSON / YAML / etc.
  │ Save: to file if --output specified
  ▼
Terminal Output / File
```

### Diagnose Command Data Flow

```
User: lhd diagnose
  │
  ▼
Doctor Pipeline (above)
  │ Produces: []*CheckResult
  ▼
Fact Extraction
  │ CheckResult → []Fact
  │ Baseline → []Fact (if baseline specified)
  ▼
Knowledge Engine
  │ Load: builtin + custom rules
  │ Infer: forward-chaining rule evaluation
  │ Chain: causal dependency graph
  │ Identify: root cause node
  ▼
Analysis Result
  │ Overall severity, root cause, causal chain
  │ Conclusions with evidence, certainty scores
  │ Remediations ordered by priority
  ▼
Output / Report
  │ Terminal: color-coded diagnosis
  │ Report: full report with RCA
  │ TUI: dashboard highlights
```

---

## License

MIT License — Copyright (c) 2024 Linux Health Doctor Contributors

```
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
