// Package list provides a scrolling list of single-line items with a
// selection cursor.
package list

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
)

// Model is a list component. Create one with New.
type Model struct {
	width, height int
	items         []string
	sel           int
	off           int // first visible item
	focused       bool

	itemStyle     lipgloss.Style
	selectedStyle lipgloss.Style
	markerStyle   lipgloss.Style
}

// New returns an empty list styled from the theme's roles.
func New(theme gotui.Theme) Model {
	return Model{
		itemStyle: lipgloss.NewStyle().Foreground(theme.Text),
		selectedStyle: lipgloss.NewStyle().
			Foreground(theme.SelectionFg).
			Background(theme.SelectionBg),
		markerStyle: lipgloss.NewStyle().
			Foreground(theme.Accent).
			Background(theme.SelectionBg),
	}
}

// SetSize sets the box the list renders in.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
	m.scrollIntoView()
}

// SetItems replaces the items, clamping the selection.
func (m *Model) SetItems(items ...string) {
	m.items = items
	m.sel = max(0, min(m.sel, len(items)-1))
	m.scrollIntoView()
}

// Items returns the current items.
func (m Model) Items() []string { return m.items }

// Len returns the number of items.
func (m Model) Len() int { return len(m.items) }

// Select moves the selection to index i, clamped to bounds.
func (m *Model) Select(i int) {
	if len(m.items) == 0 {
		m.sel = 0
		return
	}
	m.sel = max(0, min(i, len(m.items)-1))
	m.scrollIntoView()
}

// Selected returns the index of the selected item, or -1 when empty.
func (m Model) Selected() int {
	if len(m.items) == 0 {
		return -1
	}
	return m.sel
}

// SelectedItem returns the selected item, or "" when empty.
func (m Model) SelectedItem() string {
	if len(m.items) == 0 {
		return ""
	}
	return m.items[m.sel]
}

// Focus makes the list respond to navigation keys and highlights the
// selection with the selection roles.
func (m *Model) Focus() { m.focused = true }

// Blur stops the list from responding to keys.
func (m *Model) Blur() { m.focused = false }

// Focused reports whether the list handles keys.
func (m Model) Focused() bool { return m.focused }

func (m *Model) scrollIntoView() {
	if m.height <= 0 {
		return
	}
	if m.sel < m.off {
		m.off = m.sel
	}
	if m.sel >= m.off+m.height {
		m.off = m.sel - m.height + 1
	}
	m.off = max(0, min(m.off, max(0, len(m.items)-m.height)))
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
		m.Select(m.sel - m.height)
	case "pgdown":
		m.Select(m.sel + m.height)
	case "g", "home":
		m.Select(0)
	case "G", "end":
		m.Select(len(m.items) - 1)
	}
	return m, nil
}

// View renders the visible window. The selected item gets an accent marker
// and the selection roles; when the list is blurred the marker remains as a
// subtle location hint but the row is not highlighted.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	rows := make([]string, m.height)
	for i := range rows {
		idx := m.off + i
		if idx >= len(m.items) {
			rows[i] = strings.Repeat(" ", m.width)
			continue
		}
		rows[i] = m.renderItem(idx)
	}
	return strings.Join(rows, "\n")
}

func (m Model) renderItem(idx int) string {
	text := ansi.Truncate(m.items[idx], m.width-2, "…")
	pad := strings.Repeat(" ", max(0, m.width-2-ansi.StringWidth(text)))

	if idx == m.sel && m.focused {
		return m.markerStyle.Render("▌") + m.selectedStyle.Render(" "+text+pad)
	}
	marker := " "
	if idx == m.sel {
		marker = "▎"
	}
	return m.itemStyle.Render(marker+" "+text) + pad
}
