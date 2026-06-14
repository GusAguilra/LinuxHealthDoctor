package dashboard

import (
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	Header   lipgloss.Style
	Tab      lipgloss.Style
	TabActive lipgloss.Style
	Content  lipgloss.Style
	Footer   lipgloss.Style
	Gauge    lipgloss.Style
	Success  lipgloss.Style
	Warning  lipgloss.Style
	Error    lipgloss.Style
	Info     lipgloss.Style
	Muted    lipgloss.Style
	Border   lipgloss.Style
	Help     lipgloss.Style
	Title    lipgloss.Style
}

func DefaultStyles() Styles {
	subtle := lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight := lipgloss.AdaptiveColor{Light: "#2E42E5", Dark: "#7B8CFF"}
	special := lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	return Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#2E42E5")).
			Padding(0, 2).
			MarginBottom(1),

		Tab: lipgloss.NewStyle().
			Padding(0, 3).
			Foreground(lipgloss.Color("#A0A0A0")).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(subtle),

		TabActive: lipgloss.NewStyle().
			Padding(0, 3).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(highlight).
			Bold(true),

		Content: lipgloss.NewStyle().
			Padding(1, 2),

		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0A0A0")).
			Background(lipgloss.Color("#1A1A1A")).
			Padding(0, 1),

		Gauge: lipgloss.NewStyle().
			Foreground(special).
			Background(subtle),

		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#43BF6D")),

		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E8B730")),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E84A4A")),

		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7B8CFF")),

		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B6B6B")),

		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(subtle).
			Padding(1, 2),

		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B6B6B")),

		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Padding(0, 1),
	}
}
