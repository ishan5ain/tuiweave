package list

import (
	"strconv"

	tea "charm.land/bubbletea/v2"

	"github.com/ishan5ain/tuiweave/inspect"
)

// Actions reports stable local intents for list navigation and focus.
func (m Model) Actions() []inspect.Action {
	return []inspect.Action{
		{ID: ActionFocus, Label: "Focus list", Enabled: !m.focused},
		{ID: ActionBlur, Label: "Blur list", Enabled: m.focused},
		{ID: ActionNext, Label: "Next item", Description: "Select the next visible item", Enabled: m.pos < m.displayCount()-1},
		{ID: ActionPrevious, Label: "Previous item", Description: "Select the previous visible item", Enabled: m.pos > 0},
		{ID: ActionFirst, Label: "First item", Enabled: m.pos > 0},
		{ID: ActionLast, Label: "Last item", Enabled: m.displayCount() > 0 && m.pos < m.displayCount()-1},
	}
}

func (m Model) applyAction(msg tea.Msg) (Model, bool) {
	action, ok := msg.(inspect.ActionMsg)
	if !ok {
		return m, false
	}
	switch action.ID {
	case ActionFocus:
		m.Focus()
	case ActionBlur:
		m.Blur()
	case ActionNext:
		m.selectPos(m.pos + 1)
	case ActionPrevious:
		m.selectPos(m.pos - 1)
	case ActionFirst:
		m.selectPos(0)
	case ActionLast:
		m.selectPos(m.displayCount() - 1)
	default:
		return m, false
	}
	return m, true
}

// Inspect reports the list's selection, filtering, focus, and scroll state.
// The application supplies the stable ID when it assembles this node.
func (m Model) Inspect() inspect.Node {
	selected := m.Selected()
	var selection *inspect.Selection
	if selected >= 0 {
		selection = &inspect.Selection{
			Index: selected,
			Count: m.displayCount(),
			Label: m.SelectedItem(),
		}
	}
	attributes := map[string]string{
		"item_count":     strconv.Itoa(len(m.items)),
		"filtered_count": strconv.Itoa(m.displayCount()),
	}
	if m.filter != "" {
		attributes["filter"] = m.filter
	}
	return inspect.Node{
		Kind:       "list",
		Bounds:     inspect.Bounds{Width: m.width, Height: m.height},
		Focused:    m.focused,
		Selected:   selection,
		Scroll:     &inspect.Scroll{Total: m.TotalLines(), Visible: m.VisibleLines(), Offset: m.YOffset()},
		Actions:    m.Actions(),
		Attributes: attributes,
	}
}
