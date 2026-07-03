// Package table provides a scrolling table with fixed and flexible columns,
// a styled header, and row selection.
package table

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
)

// Column describes one table column. Width 0 marks a flexible column:
// flexible columns share the space left over by fixed ones equally.
type Column struct {
	Title string
	Width int
}

// Model is a table component. Create one with New.
type Model struct {
	width, height int
	cols          []Column
	rows          [][]string
	sel           int
	off           int
	focused       bool

	headerStyle   lipgloss.Style
	ruleStyle     lipgloss.Style
	cellStyle     lipgloss.Style
	selectedStyle lipgloss.Style
}

// New returns an empty table styled from the theme's roles.
func New(theme gotui.Theme) Model {
	return Model{
		headerStyle: lipgloss.NewStyle().Foreground(theme.TextMuted).Bold(true),
		ruleStyle:   lipgloss.NewStyle().Foreground(theme.BorderMuted),
		cellStyle:   lipgloss.NewStyle().Foreground(theme.Text),
		selectedStyle: lipgloss.NewStyle().
			Foreground(theme.SelectionFg).
			Background(theme.SelectionBg),
	}
}

// SetSize sets the box the table renders in. Two lines go to the header and
// its rule; the rest show rows.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
	m.scrollIntoView()
}

// SetColumns replaces the column definitions.
func (m *Model) SetColumns(cols ...Column) {
	m.cols = cols
}

// SetRows replaces the rows, clamping the selection.
func (m *Model) SetRows(rows ...[]string) {
	m.rows = rows
	m.sel = max(0, min(m.sel, len(rows)-1))
	m.scrollIntoView()
}

// Select moves the selection to row i, clamped to bounds.
func (m *Model) Select(i int) {
	if len(m.rows) == 0 {
		m.sel = 0
		return
	}
	m.sel = max(0, min(i, len(m.rows)-1))
	m.scrollIntoView()
}

// Selected returns the selected row index, or -1 when empty.
func (m Model) Selected() int {
	if len(m.rows) == 0 {
		return -1
	}
	return m.sel
}

// SelectedRow returns the selected row, or nil when empty.
func (m Model) SelectedRow() []string {
	if len(m.rows) == 0 {
		return nil
	}
	return m.rows[m.sel]
}

// Focus makes the table respond to navigation keys.
func (m *Model) Focus() { m.focused = true }

// Blur stops the table from responding to keys.
func (m *Model) Blur() { m.focused = false }

// Focused reports whether the table handles keys.
func (m Model) Focused() bool { return m.focused }

// rowArea returns how many rows are visible below the header.
func (m Model) rowArea() int { return max(0, m.height-2) }

func (m *Model) scrollIntoView() {
	area := m.rowArea()
	if area == 0 {
		return
	}
	if m.sel < m.off {
		m.off = m.sel
	}
	if m.sel >= m.off+area {
		m.off = m.sel - area + 1
	}
	m.off = max(0, min(m.off, max(0, len(m.rows)-area)))
}

// Update handles navigation while focused: up/k, down/j, pgup, pgdown,
// g/home, G/end.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "up", "k":
		m.Select(m.sel - 1)
	case "down", "j":
		m.Select(m.sel + 1)
	case "pgup":
		m.Select(m.sel - m.rowArea())
	case "pgdown":
		m.Select(m.sel + m.rowArea())
	case "g", "home":
		m.Select(0)
	case "G", "end":
		m.Select(len(m.rows) - 1)
	}
	return m, nil
}

// colWidths resolves fixed and flexible column widths for the current box,
// accounting for a one-space gap between columns.
func (m Model) colWidths() []int {
	widths := make([]int, len(m.cols))
	gaps := max(0, len(m.cols)-1)
	remaining := m.width - gaps
	flex := 0
	for i, c := range m.cols {
		if c.Width > 0 {
			widths[i] = c.Width
			remaining -= c.Width
		} else {
			flex++
		}
	}
	if flex > 0 {
		share := max(1, remaining/flex)
		for i, c := range m.cols {
			if c.Width == 0 {
				widths[i] = share
			}
		}
	}
	return widths
}

func renderLine(cells []string, widths []int) string {
	parts := make([]string, len(widths))
	for i, w := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		cell = ansi.Truncate(cell, w, "…")
		parts[i] = cell + strings.Repeat(" ", max(0, w-ansi.StringWidth(cell)))
	}
	return strings.Join(parts, " ")
}

// View renders the header, a rule, and the visible rows. The selected row is
// highlighted with the selection roles while focused.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 || len(m.cols) == 0 {
		return ""
	}
	widths := m.colWidths()

	titles := make([]string, len(m.cols))
	for i, c := range m.cols {
		titles[i] = c.Title
	}
	out := make([]string, 0, m.height)
	out = append(out, m.headerStyle.Render(renderLine(titles, widths)))
	if m.height > 1 {
		out = append(out, m.ruleStyle.Render(strings.Repeat("─", m.width)))
	}

	for i := range m.rowArea() {
		idx := m.off + i
		if idx >= len(m.rows) {
			out = append(out, strings.Repeat(" ", m.width))
			continue
		}
		line := renderLine(m.rows[idx], widths)
		if idx == m.sel && m.focused {
			out = append(out, m.selectedStyle.Render(line))
		} else {
			out = append(out, m.cellStyle.Render(line))
		}
	}
	return strings.Join(out, "\n")
}
