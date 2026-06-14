package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Column struct {
	Title string
	Width int
}

type Row []string

type Table struct {
	Columns []Column
	Rows    []Row
	HeaderStyle lipgloss.Style
	RowStyle    lipgloss.Style
	AltRowStyle lipgloss.Style
	BorderStyle lipgloss.Style
	ShowBorder bool
}

func NewTable(columns []Column) Table {
	return Table{
		Columns: columns,
		HeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#2E42E5")).
			Padding(0, 1),
		RowStyle: lipgloss.NewStyle().
			Padding(0, 1),
		AltRowStyle: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#D0D0D0")),
		BorderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#383838")),
		ShowBorder: true,
	}
}

func (t *Table) AddRow(row Row) {
	t.Rows = append(t.Rows, row)
}

func (t *Table) Render() string {
	if len(t.Columns) == 0 {
		return ""
	}

	var b strings.Builder

	headerParts := make([]string, len(t.Columns))
	for i, col := range t.Columns {
		name := col.Title
		if len(name) > col.Width {
			name = name[:col.Width]
		}
		headerParts[i] = t.HeaderStyle.Width(col.Width).Render(fmt.Sprintf("%-*s", col.Width, name))
	}
	b.WriteString(strings.Join(headerParts, " "))

	if t.ShowBorder {
		sepParts := make([]string, len(t.Columns))
		for i, col := range t.Columns {
			sepParts[i] = t.BorderStyle.Render(strings.Repeat("─", col.Width))
		}
		b.WriteString("\n" + strings.Join(sepParts, " "))
	}

	for idx, row := range t.Rows {
		b.WriteString("\n")
		style := t.RowStyle
		if idx%2 == 1 {
			style = t.AltRowStyle
		}
		parts := make([]string, len(t.Columns))
		for i, col := range t.Columns {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			if len(cell) > col.Width {
				cell = cell[:col.Width]
			}
			parts[i] = style.Width(col.Width).Render(fmt.Sprintf("%-*s", col.Width, cell))
		}
		b.WriteString(strings.Join(parts, " "))
	}

	return b.String()
}
