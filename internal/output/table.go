package output

import (
	"fmt"
	"strings"
)

type Table struct {
	Headers []string
	Rows    [][]string
	Widths  []int
}

func NewTable(headers []string) *Table {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	return &Table{
		Headers: headers,
		Widths:  widths,
	}
}

func (t *Table) AddRow(row []string) {
	for i, cell := range row {
		if i < len(t.Widths) && len(cell) > t.Widths[i] {
			t.Widths[i] = len(cell)
		}
	}
	t.Rows = append(t.Rows, row)
}

func (t *Table) Render() string {
	var b strings.Builder
	sep := t.renderSeparator()

	b.WriteString(sep)
	b.WriteString(t.renderRow(t.Headers))
	b.WriteString(sep)

	for _, row := range t.Rows {
		b.WriteString(t.renderRow(row))
	}
	b.WriteString(sep)

	return b.String()
}

func (t *Table) renderRow(row []string) string {
	cells := make([]string, len(t.Headers))
	for i, v := range row {
		if i < len(t.Widths) {
			cells[i] = fmt.Sprintf(" %-*s ", t.Widths[i], v)
		} else {
			cells[i] = " " + v + " "
		}
	}
	return "|" + strings.Join(cells, "|") + "|\n"
}

func (t *Table) renderSeparator() string {
	cells := make([]string, len(t.Widths))
	for i, w := range t.Widths {
		cells[i] = "-" + strings.Repeat("-", w) + "-"
	}
	return "+" + strings.Join(cells, "+") + "+\n"
}
