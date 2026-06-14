# Linux Health Doctor (lhd)

> Your Linux system's primary care physician.

Linux Health Doctor is a **local-first**, **offline-first**, **zero-telemetry** health diagnostics tool for Linux systems. It runs 22 checks across 8 categories and presents results in an interactive TUI dashboard with root-cause analysis and remediation guidance.

## Features

- **22 Health Checks** — CPU, Memory, Disk, Network, Kernel, Services, Security, Logs
- **Root Cause Analysis** — Rule-based knowledge engine identifies why something is wrong
- **Interactive Dashboard** — TUI with Overview, Checks, Monitor, Logs, and About tabs
- **Real-time Metrics** — Live CPU, memory, swap, and load graphs
- **Multi-Format Output** — ANSI terminal, JSON, YAML, CSV, Markdown
- **Single Binary** — Static Go binary, no runtime dependencies
- **100% Offline** — No telemetry, no tracking, no cloud

## Installation

### From source

```bash
git clone https://github.com/GusAguilra/LinuxHealthDoctor.git
cd LinuxHealthDoctor
go build -o lhd ./cmd/lhd
sudo cp lhd /usr/local/bin/
```

### Using Go

```bash
go install github.com/GusAguilra/LinuxHealthDoctor/cmd/lhd@latest
```

## Quick Start

```bash
# Run health checks and show results
sudo ./lhd doctor

# Launch the interactive dashboard
sudo ./lhd dashboard

# Run diagnostics with root cause analysis
sudo ./lhd diagnose

# Output as JSON
./lhd doctor --format json
```

## Dashboard Tabs

| Tab | Description |
|-----|-------------|
| **Overview** | Health score gauge, pass/fail summary, component issues |
| **Checks** | Per-check results with severity tags and expanded detail |
| **Monitor** | Live metric charts for CPU, memory, load averages |
| **Logs** | Alert history grouped by severity |
| **About** | App info and reference for all 22 checks |

Navigate with `←` `→` arrow keys or `1`–`5` number keys.

## Checks (22 total)

| Category | Checks | What it monitors |
|----------|--------|------------------|
| CPU | cpu.usage, cpu.load, cpu.temperature, cpu.governor | Utilization, load averages, thermal status, scaling governor |
| Memory | memory.usage, memory.swap, memory.oom | RAM usage, swap usage, OOM killer events |
| Disk | disk.usage, disk.io | Partition usage (skips loop/squashfs), I/O statistics |
| Network | network.connectivity, network.interfaces, network.dns | Ping targets, interface errors, DNS resolution |
| Kernel | kernel.version, kernel.dmesg, kernel.sysctl | Version check, ring buffer scan, sysctl hardening |
| Services | services.failed, services.critical | Failed systemd units, critical service status |
| Security | security.ports, security.suid, security.ssh | Listening ports, SUID binaries, SSH hardening |
| Logs | logs.errors, logs.journal | Recent journal errors, journal health |

## Architecture

```
CLI (Cobra) → Plugin System → Health Checks → Knowledge Engine
                ↓                              ↓
          Storage Layer                   Dashboard (TUI)
```

## Development

```bash
go build ./cmd/lhd    # Build binary
go vet ./...          # Static analysis
```

Most checks require root. Run with `sudo` for full coverage.

## License

MIT License — see [LICENSE](LICENSE) for details.
