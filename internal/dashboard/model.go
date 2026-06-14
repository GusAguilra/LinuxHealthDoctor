package dashboard

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/linuxhealthdoctor/lhd/internal/core"
)

type Tab int

const (
	Overview Tab = iota
	Checks
	Monitor
	Logs
	About
)

func (t Tab) String() string {
	switch t {
	case Overview:
		return "Overview"
	case Checks:
		return "Checks"
	case Monitor:
		return "Monitor"
	case Logs:
		return "Logs"
	case About:
		return "About"
	default:
		return "Unknown"
	}
}

type Alert struct {
	Severity string
	Message  string
	Time     time.Time
}

type CheckDetail struct {
	ID             string
	Name           string
	Status         string
	Severity       string
	Message        string
	Duration       string
	ResultDetails  map[string]interface{}
}

type tickMsg time.Time

type ResultsMsg struct {
	Result *core.AggregatedResult
}

type ComponentSummary struct {
	Component core.Component
	Total     int
	Passed    int
	Failed    int
	Errors    int
	Alerts    []Alert
	Details   []CheckDetail
}

type Model struct {
	activeTab    Tab
	ready        bool
	width        int
	height       int
	healthScore  float64
	components   []ComponentSummary
	metrics      map[string][]float64
	alerts       []Alert
	styles       Styles
	spinner      spinner.Model
	help         help.Model
	tick         *time.Ticker
	viewport     viewport.Model
	content      string
}

func (m Model) tabs() []Tab {
	return []Tab{Overview, Checks, Monitor, Logs, About}
}

func New() Model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7B8CFF"))
	s.Spinner = spinner.Dot

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle()

	return Model{
		activeTab:   Overview,
		healthScore: 100.0,
		components:  nil,
		metrics:     make(map[string][]float64),
		alerts:      make([]Alert, 0),
		styles:      DefaultStyles(),
		spinner:     s,
		help:        help.New(),
		tick:        time.NewTicker(5 * time.Second),
		viewport:    vp,
	}
}

func (m *Model) SetResults(result *core.AggregatedResult) {
	m.healthScore = result.HealthScore

	summaries := make(map[core.Component]*ComponentSummary)
	for comp, results := range result.Results {
		s := &ComponentSummary{Component: comp}

		for _, r := range results {
			s.Total++
			switch r.Status {
			case core.StatusPass:
				s.Passed++
			case core.StatusFail:
				s.Failed++
			case core.StatusError:
				s.Errors++
			}

			sev := r.Severity.String()
			if sev == "none" || sev == "" {
				if r.Status == core.StatusError {
					sev = "error"
				} else {
					sev = "info"
				}
			}

			s.Details = append(s.Details, CheckDetail{
				ID:            r.ID,
				Name:          r.Message,
				Status:        r.Status.String(),
				Severity:      sev,
				Message:       r.Message,
				Duration:      r.Duration.Truncate(time.Millisecond).String(),
				ResultDetails: r.Details,
			})

			if r.Status == core.StatusFail || r.Status == core.StatusError {
				s.Alerts = append(s.Alerts, Alert{
					Severity: sev,
					Message:  r.Message,
					Time:     r.Timestamp,
				})
			}

					for name, val := range r.Metrics {
				const maxMetrics = 100
				sl := m.metrics[name]
				if len(sl) >= maxMetrics {
					sl = sl[len(sl)-maxMetrics+1:]
				}
				m.metrics[name] = append(sl, val)
			}
		}
		summaries[comp] = s
	}

	m.components = make([]ComponentSummary, 0, len(summaries))
	m.alerts = nil
	for _, s := range summaries {
		m.components = append(m.components, *s)
		m.alerts = append(m.alerts, s.Alerts...)
	}

	m.viewport.SetContent(m.buildTabContent())
}

func (m *Model) buildTabContent() string {
	switch m.activeTab {
	case Overview:
		return m.renderOverview()
	case Checks:
		return m.renderChecks()
	case Monitor:
		return m.renderMonitor()
	case Logs:
		return m.renderLogs()
	case About:
		return m.renderAbout()
	}
	return ""
}

func (m *Model) rebuildContent() {
	m.viewport.SetContent(m.buildTabContent())
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			return tickMsg(time.Now())
		},
	)
}
