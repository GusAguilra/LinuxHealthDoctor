package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Gauge struct {
	Percent float64
	Width   int
	FilledStyle lipgloss.Style
	EmptyStyle  lipgloss.Style
	ShowLabel bool
	Label     string
}

func NewGauge(percent float64, width int) Gauge {
	return Gauge{
		Percent: percent,
		Width:   width,
		FilledStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#43BF6D")),
		EmptyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#383838")),
		ShowLabel: true,
	}
}

func (g Gauge) Render() string {
	if g.Width <= 0 {
		g.Width = 40
	}

	filled := int(g.Percent * float64(g.Width) / 100.0)
	if filled > g.Width {
		filled = g.Width
	}
	if filled < 0 {
		filled = 0
	}
	empty := g.Width - filled

	filledColor := g.FilledStyle
	emptyColor := g.EmptyStyle

	var color lipgloss.Color
	switch {
	case g.Percent >= 90:
		color = "#E84A4A"
	case g.Percent >= 70:
		color = "#E8B730"
	default:
		color = "#43BF6D"
	}
	filledColor = filledColor.Foreground(color)

	bar := filledColor.Render(strings.Repeat("█", filled)) +
		emptyColor.Render(strings.Repeat("░", empty))

	if g.ShowLabel {
		label := g.Label
		if label == "" {
			label = fmt.Sprintf("%.1f%%", g.Percent)
		}
		return fmt.Sprintf("%s %s", bar, lipgloss.NewStyle().Foreground(color).Render(label))
	}

	return bar
}
