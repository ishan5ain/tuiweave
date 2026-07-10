package textinput

import (
	"strconv"

	tea "charm.land/bubbletea/v2"

	"github.com/ishansain/gotui/inspect"
)

// Actions reports stable local intents for input focus and clearing.
func (m Model) Actions() []inspect.Action {
	return []inspect.Action{
		{ID: ActionFocus, Label: "Focus input", Enabled: !m.focused},
		{ID: ActionBlur, Label: "Blur input", Enabled: m.focused},
		{ID: ActionClear, Label: "Clear input", Enabled: len(m.value) > 0},
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
	case ActionClear:
		m.Reset()
	default:
		return m, false
	}
	return m, true
}

// Inspect reports input focus and non-sensitive editing state. The current
// value is not included automatically; applications can expose it separately
// when their inspection surface is trusted to receive user input.
func (m Model) Inspect() inspect.Node {
	return inspect.Node{
		Kind:    "textinput",
		Bounds:  inspect.Bounds{Width: m.width, Height: m.height},
		Label:   m.Prompt,
		Focused: m.focused,
		Actions: m.Actions(),
		Attributes: map[string]string{
			"empty":       strconv.FormatBool(len(m.value) == 0),
			"placeholder": m.Placeholder,
			"cursor":      strconv.Itoa(m.pos),
		},
	}
}
