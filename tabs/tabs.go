// Package tabs provides a one-line, focusable tab strip for navigating
// sibling views. The application owns the content shown for the selected tab;
// this component renders only the navigation chrome.
package tabs

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/action"
	"github.com/ishan5ain/tuiweave/inspect"
	"github.com/ishan5ain/tuiweave/layout"
)

const (
	// ActionFocus focuses the tab strip.
	ActionFocus = "focus"
	// ActionBlur blurs the tab strip.
	ActionBlur = "blur"
	// ActionNext selects the next tab.
	ActionNext = "next"
	// ActionPrevious selects the previous tab.
	ActionPrevious = "previous"
	// ActionFirst selects the first tab.
	ActionFirst = "first"
	// ActionLast selects the last tab.
	ActionLast = "last"
	// ActionSelectPrefix prefixes stable tab-selection actions. For example,
	// a tab with ID "logs" exposes the local action "select.logs".
	ActionSelectPrefix = action.SelectPrefix
)

// Tab is one navigable tab. ID should be stable across updates when the
// application wants to expose semantic selection actions.
type Tab struct {
	ID    string
	Label string
}

// Model is a one-line tab strip. Create one with New.
type Model struct {
	width, height int
	tabs          []Tab
	selected      int
	off           int
	focused       bool

	barStyle             lipgloss.Style
	itemStyle            lipgloss.Style
	selectedStyle        lipgloss.Style
	selectedBlurredStyle lipgloss.Style
	separatorStyle       lipgloss.Style
}

// New returns an empty tab strip styled from the theme's roles.
func New(theme tuiweave.Theme) Model {
	bar := lipgloss.NewStyle().Background(theme.SurfaceRaised)
	return Model{
		selected: -1,
		barStyle: bar,
		itemStyle: bar.
			Foreground(theme.TextMuted),
		selectedStyle: lipgloss.NewStyle().
			Foreground(theme.SelectionFg).
			Background(theme.SelectionBg).
			Bold(true),
		selectedBlurredStyle: bar.
			Foreground(theme.Accent).
			Bold(true),
		separatorStyle: bar.Foreground(theme.TextFaint),
	}
}

// SetSize sets the box the tab strip renders in. It renders one row whenever
// height is nonzero; height is otherwise ignored like other one-line bars.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
	m.scrollIntoView()
}

// SizeMode reports that the tab strip is width-constrained with a natural
// height of one row.
func (m Model) SizeMode() layout.SizeMode { return layout.SizeWidthBounded }

// SetTabs replaces the tabs, preserving the selected tab by ID when possible.
// When the old selection is unavailable, the first tab is selected; an empty
// set has no selection.
func (m *Model) SetTabs(tabs ...Tab) {
	selectedID := m.SelectedID()
	m.tabs = append([]Tab(nil), tabs...)
	m.selected = -1
	if len(m.tabs) > 0 {
		m.selected = 0
		if selectedID != "" {
			for i, tab := range m.tabs {
				if tab.ID == selectedID {
					m.selected = i
					break
				}
			}
		}
	}
	m.scrollIntoView()
}

// Tabs returns the configured tabs.
func (m Model) Tabs() []Tab { return append([]Tab(nil), m.tabs...) }

// Select selects a tab by index. Out-of-range indices clamp to the available
// set; an empty set has no selection.
func (m *Model) Select(index int) {
	if len(m.tabs) == 0 {
		m.selected = -1
		m.off = 0
		return
	}
	m.selected = max(0, min(index, len(m.tabs)-1))
	m.scrollIntoView()
}

// SelectID selects the tab with the given stable ID. It is a no-op when no tab
// has that ID.
func (m *Model) SelectID(id string) {
	for i, tab := range m.tabs {
		if tab.ID == id {
			m.Select(i)
			return
		}
	}
}

// Selected returns the selected tab index, or -1 when there are no tabs.
func (m Model) Selected() int { return m.selected }

// SelectedTab returns the selected tab, or a zero Tab when there is no
// selection.
func (m Model) SelectedTab() Tab {
	if m.selected < 0 || m.selected >= len(m.tabs) {
		return Tab{}
	}
	return m.tabs[m.selected]
}

// SelectedID returns the selected tab's stable ID, or an empty string when
// there is no selection or the selected tab has no ID.
func (m Model) SelectedID() string { return m.SelectedTab().ID }

// Focus makes the tab strip respond to navigation keys and applies focused
// selection styling.
func (m *Model) Focus() { m.focused = true }

// Blur stops the tab strip from responding to navigation keys.
func (m *Model) Blur() { m.focused = false }

// Focused reports whether the tab strip handles navigation keys.
func (m Model) Focused() bool { return m.focused }

// Update handles semantic actions and, while focused, left/right arrows,
// h/l, home/g, and end/G navigation keys.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if next, handled := m.applyAction(msg); handled {
		return next, nil
	}
	if !m.focused {
		return m, nil
	}
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "left", "h":
		m.Select(m.selected - 1)
	case "right", "l":
		m.Select(m.selected + 1)
	case "home", "g":
		m.Select(0)
	case "end", "G":
		m.Select(len(m.tabs) - 1)
	}
	return m, nil
}

// View renders the visible tab strip as one row. Whole tabs are kept when
// possible; if the selected tab alone is wider than the box, its label is
// truncated to fit. The returned row is exactly width cells wide.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	if len(m.tabs) == 0 {
		return m.barStyle.Render(strings.Repeat(" ", m.width))
	}

	parts := make([]string, 0, len(m.tabs)-m.off)
	used := 0
	for i := m.off; i < len(m.tabs); i++ {
		tab := m.renderTab(i)
		separator := 0
		if len(parts) > 0 {
			separator = 1
		}
		if used+separator+lipgloss.Width(tab) > m.width {
			break
		}
		if separator > 0 {
			parts = append(parts, m.separatorStyle.Render(" "))
			used++
		}
		parts = append(parts, tab)
		used += lipgloss.Width(tab)
	}

	if len(parts) == 0 {
		selected := m.selected
		if selected < 0 {
			selected = 0
		}
		prefix := " "
		if selected == m.selected {
			prefix = "▸"
		}
		label := ansi.Truncate(prefix+m.tabs[selected].Label+" ", m.width, "…")
		parts = append(parts, m.styleFor(selected).Render(label))
		used = lipgloss.Width(parts[0])
	}
	if used < m.width {
		parts = append(parts, m.barStyle.Render(strings.Repeat(" ", m.width-used)))
	}
	return strings.Join(parts, "")
}

func (m Model) styleFor(index int) lipgloss.Style {
	if index == m.selected {
		if m.focused {
			return m.selectedStyle
		}
		return m.selectedBlurredStyle
	}
	return m.itemStyle
}

func (m Model) renderTab(index int) string {
	prefix := " "
	if index == m.selected {
		prefix = "▸"
	}
	return m.styleFor(index).Render(prefix + m.tabs[index].Label + " ")
}

func (m Model) tabWidth(index int) int {
	return lipgloss.Width(" " + m.tabs[index].Label + " ")
}

func (m *Model) scrollIntoView() {
	if len(m.tabs) == 0 {
		m.off = 0
		return
	}
	m.selected = max(0, min(m.selected, len(m.tabs)-1))
	if m.selected < m.off {
		m.off = m.selected
	}
	used := 0
	for i := m.off; i <= m.selected; i++ {
		if i > m.off {
			used++
		}
		used += m.tabWidth(i)
	}
	if used > m.width {
		m.off = m.selected
	}
	m.off = max(0, min(m.off, len(m.tabs)-1))
}

func (m Model) applyAction(msg tea.Msg) (Model, bool) {
	event, ok := msg.(inspect.ActionMsg)
	if !ok {
		return m, false
	}
	switch event.ID {
	case ActionFocus:
		m.Focus()
	case ActionBlur:
		m.Blur()
	case ActionNext:
		m.Select(m.selected + 1)
	case ActionPrevious:
		m.Select(m.selected - 1)
	case ActionFirst:
		m.Select(0)
	case ActionLast:
		m.Select(len(m.tabs) - 1)
	default:
		id, ok := action.ParseSelectID(event.ID)
		if !ok {
			return m, false
		}
		for i, tab := range m.tabs {
			if tab.ID == id {
				m.Select(i)
				return m, true
			}
		}
		return m, false
	}
	return m, true
}
