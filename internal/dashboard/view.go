package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if !m.ready {
		return m.renderLoading()
	}

	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteRune('\n')
	b.WriteString(m.renderTabs())
	b.WriteRune('\n')
	b.WriteString(m.renderScrollableView())
	b.WriteRune('\n')
	b.WriteString(m.renderFooter())
	b.WriteRune('\n')

	return b.String()
}

func (m Model) renderLoading() string {
	return m.styles.Header.Render("Linux Health Doctor") + "\n\n" +
		m.spinner.View() + " Initializing...\n"
}

func (m Model) renderHeader() string {
	col := "#43BF6D"
	status := "Excellent"
	switch {
	case m.healthScore >= 90:
	case m.healthScore >= 70:
		col = "#E8B730"
		status = "Good"
	case m.healthScore >= 50:
		col = "#E8B730"
		status = "Fair"
	case m.healthScore >= 30:
		col = "#E84A4A"
		status = "Poor"
	default:
		col = "#E84A4A"
		status = "Critical"
	}
	score := lipgloss.NewStyle().Foreground(lipgloss.Color(col)).Bold(true).Render(fmt.Sprintf("%.0f%%", m.healthScore))
	return m.styles.Header.Width(m.width).Render(fmt.Sprintf("Linux Health Doctor  %s  %s", score, status))
}

func (m Model) renderTabs() string {
	var parts []string
	for _, t := range m.tabs() {
		if t == m.activeTab {
			parts = append(parts, m.styles.TabActive.Render(t.String()))
		} else {
			parts = append(parts, m.styles.Tab.Render(t.String()))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func (m Model) renderScrollableView() string {
	view := m.viewport.View()
	if m.components == nil {
		return view
	}

	vpLines := strings.Split(strings.TrimRight(view, "\n"), "\n")
	visibleH := len(vpLines)
	if visibleH == 0 {
		return view
	}

	totalContent := m.buildTabContent()
	totalLines := len(strings.Split(strings.TrimRight(totalContent, "\n"), "\n"))
	if totalLines <= visibleH {
		return view
	}

	pct := m.viewport.ScrollPercent()
	barH := visibleH
	thumbH := max(1, barH*visibleH/totalLines)
	thumbPos := int(pct * float64(barH-thumbH))

	scrollStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B8CFF"))
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#383838"))

	contentW := m.width - 1
	out := make([]string, visibleH)
	for i := 0; i < visibleH; i++ {
		line := vpLines[i]
		w := lipgloss.Width(line)
		if w < contentW {
			line += strings.Repeat(" ", contentW-w)
		} else if w > contentW {
			line = lipgloss.NewStyle().Width(contentW).Render(line)
		}
		if i >= thumbPos && i < thumbPos+thumbH {
			line += scrollStyle.Render("▌")
		} else {
			line += trackStyle.Render("▌")
		}
		out[i] = line
	}
	return strings.Join(out, "\n")
}

var componentLabels = map[string]string{
	"cpu":      "CPU",
	"memory":   "Memory",
	"disk":     "Disk",
	"network":  "Network",
	"kernel":   "Kernel",
	"services": "Services",
	"security": "Security",
	"logs":     "Logs",
}

// ---------- Overview ----------

func (m Model) renderOverview() string {
	var b strings.Builder

	b.WriteString(m.renderGauge(m.healthScore))
	b.WriteString("\n\n")

	status := "Excellent"
	switch {
	case m.healthScore >= 90:
		status = "Excellent"
	case m.healthScore >= 70:
		status = "Good"
	case m.healthScore >= 50:
		status = "Fair"
	case m.healthScore >= 30:
		status = "Poor"
	default:
		status = "Critical"
	}
	col := "#43BF6D"
	if m.healthScore >= 70 {
	} else if m.healthScore >= 30 {
		col = "#E8B730"
	} else {
		col = "#E84A4A"
	}
	b.WriteString(fmt.Sprintf("  Overall Status: %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color(col)).Render(status)))

	if m.components == nil {
		b.WriteString(fmt.Sprintf("\n  %s Running checks...\n", m.spinner.View()))
		return b.String()
	}

	total, passed, failed, errored := 0, 0, 0, 0
	for _, s := range m.components {
		total += s.Total
		passed += s.Passed
		failed += s.Failed
		errored += s.Errors
	}

	b.WriteString("\n")
	b.WriteString(m.styles.Success.Render(fmt.Sprintf("  ✓ %d Passed", passed)))
	if failed+errored > 0 {
		b.WriteString("  ")
		b.WriteString(m.styles.Error.Render(fmt.Sprintf("✗ %d Failed/Errors", failed+errored)))
	}
	b.WriteString(fmt.Sprintf("\n  %d total checks\n\n", total))

	// Only show components that have issues
	hasIssues := false
	for _, s := range m.components {
		if s.Failed > 0 || s.Errors > 0 {
			hasIssues = true
			break
		}
	}
	if !hasIssues {
		b.WriteString(m.styles.Success.Render("  All components healthy"))
		b.WriteString("\n")
		return b.String()
	}

	b.WriteString(m.styles.Info.Render("  Issues"))
	b.WriteString("\n")
	for _, s := range m.components {
		if s.Failed == 0 && s.Errors == 0 {
			continue
		}
		label := componentLabels[string(s.Component)]
		if label == "" {
			label = string(s.Component)
		}
		b.WriteString(m.styles.Error.Render(fmt.Sprintf("  ✗ %s", label)))
		b.WriteString(fmt.Sprintf("  — %d/%d\n", s.Passed, s.Total))
		for _, d := range s.Details {
			if d.Status == "pass" {
				continue
			}
			sev := d.Severity
			if sev == "" {
				sev = "warning"
			}
			var sevStyle lipgloss.Style
			switch sev {
			case "critical", "fatal":
				sevStyle = m.styles.Error
			case "error":
				sevStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
			case "warning":
				sevStyle = m.styles.Warning
			default:
				sevStyle = m.styles.Info
			}
			msg := d.Message
			if msg == "" {
				msg = "Unknown error"
			}
			b.WriteString(fmt.Sprintf("    %s %s\n", sevStyle.Render(fmt.Sprintf("[%s]", sev)), msg))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderGauge(score float64) string {
	w := m.width - 10
	if w < 10 {
		w = 10
	}
	if w > 80 {
		w = 80
	}
	filled := int(score * float64(w) / 100.0)
	empty := w - filled

	col := "#43BF6D"
	switch {
	case score >= 90:
	case score >= 70:
		col = "#E8B730"
	case score >= 30:
		col = "#E84A4A"
	default:
		col = "#E84A4A"
	}

	bar := lipgloss.NewStyle().Foreground(lipgloss.Color(col)).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#383838")).Render(strings.Repeat("░", empty))

	return fmt.Sprintf("  Health: %s %.1f%%", bar, score)
}

// ---------- Checks ----------

func (m Model) renderChecks() string {
	var b strings.Builder
	if m.components == nil {
		b.WriteString(fmt.Sprintf("\n  %s Running checks...\n", m.spinner.View()))
		return b.String()
	}

	total, passed, failed, errored := 0, 0, 0, 0
	for _, s := range m.components {
		total += s.Total
		passed += s.Passed
		failed += s.Failed
		errored += s.Errors
	}
	b.WriteString(m.styles.Info.Render(fmt.Sprintf("  Results: %d total  ✓ %d passed  ✗ %d failed  %d errors", total, passed, failed, errored)))
	b.WriteString("\n\n")

	for _, s := range m.components {
		label := componentLabels[string(s.Component)]
		if label == "" {
			label = string(s.Component)
		}

		if s.Failed > 0 || s.Errors > 0 {
			b.WriteString(m.styles.Error.Render(fmt.Sprintf("  ✗ %s", label)))
		} else {
			b.WriteString(m.styles.Success.Render(fmt.Sprintf("  ✓ %s", label)))
		}
		b.WriteString(fmt.Sprintf("  — %d/%d\n", s.Passed, s.Total))

		for _, d := range s.Details {
			switch d.Status {
			case "pass":
				b.WriteString(fmt.Sprintf("    %s %s\n",
					m.styles.Success.Render("✓"),
					d.Message,
				))
			case "fail":
				sev := d.Severity
				if sev == "" {
					sev = "warning"
				}
				var sevStyle lipgloss.Style
				switch sev {
				case "critical", "fatal":
					sevStyle = m.styles.Error
				case "error":
					sevStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
				case "warning":
					sevStyle = m.styles.Warning
				default:
					sevStyle = m.styles.Info
				}
				b.WriteString(fmt.Sprintf("    %s %s\n",
					sevStyle.Render(fmt.Sprintf("[%s]", sev)),
					d.Message,
				))
		case "error":
			errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
			b.WriteString(fmt.Sprintf("    %s %s\n",
				errStyle.Render("[error]"),
				d.Message,
			))
			default:
				b.WriteString(fmt.Sprintf("    %s %s\n", m.styles.Muted.Render("~"), d.Message))
			}

			// Extra detail from ResultDetails
			extra := formatCheckDetails(d.ID, d.ResultDetails)
			for _, line := range extra {
				b.WriteString(fmt.Sprintf("      %s\n", m.styles.Muted.Render(line)))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

func formatCheckDetails(id string, details map[string]interface{}) []string {
	if details == nil {
		return nil
	}
	switch id {
	case "security.ports":
		return formatPortsDetail(details)
	case "security.suid":
		return formatSUIDDetail(details)
	case "security.ssh":
		return formatSSHDetail(details)
	case "disk.usage":
		return formatDiskUsageDetail(details)
	case "disk.io":
		return formatDiskIODetail(details)
	case "services.failed":
		return formatServiceFailedDetail(details)
	case "services.critical":
		return formatServiceCriticalDetail(details)
	case "network.interfaces":
		return formatNetworkInterfacesDetail(details)
	case "kernel.sysctl":
		return formatSysctlDetail(details)
	case "kernel.dmesg":
		return formatDmesgDetail(details)
	case "cpu.load":
		return formatCPULoadDetail(details)
	case "memory.usage":
		return formatMemoryUsageDetail(details)
	case "memory.swap":
		return formatMemorySwapDetail(details)
	case "memory.oom":
		return formatMemoryOOMDetail(details)
	case "logs.journal":
		return formatJournalDetail(details)
	case "kernel.version":
		return formatKernelVersionDetail(details)
	case "network.connectivity":
		return formatConnectivityDetail(details)
	case "network.dns":
		return formatDNSDetail(details)
	case "cpu.usage":
		return formatCPUUsageDetail(details)
	}
	return nil
}


func formatPortsDetail(d map[string]interface{}) []string {
	rawPorts, _ := d["ports"].([]map[string]string)
	limit := rawPorts
	if len(limit) > 6 {
		limit = limit[:6]
	}
	var lines []string
	lines = append(lines, fmt.Sprintf("Ports: %v listening", d["listening_ports"]))
	for _, pm := range limit {
		addr := pm["address"]
		proc := pm["process"]
		if proc != "" {
			lines = append(lines, fmt.Sprintf("  %s  (%s)", addr, proc))
		} else {
			lines = append(lines, fmt.Sprintf("  %s", addr))
		}
	}
	return lines
}

func formatSUIDDetail(d map[string]interface{}) []string {
	files, _ := d["suid_files"].([]string)
	limit := files
	if len(limit) > 5 {
		limit = limit[:5]
	}
	var lines []string
	lines = append(lines, fmt.Sprintf("Count: %v", d["suid_count"]))
	for _, f := range limit {
		lines = append(lines, fmt.Sprintf("  %s", f))
	}
	if len(files) > 5 {
		lines = append(lines, fmt.Sprintf("  ... and %d more", len(files)-5))
	}
	return lines
}

func formatSSHDetail(d map[string]interface{}) []string {
	issues, _ := d["issues"].([]string)
	if len(issues) == 0 {
		return nil
	}
	var lines []string
	for _, iss := range issues {
		lines = append(lines, fmt.Sprintf("  %s", iss))
	}
	return lines
}

func formatDiskUsageDetail(d map[string]interface{}) []string {
	parts, _ := d["partitions"].([]map[string]interface{})
	var lines []string
	for _, pm := range parts {
		mp, _ := pm["mountpoint"].(string)
		usage, _ := pm["usage_percent"].(string)
		total, _ := pm["total_gb"].(string)
		free, _ := pm["free_gb"].(string)
		lines = append(lines, fmt.Sprintf("  %s  %s used  (%s / %s free)", mp, usage, free, total))
	}
	return lines
}

func formatDiskIODetail(d map[string]interface{}) []string {
	devs, _ := d["devices"].([]map[string]interface{})
	limit := devs
	if len(limit) > 5 {
		limit = limit[:5]
	}
	var lines []string
	for _, dm := range limit {
		name, _ := dm["device"].(string)
		r, _ := dm["read_count"].(uint64)
		w, _ := dm["write_count"].(uint64)
		lines = append(lines, fmt.Sprintf("  %s: %d reads, %d writes", name, r, w))
	}
	return lines
}

func formatServiceFailedDetail(d map[string]interface{}) []string {
	units, _ := d["failed_units"].([]string)
	if len(units) == 0 {
		return nil
	}
	var lines []string
	for _, u := range units {
		lines = append(lines, fmt.Sprintf("  %s", u))
	}
	return lines
}

func formatServiceCriticalDetail(d map[string]interface{}) []string {
	running, _ := d["running"].(int)
	stopped, _ := d["stopped"].([]string)
	var lines []string
	if len(stopped) > 0 {
		lines = append(lines, fmt.Sprintf("Stopped services:"))
		for _, s := range stopped {
			lines = append(lines, fmt.Sprintf("  %s", s))
		}
	}
	lines = append(lines, fmt.Sprintf("Running: %d", running))
	return lines
}

func formatNetworkInterfacesDetail(d map[string]interface{}) []string {
	ifaces, _ := d["interfaces"].([]map[string]interface{})
	var lines []string
	lines = append(lines, fmt.Sprintf("Interfaces: %d total", len(ifaces)))
	limit := ifaces
	if len(limit) > 5 {
		limit = limit[:5]
	}
	for _, im := range limit {
		name, _ := im["name"].(string)
		eIn, _ := im["errors_in"].(uint64)
		eOut, _ := im["errors_out"].(uint64)
		line := fmt.Sprintf("  %s", name)
		if eIn+eOut > 0 {
			line += fmt.Sprintf(" (%d errors)", eIn+eOut)
		}
		lines = append(lines, line)
	}
	return lines
}

func formatSysctlDetail(d map[string]interface{}) []string {
	anomalies, _ := d["anomalies"].([]string)
	if len(anomalies) == 0 {
		return nil
	}
	var lines []string
	for _, a := range anomalies {
		lines = append(lines, fmt.Sprintf("  %s", a))
	}
	return lines
}

func formatDmesgDetail(d map[string]interface{}) []string {
	samples, _ := d["log_sample"].([]string)
	if len(samples) == 0 {
		return nil
	}
	var lines []string
	for _, s := range samples {
		lines = append(lines, fmt.Sprintf("  %s", s))
	}
	return lines
}

func formatCPULoadDetail(d map[string]interface{}) []string {
	return []string{
		fmt.Sprintf("  Load: 1m=%v 5m=%v 15m=%v (cores: %v)",
			d["load_1m"], d["load_5m"], d["load_15m"], d["cpu_cores"]),
	}
}

func formatCPUUsageDetail(d map[string]interface{}) []string {
	u, _ := d["usage_percent"].(string)
	if u == "" {
		return nil
	}
	return []string{fmt.Sprintf("  Usage: %s", u)}
}

func formatMemoryUsageDetail(d map[string]interface{}) []string {
	return []string{
		fmt.Sprintf("  Total: %v  Used: %v  Free: %v", d["total"], d["used"], d["free"]),
	}
}

func formatMemorySwapDetail(d map[string]interface{}) []string {
	return []string{
		fmt.Sprintf("  Total: %v  Used: %v", d["total"], d["used"]),
	}
}

func formatMemoryOOMDetail(d map[string]interface{}) []string {
	return []string{
		fmt.Sprintf("  Swap total: %v  Swap cached: %v", d["swap_total"], d["swap_cached"]),
	}
}

func formatJournalDetail(d map[string]interface{}) []string {
	du, _ := d["disk_usage"].(string)
	if du == "" {
		return nil
	}
	return []string{fmt.Sprintf("  %s", du)}
}

func formatKernelVersionDetail(d map[string]interface{}) []string {
	v, _ := d["kernel_version"].(string)
	if v == "" {
		return nil
	}
	return []string{fmt.Sprintf("  Version: %s", v)}
}

func formatConnectivityDetail(d map[string]interface{}) []string {
	t, _ := d["targets_tested"].(int)
	r, _ := d["targets_reachable"].(int)
	return []string{fmt.Sprintf("  %d/%d targets reachable", r, t)}
}

func formatDNSDetail(d map[string]interface{}) []string {
	r, _ := d["hosts_resolved"].(int)
	hosts, _ := d["hosts_tested"].([]string)
	return []string{fmt.Sprintf("  %d/%d hosts resolved", r, len(hosts))}
}

// ---------- Monitor ----------

func (m Model) renderMonitor() string {
	var b strings.Builder
	if len(m.metrics) == 0 {
		b.WriteString("\n  Waiting for metrics...\n")
		return b.String()
	}

	type metricInfo struct {
		Label       string
		Description string
	}
	keyNames := map[string]metricInfo{
		"usage_percent":        {"CPU Usage", "Current CPU utilization across all cores"},
		"memory_usage_percent": {"Memory Usage", "RAM currently in use"},
		"swap_usage_percent":   {"Swap Usage", "Swap space currently in use"},
		"load_1m":              {"Load (1m)", "System load average over 1 minute"},
		"load_5m":              {"Load (5m)", "System load average over 5 minutes"},
	}

	b.WriteString(m.styles.Info.Render("  Live Metrics"))
	b.WriteString("\n\n")
	for name, info := range keyNames {
		if vals, ok := m.metrics[name]; ok && len(vals) > 0 {
			last := vals[len(vals)-1]
			col := "#43BF6D"
			threshold := ""
			if last > 80 {
				col = "#E8B730"
				threshold = " (high)"
			}
			if last > 95 {
				col = "#E84A4A"
				threshold = " (critical)"
			}
			val := lipgloss.NewStyle().Foreground(lipgloss.Color(col)).Render(fmt.Sprintf("%.1f%%", last))
			b.WriteString(fmt.Sprintf("  %-20s %s%s\n", info.Label+":", val, threshold))
			b.WriteString(fmt.Sprintf("  %-20s %s\n", "", m.styles.Muted.Render(info.Description)))
		}
	}
	b.WriteString(fmt.Sprintf("\n  %s", m.styles.Muted.Render("Values snapshot from last check run")))

	return b.String()
}

// ---------- About ----------

func (m Model) renderAbout() string {
	var b strings.Builder

	b.WriteString(m.styles.Info.Render("  Linux Health Doctor"))
	b.WriteString("\n\n")
	b.WriteString("  A system health monitoring and diagnostic tool for Linux.\n")
	b.WriteString("  It runs a battery of checks across 8 components and reports\n")
	b.WriteString("  the overall health of the system.\n\n")

	b.WriteString(m.styles.Info.Render("  Components"))
	b.WriteString("\n\n")

	type compInfo struct {
		label string
		count int
	}
	comps := []compInfo{
		{"CPU", 4},
		{"Memory", 3},
		{"Disk", 2},
		{"Network", 3},
		{"Kernel", 3},
		{"Services", 2},
		{"Security", 3},
		{"Logs", 2},
	}
	for _, c := range comps {
		b.WriteString(fmt.Sprintf("  • %s: %d checks\n", c.label, c.count))
	}

	b.WriteString(fmt.Sprintf("\n  Total: %d checks across %d categories\n\n", 22, len(comps)))

	b.WriteString(m.styles.Info.Render("  Check Reference"))
	b.WriteString("\n\n")

	type checkRef struct {
		id      string
		name    string
		desc    string
		source  string
		details string
	}
	type group struct {
		label  string
		checks []checkRef
	}
	groups := []group{
		{
			"CPU", []checkRef{
				{"cpu.usage", "CPU Usage", "Current CPU utilization percentage",
					"gopsutil/cpu.Percent", "Alert: >80% warning, >95% critical"},
				{"cpu.load", "CPU Load Average", "System load averages (1m, 5m, 15m)",
					"gopsutil/load.Avg", "Compared against CPU core count"},
				{"cpu.temperature", "CPU Temperature", "Thermal status of CPU cores",
					"gopsutil/host.SensorsTemperature", "Alert: >80°C warning, >95°C critical"},
				{"cpu.governor", "CPU Scaling Governor", "Frequency scaling governor in use",
					"sysfs /sys/devices/system/cpu/*/cpufreq/scaling_governor", "Verifies governor matches expected policy"},
			},
		},
		{
			"Memory", []checkRef{
				{"memory.usage", "Memory Usage", "RAM usage percentage and totals",
					"gopsutil/mem.VirtualMemory", "Alert: >80% warning, >95% critical"},
				{"memory.swap", "Swap Usage", "Swap space usage percentage",
					"gopsutil/mem.SwapMemory", "Alert: >50% warning, >80% critical"},
				{"memory.oom", "OOM Status", "Recent OOM killer events",
					"journalctl -q -u systemd-oomd / dmesg", "Scans for OOM-related kernel messages"},
			},
		},
		{
			"Disk", []checkRef{
				{"disk.usage", "Disk Usage", "Partition usage percentages",
					"gopsutil/disk.Usage (skips loop/squashfs/tmpfs)", "Alert: >85% warning, >95% critical"},
				{"disk.io", "Disk I/O", "Disk read/write statistics",
					"gopsutil/disk.IOCounters", "Monitors read/write counts and throughput per device"},
			},
		},
		{
			"Network", []checkRef{
				{"network.connectivity", "Network Connectivity", "Ping reachability of configured targets",
					"net.Dial / ICMP echo", "Reports reachable/total targets"},
				{"network.interfaces", "Network Interfaces", "Interface error and drop counters",
					"gopsutil/net.IOCounters", "Flags interfaces with non-zero error or drop counts"},
				{"network.dns", "DNS Resolution", "DNS lookup success rate for configured hosts",
					"net.LookupHost", "Reports resolved/total hosts"},
			},
		},
		{
			"Kernel", []checkRef{
				{"kernel.version", "Kernel Version", "Kernel version meets minimum requirements",
					"uname -r", "Compares against baseline.min_kernel config"},
				{"kernel.dmesg", "Kernel Dmesg", "Scan kernel ring buffer for error/warn messages",
					"dmesg --level=err,warn (fallback: dmesg)", "Flags critical keywords: error, panic, bug, oops, failed"},
				{"kernel.sysctl", "Sysctl Anomalies", "Check for insecure sysctl parameter values",
					"sysctl -n <param>", "Expected: ip_forward=0, rp_filter=1, dmesg_restrict=1"},
			},
		},
		{
			"Services", []checkRef{
				{"services.failed", "Failed Services", "Failed systemd units",
					"systemctl list-units --state=failed", "Lists all units in failed state"},
				{"services.critical", "Critical Services", "Critical system services status",
					"systemctl is-active <unit>", "Verifies must-be-running services are active"},
			},
		},
		{
			"Security", []checkRef{
				{"security.ports", "Open Ports", "Listening ports and associated services",
					"ss -tlnp", "Lists unexpected or unauthorized listening services"},
				{"security.suid", "SUID Files", "SUID binary inventory",
					"find / -perm -4000", "Reports SUID binaries outside expected set"},
				{"security.ssh", "SSH Configuration", "SSH daemon security hardening check",
					"sshd -T", "Verifies PermitRootLogin, PasswordAuthentication, PubkeyAuthentication"},
			},
		},
		{
			"Logs", []checkRef{
				{"logs.errors", "Recent Log Errors", "Journal error entries in the last hour",
					"journalctl --since '1 hour ago' -p err", "Alert: >10 warning, >50 critical"},
				{"logs.journal", "Journal Health", "Systemd journal disk usage and accessibility",
					"journalctl --disk-usage / --list-boots", "Verifies journal is readable and reports disk consumption"},
			},
		},
	}

	divider := m.styles.Muted.Render(strings.Repeat("─", 40))

	for _, g := range groups {
		b.WriteString(fmt.Sprintf("  %s\n\n", m.styles.TabActive.Render(g.label)))
		for _, ch := range g.checks {
			b.WriteString(fmt.Sprintf("  %s\n", m.styles.Title.Render(ch.id)))
			b.WriteString(fmt.Sprintf("    %s  ", m.styles.Muted.Render("Name:")))
			b.WriteString(fmt.Sprintf("%s\n", ch.name))
			b.WriteString(fmt.Sprintf("    %s  ", m.styles.Muted.Render("What:")))
			b.WriteString(fmt.Sprintf("%s\n", ch.desc))
			b.WriteString(fmt.Sprintf("    %s  ", m.styles.Muted.Render("Src:")))
			b.WriteString(fmt.Sprintf("%s\n", ch.source))
			b.WriteString(fmt.Sprintf("    %s  ", m.styles.Muted.Render("Rule:")))
			b.WriteString(fmt.Sprintf("%s\n", ch.details))
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("  %s\n\n", divider))
	}

	b.WriteString(m.styles.Info.Render("  Key"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  %s  Pass      %s  Warning   %s  Fail/Critical   %s  Error\n\n",
		m.styles.Success.Render("●"),
		m.styles.Warning.Render("●"),
		m.styles.Error.Render("●"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("●"),
	))

	return b.String()
}

// ---------- Logs ----------

func (m Model) renderLogs() string {
	var b strings.Builder

	if m.components == nil {
		b.WriteString("\n  Waiting for check results...\n")
		return b.String()
	}

	if len(m.alerts) == 0 {
		b.WriteString(m.styles.Success.Render("\n  No issues detected. All checks passed."))
		return b.String()
	}

	bySeverity := map[string][]Alert{}
	order := []string{"critical", "error", "warning", "info"}
	for _, a := range m.alerts {
		sev := a.Severity
		if sev == "" {
			sev = "info"
		}
		bySeverity[sev] = append(bySeverity[sev], a)
	}

	b.WriteString(m.styles.Info.Render(fmt.Sprintf("  %d Alerts from last check run", len(m.alerts))))
	b.WriteString("\n\n")
	for _, sev := range order {
		alerts, ok := bySeverity[sev]
		if !ok {
			continue
		}
		var sevStyle lipgloss.Style
		switch sev {
		case "critical", "fatal":
			sevStyle = m.styles.Error
		case "warning":
			sevStyle = m.styles.Warning
		case "error":
			sevStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		default:
			sevStyle = m.styles.Info
		}
		b.WriteString(fmt.Sprintf("  %s\n", sevStyle.Render(strings.ToUpper(sev))))
		for _, a := range alerts {
			b.WriteString(fmt.Sprintf("    %s  %s\n",
				a.Time.Format("15:04:05"),
				a.Message,
			))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m Model) renderFooter() string {
	return m.styles.Footer.Width(m.width).Render("  ←/→ tab  •  1-5 goto  •  ↑/↓ scroll  •  r refresh  •  q quit")
}
