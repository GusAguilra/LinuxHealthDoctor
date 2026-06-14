package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type StatusBar struct {
	LeftText   string
	RightText  string
	CenterText string
	Time       time.Time
	ShowTime   bool
	Style      lipgloss.Style
}

func NewStatusBar() StatusBar {
	return StatusBar{
		ShowTime: true,
		Style: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0A0A0")).
			Background(lipgloss.Color("#1A1A1A")).
			Padding(0, 1).
			Width(80),
	}
}

func (s *StatusBar) Render(width int) string {
	style := s.Style.Width(width)

	right := s.RightText
	if s.ShowTime {
		right = time.Now().Format("15:04:05")
		if s.RightText != "" {
			right = s.RightText + " | " + right
		}
	}

	var center string
	if s.CenterText != "" {
		center = s.CenterText
	}

	leftWidth := width - lipgloss.Width(right)
	if center != "" {
		leftWidth -= lipgloss.Width(center)
	}
	if leftWidth < 0 {
		leftWidth = 0
	}

	left := s.LeftText
	if len(left) > leftWidth {
		left = left[:leftWidth]
	}

	leftPadded := fmt.Sprintf("%-*s", leftWidth, left)

	var result string
	if center != "" {
		result = leftPadded + center + right
	} else {
		result = strings.TrimRight(leftPadded, " ") + right
	}

	return style.Render(result)
}
