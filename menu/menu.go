// Package menu provides a focusable vertical menu of application-owned
// actions. It renders the choices and emits a SelectedMsg on activation; the
// application owns the resulting command's behavior.
package menu

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/inspect"
)

const (
	// ActionFocus focuses the menu.
	ActionFocus = "focus"
	// ActionBlur blurs the menu.
	ActionBlur = "blur"
	// ActionNext selects the next enabled item.
	ActionNext = "next"
	// ActionPrevious selects the previous enabled item.
	ActionPrevious = "previous"
	// ActionFirst selects the first enabled item.
	ActionFirst = "first"
	// ActionLast selects the last enabled item.
	ActionLast = "last"
	// ActionActivate activates the selected item.
	ActionActivate = "activate"
	// ActionSelectPrefix prefixes stable item-selection actions. For example,
	// an item with ID "open" exposes the local action "select.open".
	ActionSelectPrefix = "select."
)

// Item is one menu action. ID should be stable across updates when the
// application wants to expose semantic selection actions and stable
// SelectedMsg values. Disabled items remain visible but cannot be selected or
// activated.
type Item struct {
	ID          string
	Label       string
	Description string
	Disabled    bool
}

// SelectedMsg is emitted when a menu item is activated with enter or the
// ActionActivate semantic action.
type SelectedMsg struct {
	ID    string
	Index int
	Label string
}

// Model is a focusable, scrolling menu. Create one with New.
type Model struct {
	width, height int
	items         []Item
	selected      int
	off           int
	focused       bool

	itemStyle            lipgloss.Style
	disabledStyle        lipgloss.Style
	selectedStyle        lipgloss.Style
	selectedBlurredStyle lipgloss.Style
}

// New returns an empty menu styled from the theme's roles.
func New(theme gotui.Theme) Model {
	return Model{
		selected:  -1,
		itemStyle: lipgloss.NewStyle().Foreground(theme.Text),
		disabledStyle: lipgloss.NewStyle().
			Foreground(theme.TextFaint),
		selectedStyle: lipgloss.NewStyle().
			Foreground(theme.SelectionFg).
			Background(theme.SelectionBg),
		selectedBlurredStyle: lipgloss.NewStyle().
			Foreground(theme.Accent),
	}
}

// SetSize sets the box the menu renders in.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
	m.scrollIntoView()
}

// SetItems replaces the menu items, preserving the selected ID when it is
// still enabled. Otherwise the first enabled item is selected.
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

// Items returns the configured menu items.
func (m Model) Items() []Item { return append([]Item(nil), m.items...) }

// Select selects an enabled item by index. Disabled or out-of-range indices
// leave the current selection unchanged.
func (m *Model) Select(index int) {
	if index < 0 || index >= len(m.items) || m.items[index].Disabled {
		return
	}
	m.selected = index
	m.scrollIntoView()
}

// SelectID selects an enabled item by stable ID. It is a no-op when no enabled
// item has that ID.
func (m *Model) SelectID(id string) {
	for i, item := range m.items {
		if item.ID == id && !item.Disabled {
			m.Select(i)
			return
		}
	}
}

// Selected returns the selected item index, or -1 when no enabled item exists.
func (m Model) Selected() int { return m.selected }

// SelectedItem returns the selected item, or a zero Item when there is no
// selection.
func (m Model) SelectedItem() Item {
	if m.selected < 0 || m.selected >= len(m.items) {
		return Item{}
	}
	return m.items[m.selected]
}

// SelectedID returns the selected item's stable ID, or an empty string when
// there is no selection or the item has no ID.
func (m Model) SelectedID() string { return m.SelectedItem().ID }

// Focus makes the menu respond to navigation and activation keys.
func (m *Model) Focus() { m.focused = true }

// Blur stops the menu from responding to keys.
func (m *Model) Blur() { m.focused = false }

// Focused reports whether the menu handles keys.
func (m Model) Focused() bool { return m.focused }

// TotalLines returns the configured item count for scrollbar.Scrollable.
func (m Model) TotalLines() int { return len(m.items) }

// VisibleLines returns the number of visible menu rows.
func (m Model) VisibleLines() int { return m.height }

// YOffset returns the first visible item index.
func (m Model) YOffset() int { return m.off }

// Activate returns a command that emits SelectedMsg for the current enabled
// selection, or nil when the menu has no selection.
func (m Model) Activate() tea.Cmd {
	item := m.SelectedItem()
	if m.selected < 0 || item.Disabled {
		return nil
	}
	return func() tea.Msg {
		return SelectedMsg{ID: item.ID, Index: m.selected, Label: item.Label}
	}
}

// Update handles semantic actions and, while focused, up/down arrows, j/k,
// page navigation, home/g, end/G, and enter activation.
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
	case "up", "k":
		m.selectRelative(-1)
	case "down", "j":
		m.selectRelative(1)
	case "pgup":
		m.selectRelative(-max(1, m.height))
	case "pgdown":
		m.selectRelative(max(1, m.height))
	case "home", "g":
		m.Select(m.firstEnabled())
	case "end", "G":
		m.Select(m.lastEnabled())
	case "enter":
		return m, m.Activate()
	}
	return m, nil
}

// View renders the visible menu rows. Selection uses SelectionBg/SelectionFg
// only while focused; a blurred menu keeps a subtle accent marker.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	rows := make([]string, m.height)
	for i := range rows {
		index := m.off + i
		if index >= len(m.items) {
			rows[i] = strings.Repeat(" ", m.width)
			continue
		}
		rows[i] = m.renderItem(index)
	}
	return strings.Join(rows, "\n")
}

func (m Model) renderItem(index int) string {
	item := m.items[index]
	if m.width == 1 {
		text := ansi.Truncate(item.Label, 1, "…")
		return m.styleFor(index).Render(text)
	}

	textWidth := m.width - 2
	text := ansi.Truncate(item.Label, textWidth, "…")
	text += strings.Repeat(" ", max(0, textWidth-ansi.StringWidth(text)))
	marker := " "
	if index == m.selected {
		marker = "▸"
	}
	return m.styleFor(index).Render(marker + " " + text)
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
	remaining := abs(delta)
	index := m.selected
	for remaining > 0 {
		index += step
		for index >= 0 && index < len(m.items) && m.items[index].Disabled {
			index += step
		}
		if index < 0 || index >= len(m.items) {
			return
		}
		remaining--
	}
	m.Select(index)
}

func (m *Model) scrollIntoView() {
	if len(m.items) == 0 || m.height <= 0 {
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
	if m.selected >= m.off+m.height {
		m.off = m.selected - m.height + 1
	}
	m.off = max(0, min(m.off, max(0, len(m.items)-m.height)))
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
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
