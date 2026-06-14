package components

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Sparkline struct {
	Data     []float64
	Width    int
	BarStyle lipgloss.Style
}

func NewSparkline(data []float64, width int) Sparkline {
	return Sparkline{
		Data:  data,
		Width: width,
		BarStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7B8CFF")),
	}
}

func (s *Sparkline) Render() string {
	if len(s.Data) == 0 || s.Width <= 0 {
		return ""
	}

	data := s.Data
	if len(data) > s.Width {
		data = data[len(data)-s.Width:]
	}

	min, max := math.Inf(1), math.Inf(-1)
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	if max == min {
		max = min + 1
	}

	range_ := max - min
	chars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

	var b strings.Builder
	for _, v := range data {
		normalized := (v - min) / range_
		idx := int(normalized * float64(len(chars)-1))
		if idx >= len(chars) {
			idx = len(chars) - 1
		}
		if idx < 0 {
			idx = 0
		}
		b.WriteString(s.BarStyle.Render(chars[idx]))
	}

	return b.String()
}

type BarChart struct {
	Items      []BarItem
	Width      int
	MaxHeight  int
	ShowLabels bool
}

type BarItem struct {
	Label string
	Value float64
	Color string
}

func NewBarChart(width, maxHeight int) BarChart {
	return BarChart{
		Width:      width,
		MaxHeight:  maxHeight,
		ShowLabels: true,
	}
}

func (bc *BarChart) AddItem(label string, value float64, color string) {
	bc.Items = append(bc.Items, BarItem{Label: label, Value: value, Color: color})
}

func (bc *BarChart) Render() string {
	if len(bc.Items) == 0 {
		return ""
	}

	maxVal := 0.0
	for _, item := range bc.Items {
		if item.Value > maxVal {
			maxVal = item.Value
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	var b strings.Builder
	numBars := len(bc.Items)
	barWidth := bc.Width / numBars
	if barWidth < 2 {
		barWidth = 2
	}
	barChar := "█"

	for h := bc.MaxHeight; h > 0; h-- {
		threshold := float64(h) / float64(bc.MaxHeight) * maxVal
		for _, item := range bc.Items {
			style := lipgloss.NewStyle()
			if item.Color != "" {
				style = style.Foreground(lipgloss.Color(item.Color))
			}
			if item.Value >= threshold {
				b.WriteString(style.Render(strings.Repeat(barChar, barWidth)))
			} else {
				b.WriteString(strings.Repeat(" ", barWidth))
			}
		}
		b.WriteString("\n")
	}

	if bc.ShowLabels {
		for _, item := range bc.Items {
			label := item.Label
			if len(label) > barWidth {
				label = label[:barWidth]
			}
			b.WriteString(fmt.Sprintf("%-*s", barWidth, label))
		}
		b.WriteString("\n")
	}

	return b.String()
}
