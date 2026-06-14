package dashboard

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ready = true
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.viewport.Width = m.width - 2
		m.viewport.Height = m.height - 6
		if m.components != nil {
			m.viewport.SetContent(m.buildTabContent())
		} else {
			m.viewport.SetContent("  Running checks...")
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.viewport.LineUp(1)
			return m, nil
		case "down", "j":
			m.viewport.LineDown(1)
			return m, nil
		case "pgup", "b":
			m.viewport.HalfViewUp()
			return m, nil
		case "pgdown", "f":
			m.viewport.HalfViewDown()
			return m, nil
		default:
			return m.handleKeyPress(msg)
		}

	case tickMsg:
		return m, func() tea.Msg {
			return tickMsg(time.Now())
		}

	case spinner.TickMsg:
		if m.components != nil {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case ResultsMsg:
		m.SetResults(msg.Result)
		return m, nil

	default:
		return m, nil
	}
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.tick.Stop()
		return m, tea.Quit

	case "tab", "right":
		tabs := m.tabs()
		m.activeTab = tabs[(int(m.activeTab)+1)%len(tabs)]
		m.viewport.GotoTop()
		if m.components != nil {
			m.viewport.SetContent(m.buildTabContent())
		}

	case "shift+tab", "left":
		tabs := m.tabs()
		m.activeTab = tabs[(int(m.activeTab)-1+len(tabs))%len(tabs)]
		m.viewport.GotoTop()
		if m.components != nil {
			m.viewport.SetContent(m.buildTabContent())
		}

	case "1":
		m.activeTab = Overview
		m.viewport.GotoTop()
		if m.components != nil {
			m.viewport.SetContent(m.buildTabContent())
		}
	case "2":
		m.activeTab = Checks
		m.viewport.GotoTop()
		if m.components != nil {
			m.viewport.SetContent(m.buildTabContent())
		}
	case "3":
		m.activeTab = Monitor
		m.viewport.GotoTop()
		if m.components != nil {
			m.viewport.SetContent(m.buildTabContent())
		}
	case "4":
		m.activeTab = Logs
		m.viewport.GotoTop()
		if m.components != nil {
			m.viewport.SetContent(m.buildTabContent())
		}
	case "5":
		m.activeTab = About
		m.viewport.GotoTop()
		if m.components != nil {
			m.viewport.SetContent(m.buildTabContent())
		}

	case "r":
		return m, func() tea.Msg {
			return tickMsg(time.Now())
		}
	}

	return m, nil
}
