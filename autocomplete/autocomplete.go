// Package autocomplete provides a conditional, width-bounded suggestion window for an
// application-owned text input. The app owns the query and insertion policy;
// this component owns matching, navigation, rendering, and semantic
// activation.
package autocomplete

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
	// ActionFocus focuses the suggestion window.
	ActionFocus = "focus"
	// ActionBlur blurs the suggestion window.
	ActionBlur = "blur"
	// ActionNext selects the next enabled suggestion.
	ActionNext = "next"
	// ActionPrevious selects the previous enabled suggestion.
	ActionPrevious = "previous"
	// ActionFirst selects the first enabled suggestion.
	ActionFirst = "first"
	// ActionLast selects the last enabled suggestion.
	ActionLast = "last"
	// ActionClear clears the query and restores all suggestions.
	ActionClear = "clear"
	// ActionActivate accepts the selected suggestion.
	ActionActivate = "activate"
	// ActionSelectPrefix prefixes stable suggestion-selection IDs.
	ActionSelectPrefix = action.SelectPrefix
)

// Item is one completion candidate. Value is inserted by the application;
// Label is displayed when non-empty and otherwise falls back to Value.
type Item struct {
	ID          string
	Value       string
	Label       string
	Description string
	Disabled    bool
}

// SelectedMsg reports an accepted completion. The application decides how to
// replace or append text in its input using Value.
type SelectedMsg struct {
	ID    string
	Index int
	Value string
	Label string
}

// Model is a width-bounded, focusable suggestion window. It fills its assigned
// height while matches exist and renders empty when there are no matches. The
// query is deliberately app-owned so the same component can sit beside
// textinput, textarea, or a domain-specific editor.
type Model struct {
	width, height int
	items         []Item
	matches       []int
	query         string
	pos           int
	off           int
	focused       bool

	itemStyle            lipgloss.Style
	descriptionStyle     lipgloss.Style
	selectedStyle        lipgloss.Style
	selectedBlurredStyle lipgloss.Style
	disabledStyle        lipgloss.Style
}

// New returns an empty autocomplete window styled from semantic theme roles.
func New(theme tuiweave.Theme) Model {
	base := lipgloss.NewStyle().Background(theme.SurfaceRaised)
	return Model{
		pos: -1,
		itemStyle: base.
			Foreground(theme.Text),
		descriptionStyle: base.
			Foreground(theme.TextMuted),
		selectedStyle: lipgloss.NewStyle().
			Foreground(theme.SelectionFg).
			Background(theme.SelectionBg),
		selectedBlurredStyle: base.
			Foreground(theme.Accent).
			Bold(true),
		disabledStyle: base.
			Foreground(theme.TextFaint),
	}
}

// SetSize sets the suggestion window's bounded box. A non-empty window fills
// its assigned height with visible suggestions and blank rows.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
	m.scrollIntoView()
}

// SizeMode reports that the suggestion window is width-constrained while its
// height is conditional: it fills the assigned height when visible and renders
// empty when there are no matches.
func (m Model) SizeMode() layout.SizeMode { return layout.SizeWidthBounded }

// SetItems replaces candidates while preserving the selected ID when it is
// still enabled and matches the current query.
func (m *Model) SetItems(items ...Item) {
	keepIndex, keepID := m.Selected(), m.SelectedID()
	m.items = append([]Item(nil), items...)
	m.applyFilter(keepIndex, keepID)
}

// Items returns the configured candidates.
func (m Model) Items() []Item { return append([]Item(nil), m.items...) }

// SetQuery updates the app-owned query. Matching is case-insensitive and uses
// a prefix of Value or Label, with Description as a searchable fallback.
func (m *Model) SetQuery(query string) {
	keepIndex, keepID := m.Selected(), m.SelectedID()
	m.query = query
	m.applyFilter(keepIndex, keepID)
}

// Query returns the current query used for matching.
func (m Model) Query() string { return m.query }

// FilteredLen returns the number of matching candidates, including disabled
// candidates that remain visible as unavailable suggestions.
func (m Model) FilteredLen() int { return len(m.matches) }

// Select selects an enabled candidate by original item index when it matches
// the current query.
func (m *Model) Select(index int) {
	for position, original := range m.matches {
		if original == index && !m.items[original].Disabled {
			m.pos = position
			m.scrollIntoView()
			return
		}
	}
}

// SelectID selects an enabled candidate by stable ID.
func (m *Model) SelectID(id string) {
	for index, item := range m.items {
		if item.ID == id && !item.Disabled {
			m.Select(index)
			return
		}
	}
}

// Selected returns the selected candidate's original item index, or -1.
func (m Model) Selected() int {
	if m.pos < 0 || m.pos >= len(m.matches) {
		return -1
	}
	original := m.matches[m.pos]
	if m.items[original].Disabled {
		return -1
	}
	return original
}

// SelectedItem returns the selected candidate, or a zero Item.
func (m Model) SelectedItem() Item {
	selected := m.Selected()
	if selected < 0 {
		return Item{}
	}
	return m.items[selected]
}

// SelectedID returns the selected candidate's stable ID, or an empty string.
func (m Model) SelectedID() string { return m.SelectedItem().ID }

// Focus makes the suggestion window respond to navigation and activation.
func (m *Model) Focus() { m.focused = true }

// Blur stops the suggestion window from responding to keys.
func (m *Model) Blur() { m.focused = false }

// Focused reports whether the suggestion window accepts keyboard input.
func (m Model) Focused() bool { return m.focused }

// TotalLines returns the number of matching candidates.
func (m Model) TotalLines() int { return len(m.matches) }

// VisibleLines returns the number of matching rows that fit in the box.
func (m Model) VisibleLines() int { return min(m.height, len(m.matches)) }

// YOffset returns the first visible matching position.
func (m Model) YOffset() int { return m.off }

// Activate returns a command for the selected candidate, or nil when no
// enabled candidate is selected.
func (m Model) Activate() tea.Cmd {
	item := m.SelectedItem()
	if item.ID == "" && item.Value == "" && item.Label == "" {
		return nil
	}
	value := item.Value
	if value == "" {
		value = item.Label
	}
	label := item.Label
	if label == "" {
		label = value
	}
	return func() tea.Msg {
		return SelectedMsg{ID: item.ID, Index: m.Selected(), Value: value, Label: label}
	}
}

// Update handles semantic actions and, while focused, navigation and enter.
// Query editing remains in the application-owned input.
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
		m.selectRelative(-max(1, m.VisibleLines()))
	case "pgdown":
		m.selectRelative(max(1, m.VisibleLines()))
	case "home", "g":
		m.selectFirst()
	case "end", "G":
		m.selectLast()
	case "enter":
		return m, m.Activate()
	}
	return m, nil
}

// View renders visible suggestions at the assigned width. It returns empty
// when there are no matches so an application can hide its popup naturally.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 || len(m.matches) == 0 {
		return ""
	}
	rows := make([]string, m.height)
	for i := range rows {
		position := m.off + i
		if position >= len(m.matches) {
			rows[i] = strings.Repeat(" ", m.width)
			continue
		}
		rows[i] = m.renderItem(position)
	}
	return strings.Join(rows, "\n")
}

func (m Model) renderItem(position int) string {
	item := m.items[m.matches[position]]
	label := item.Label
	if label == "" {
		label = item.Value
	}
	if m.width == 1 {
		return m.styleFor(position, item).Render(ansi.Truncate(label, 1, "…"))
	}

	marker := "  "
	if position == m.pos {
		marker = "▸ "
	}
	available := max(0, m.width-2)
	label = ansi.Truncate(label, available, "…")
	description := ""
	gap := 0
	if item.Description != "" && available >= 4 {
		label = ansi.Truncate(label, available-3, "…")
		gap = 2
		descriptionWidth := available - ansi.StringWidth(label) - gap
		if descriptionWidth > 0 {
			description = ansi.Truncate(item.Description, descriptionWidth, "…")
		} else {
			gap = 0
		}
	}
	used := 2 + ansi.StringWidth(label) + gap + ansi.StringWidth(description)
	padding := strings.Repeat(" ", max(0, m.width-used))
	if item.Disabled || position == m.pos {
		line := marker + label + strings.Repeat(" ", gap) + description + padding
		return m.styleFor(position, item).Render(line)
	}
	return m.itemStyle.Render(marker+label+strings.Repeat(" ", gap)) +
		m.descriptionStyle.Render(description) + m.itemStyle.Render(padding)
}

func (m Model) styleFor(position int, item Item) lipgloss.Style {
	if item.Disabled {
		return m.disabledStyle
	}
	if position == m.pos {
		if m.focused {
			return m.selectedStyle
		}
		return m.selectedBlurredStyle
	}
	return m.itemStyle
}

func (m *Model) selectFirst() {
	m.pos = m.firstEnabledPosition()
	m.scrollIntoView()
}

func (m *Model) selectLast() {
	m.pos = m.lastEnabledPosition()
	m.scrollIntoView()
}

func (m *Model) selectRelative(delta int) {
	if len(m.matches) == 0 {
		m.pos = -1
		return
	}
	if m.Selected() < 0 {
		if delta < 0 {
			m.selectLast()
		} else {
			m.selectFirst()
		}
		return
	}
	step := 1
	if delta < 0 {
		step = -1
	}
	remaining := abs(delta)
	position := m.pos
	for remaining > 0 {
		position += step
		for position >= 0 && position < len(m.matches) && m.items[m.matches[position]].Disabled {
			position += step
		}
		if position < 0 || position >= len(m.matches) {
			return
		}
		remaining--
	}
	m.pos = position
	m.scrollIntoView()
}

func (m *Model) scrollIntoView() {
	visible := m.VisibleLines()
	if visible <= 0 || m.pos < 0 {
		m.off = 0
		return
	}
	if m.pos < m.off {
		m.off = m.pos
	}
	if m.pos >= m.off+visible {
		m.off = m.pos - visible + 1
	}
	m.off = max(0, min(m.off, max(0, len(m.matches)-visible)))
}

func (m *Model) applyFilter(keepIndex int, keepID string) {
	query := strings.ToLower(m.query)
	m.matches = m.matches[:0]
	for i, item := range m.items {
		if query == "" || matches(query, item) {
			m.matches = append(m.matches, i)
		}
	}

	m.pos = -1
	if keepID != "" {
		for position, original := range m.matches {
			if m.items[original].ID == keepID && !m.items[original].Disabled {
				m.pos = position
				break
			}
		}
	}
	if m.pos < 0 && keepIndex >= 0 {
		for position, original := range m.matches {
			if original == keepIndex && !m.items[original].Disabled {
				m.pos = position
				break
			}
		}
	}
	if m.pos < 0 {
		m.pos = m.firstEnabledPosition()
	}
	m.scrollIntoView()
}

func matches(query string, item Item) bool {
	for _, value := range []string{item.Value, item.Label} {
		if strings.HasPrefix(strings.ToLower(value), query) {
			return true
		}
	}
	return strings.Contains(strings.ToLower(item.Description), query)
}

func (m Model) firstEnabledPosition() int {
	for position, original := range m.matches {
		if !m.items[original].Disabled {
			return position
		}
	}
	return -1
}

func (m Model) lastEnabledPosition() int {
	for position := len(m.matches) - 1; position >= 0; position-- {
		if !m.items[m.matches[position]].Disabled {
			return position
		}
	}
	return -1
}

func (m Model) applyAction(msg tea.Msg) (Model, tea.Cmd, bool) {
	event, ok := msg.(inspect.ActionMsg)
	if !ok {
		return m, nil, false
	}
	switch event.ID {
	case ActionFocus:
		m.Focus()
	case ActionBlur:
		m.Blur()
	case ActionNext:
		m.selectRelative(1)
	case ActionPrevious:
		m.selectRelative(-1)
	case ActionFirst:
		m.selectFirst()
	case ActionLast:
		m.selectLast()
	case ActionClear:
		m.SetQuery("")
	case ActionActivate:
		return m, m.Activate(), true
	default:
		id, ok := action.ParseSelectID(event.ID)
		if !ok {
			return m, nil, false
		}
		for _, item := range m.items {
			if item.ID == id && !item.Disabled {
				m.SelectID(id)
				return m, nil, true
			}
		}
		return m, nil, false
	}
	return m, nil, true
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
