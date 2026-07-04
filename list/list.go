// Package list provides a scrolling list of single-line items with a
// selection cursor and optional filtering.
//
// Filtering is display-only state: SetFilter narrows what is shown and
// navigated, but Selected always reports the index into the original items,
// so app logic never deals with filtered indices.
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
	filter        string
	matches       []int // original indices matching filter; nil when no filter
	pos           int   // cursor position in display space
	off           int   // first visible display position
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

// SetItems replaces the items, reapplying any active filter and clamping the
// selection.
func (m *Model) SetItems(items ...string) {
	m.items = items
	m.applyFilter(m.Selected())
}

// Items returns the original, unfiltered items.
func (m Model) Items() []string { return m.items }

// Len returns the number of original items.
func (m Model) Len() int { return len(m.items) }

// SetFilter narrows the list to items containing query, case-insensitively.
// An empty query clears the filter. The current selection is kept when it
// still matches; otherwise the first match is selected.
func (m *Model) SetFilter(query string) {
	keep := m.Selected()
	m.filter = query
	m.applyFilter(keep)
}

// Filter returns the active filter query.
func (m Model) Filter() string { return m.filter }

// FilteredLen returns how many items are currently displayed.
func (m Model) FilteredLen() int { return m.displayCount() }

// applyFilter rebuilds the match set, trying to keep the item with original
// index keep selected.
func (m *Model) applyFilter(keep int) {
	if m.filter == "" {
		m.matches = nil
	} else {
		q := strings.ToLower(m.filter)
		m.matches = make([]int, 0, len(m.items))
		for i, item := range m.items {
			if strings.Contains(strings.ToLower(item), q) {
				m.matches = append(m.matches, i)
			}
		}
	}
	m.pos = 0
	for p := range m.displayCount() {
		if m.origIndex(p) == keep {
			m.pos = p
			break
		}
	}
	m.clampPos()
	m.scrollIntoView()
}

// displayCount is how many items are shown under the current filter.
func (m Model) displayCount() int {
	if m.matches == nil {
		return len(m.items)
	}
	return len(m.matches)
}

// origIndex maps a display position to an original item index.
func (m Model) origIndex(pos int) int {
	if m.matches == nil {
		return pos
	}
	return m.matches[pos]
}

// Select moves the selection to the item with original index i. Without a
// filter the index is clamped; with a filter active, non-matching indices
// leave the selection unchanged.
func (m *Model) Select(i int) {
	if m.matches == nil {
		m.pos = max(0, min(i, len(m.items)-1))
		m.scrollIntoView()
		return
	}
	for p, orig := range m.matches {
		if orig == i {
			m.pos = p
			m.scrollIntoView()
			return
		}
	}
}

func (m *Model) selectPos(p int) {
	m.pos = max(0, min(p, m.displayCount()-1))
	m.clampPos()
	m.scrollIntoView()
}

func (m *Model) clampPos() {
	m.pos = max(0, min(m.pos, max(0, m.displayCount()-1)))
}

// Selected returns the original index of the selected item, or -1 when
// nothing is displayed.
func (m Model) Selected() int {
	if m.displayCount() == 0 {
		return -1
	}
	return m.origIndex(m.pos)
}

// SelectedItem returns the selected item, or "" when nothing is displayed.
func (m Model) SelectedItem() string {
	if m.displayCount() == 0 {
		return ""
	}
	return m.items[m.origIndex(m.pos)]
}

// Focus makes the list respond to navigation keys and highlights the
// selection with the selection roles.
func (m *Model) Focus() { m.focused = true }

// Blur stops the list from responding to keys.
func (m *Model) Blur() { m.focused = false }

// Focused reports whether the list handles keys.
func (m Model) Focused() bool { return m.focused }

// TotalLines returns the displayed item count (scrollbar.Scrollable).
func (m Model) TotalLines() int { return m.displayCount() }

// VisibleLines returns how many rows are shown at once (scrollbar.Scrollable).
func (m Model) VisibleLines() int { return m.height }

// YOffset returns the first visible display position (scrollbar.Scrollable).
func (m Model) YOffset() int { return m.off }

func (m *Model) scrollIntoView() {
	if m.height <= 0 {
		return
	}
	if m.pos < m.off {
		m.off = m.pos
	}
	if m.pos >= m.off+m.height {
		m.off = m.pos - m.height + 1
	}
	m.off = max(0, min(m.off, max(0, m.displayCount()-m.height)))
}

// Update handles navigation while focused: up/k, down/j, pgup, pgdown,
// g/home, G/end — all within the displayed (filtered) items.
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
		m.selectPos(m.pos - 1)
	case "down", "j":
		m.selectPos(m.pos + 1)
	case "pgup":
		m.selectPos(m.pos - m.height)
	case "pgdown":
		m.selectPos(m.pos + m.height)
	case "g", "home":
		m.selectPos(0)
	case "G", "end":
		m.selectPos(m.displayCount() - 1)
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
		p := m.off + i
		if p >= m.displayCount() {
			rows[i] = strings.Repeat(" ", m.width)
			continue
		}
		rows[i] = m.renderItem(p)
	}
	return strings.Join(rows, "\n")
}

func (m Model) renderItem(pos int) string {
	text := ansi.Truncate(m.items[m.origIndex(pos)], m.width-2, "…")
	pad := strings.Repeat(" ", max(0, m.width-2-ansi.StringWidth(text)))

	if pos == m.pos && m.focused {
		return m.markerStyle.Render("▌") + m.selectedStyle.Render(" "+text+pad)
	}
	marker := " "
	if pos == m.pos {
		marker = "▎"
	}
	return m.itemStyle.Render(marker+" "+text) + pad
}
