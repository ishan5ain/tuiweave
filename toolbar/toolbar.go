// Package toolbar provides a one-line, focusable strip of application-owned
// actions. It renders horizontal action choices and emits a SelectedMsg when
// the focused choice is activated.
package toolbar

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/inspect"
)

const (
	// ActionFocus focuses the toolbar.
	ActionFocus = "focus"
	// ActionBlur blurs the toolbar.
	ActionBlur = "blur"
	// ActionNext selects the next enabled action.
	ActionNext = "next"
	// ActionPrevious selects the previous enabled action.
	ActionPrevious = "previous"
	// ActionFirst selects the first enabled action.
	ActionFirst = "first"
	// ActionLast selects the last enabled action.
	ActionLast = "last"
	// ActionActivate activates the selected action.
	ActionActivate = "activate"
	// ActionSelectPrefix prefixes stable action-selection IDs.
	ActionSelectPrefix = "select."
)

// Item is one toolbar action. Disabled items remain visible but cannot be
// selected or activated.
type Item struct {
	ID          string
	Label       string
	Description string
	Disabled    bool
}

// SelectedMsg is emitted when a toolbar item is activated with enter or the
// ActionActivate semantic action.
type SelectedMsg struct {
	ID    string
	Index int
	Label string
}

// Model is a one-line focusable toolbar. Create one with New.
type Model struct {
	width, height int
	items         []Item
	selected      int
	off           int
	focused       bool

	barStyle             lipgloss.Style
	itemStyle            lipgloss.Style
	disabledStyle        lipgloss.Style
	selectedStyle        lipgloss.Style
	selectedBlurredStyle lipgloss.Style
	separatorStyle       lipgloss.Style
}

// New returns an empty toolbar styled from the theme's roles.
func New(theme gotui.Theme) Model {
	bar := lipgloss.NewStyle().Background(theme.SurfaceRaised)
	return Model{
		selected: -1,
		barStyle: bar,
		itemStyle: bar.
			Foreground(theme.Text),
		disabledStyle: bar.
			Foreground(theme.TextFaint),
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

// SetSize sets the toolbar's box. The toolbar renders one row whenever height
// is nonzero; height is otherwise ignored like other one-line bars.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
	m.scrollIntoView()
}

// SetItems replaces the actions, preserving the selected ID when it remains
// enabled. Otherwise the first enabled action is selected.
func (m *Model) SetItems(items ...Item) {
	selectedID := m.SelectedID()
	m.items = append([]Item(nil), items...)
	m.selected = -1
	if selectedID != "" {
		for i, item := range m.items {
			if item.ID == selectedID && !item.Disabled {
				m.selected = i
				break
			}
		}
	}
	if m.selected < 0 {
		m.selected = m.firstEnabled()
	}
	m.scrollIntoView()
}

// Items returns the configured toolbar actions.
func (m Model) Items() []Item { return append([]Item(nil), m.items...) }

// Select selects an enabled action by index. Disabled or out-of-range indices
// leave the current selection unchanged.
func (m *Model) Select(index int) {
	if index < 0 || index >= len(m.items) || m.items[index].Disabled {
		return
	}
	m.selected = index
	m.scrollIntoView()
}

// SelectID selects an enabled action by stable ID.
func (m *Model) SelectID(id string) {
	for i, item := range m.items {
		if item.ID == id && !item.Disabled {
			m.Select(i)
			return
		}
	}
}

// Selected returns the selected action index, or -1 when no enabled action
// exists.
func (m Model) Selected() int { return m.selected }

// SelectedItem returns the selected action, or a zero Item when there is no
// selection.
func (m Model) SelectedItem() Item {
	if m.selected < 0 || m.selected >= len(m.items) {
		return Item{}
	}
	return m.items[m.selected]
}

// SelectedID returns the selected action's stable ID, or an empty string when
// there is no selection or the action has no ID.
func (m Model) SelectedID() string { return m.SelectedItem().ID }

// Focus makes the toolbar respond to horizontal navigation and activation.
func (m *Model) Focus() { m.focused = true }

// Blur stops the toolbar from responding to keys.
func (m *Model) Blur() { m.focused = false }

// Focused reports whether the toolbar handles keys.
func (m Model) Focused() bool { return m.focused }

// Activate returns a command that emits SelectedMsg for the current enabled
// selection, or nil when the toolbar has no selection.
func (m Model) Activate() tea.Cmd {
	item := m.SelectedItem()
	if m.selected < 0 || item.Disabled {
		return nil
	}
	return func() tea.Msg {
		return SelectedMsg{ID: item.ID, Index: m.selected, Label: item.Label}
	}
}

// Update handles semantic actions and, while focused, left/right arrows,
// h/l, home/g, end/G, and enter activation.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if next, cmd, handled := m.applyAction(msg); handled {
		return next, cmd
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
		m.selectRelative(-1)
	case "right", "l":
		m.selectRelative(1)
	case "home", "g":
		m.Select(m.firstEnabled())
	case "end", "G":
		m.Select(m.lastEnabled())
	case "enter":
		return m, m.Activate()
	}
	return m, nil
}

// View renders one exact-width row. Whole actions are kept when possible;
// when the selected action alone is too wide, its label is truncated.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	if len(m.items) == 0 {
		return m.barStyle.Render(strings.Repeat(" ", m.width))
	}

	parts := make([]string, 0, len(m.items)-m.off)
	used := 0
	for i := m.off; i < len(m.items); i++ {
		item := m.renderItem(i)
		separator := 0
		if len(parts) > 0 {
			separator = 1
		}
		if used+separator+lipgloss.Width(item) > m.width {
			break
		}
		if separator > 0 {
			parts = append(parts, m.separatorStyle.Render(" "))
			used++
		}
		parts = append(parts, item)
		used += lipgloss.Width(item)
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
		label := ansi.Truncate(prefix+m.items[selected].Label+" ", m.width, "…")
		parts = append(parts, m.styleFor(selected).Render(label))
		used = lipgloss.Width(parts[0])
	}
	if used < m.width {
		parts = append(parts, m.barStyle.Render(strings.Repeat(" ", m.width-used)))
	}
	return strings.Join(parts, "")
}

func (m Model) renderItem(index int) string {
	prefix := " "
	if index == m.selected {
		prefix = "▸"
	}
	return m.styleFor(index).Render(prefix + m.items[index].Label + " ")
}

func (m Model) styleFor(index int) lipgloss.Style {
	item := m.items[index]
	if item.Disabled {
		return m.disabledStyle
	}
	if index == m.selected {
		if m.focused {
			return m.selectedStyle
		}
		return m.selectedBlurredStyle
	}
	return m.itemStyle
}

func (m Model) itemWidth(index int) int {
	return lipgloss.Width(" " + m.items[index].Label + " ")
}

func (m Model) firstEnabled() int {
	for i, item := range m.items {
		if !item.Disabled {
			return i
		}
	}
	return -1
}

func (m Model) lastEnabled() int {
	for i := len(m.items) - 1; i >= 0; i-- {
		if !m.items[i].Disabled {
			return i
		}
	}
	return -1
}

func (m *Model) selectRelative(delta int) {
	if len(m.items) == 0 {
		return
	}
	if m.selected < 0 {
		if delta < 0 {
			m.Select(m.lastEnabled())
		} else {
			m.Select(m.firstEnabled())
		}
		return
	}
	step := 1
	if delta < 0 {
		step = -1
	}
	index := m.selected
	for {
		index += step
		if index < 0 || index >= len(m.items) {
			return
		}
		if !m.items[index].Disabled {
			m.Select(index)
			return
		}
	}
}

func (m *Model) scrollIntoView() {
	if len(m.items) == 0 {
		m.off = 0
		return
	}
	if m.selected < 0 {
		m.off = 0
		return
	}
	if m.selected < m.off {
		m.off = m.selected
	}
	used := 0
	for i := m.off; i <= m.selected; i++ {
		if i > m.off {
			used++
		}
		used += m.itemWidth(i)
	}
	if used > m.width {
		m.off = m.selected
	}
	m.off = max(0, min(m.off, len(m.items)-1))
}

func (m Model) applyAction(msg tea.Msg) (Model, tea.Cmd, bool) {
	action, ok := msg.(inspect.ActionMsg)
	if !ok {
		return m, nil, false
	}
	switch action.ID {
	case ActionFocus:
		m.Focus()
	case ActionBlur:
		m.Blur()
	case ActionNext:
		m.selectRelative(1)
	case ActionPrevious:
		m.selectRelative(-1)
	case ActionFirst:
		m.Select(m.firstEnabled())
	case ActionLast:
		m.Select(m.lastEnabled())
	case ActionActivate:
		return m, m.Activate(), true
	default:
		if !strings.HasPrefix(action.ID, ActionSelectPrefix) {
			return m, nil, false
		}
		id := strings.TrimPrefix(action.ID, ActionSelectPrefix)
		for i, item := range m.items {
			if item.ID == id && !item.Disabled {
				m.Select(i)
				return m, nil, true
			}
		}
		return m, nil, false
	}
	return m, nil, true
}
