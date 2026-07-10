package button

import (
	"strconv"

	"github.com/ishansain/gotui/inspect"
)

// Actions reports stable local intents for button focus and activation.
func (m Model) Actions() []inspect.Action {
	return []inspect.Action{
		{ID: ActionFocus, Label: "Focus button", Enabled: !m.focused && !m.disabled},
		{ID: ActionBlur, Label: "Blur button", Enabled: m.focused},
		{ID: ActionActivate, Label: "Activate " + m.label, Enabled: !m.disabled},
	}
}

// Inspect reports the button's state and stable semantic actions.
func (m Model) Inspect() inspect.Node {
	attributes := map[string]string{
		"disabled": strconv.FormatBool(m.disabled),
	}
	if m.ID != "" {
		attributes["id"] = m.ID
	}
	status := "ready"
	if m.disabled {
		status = "disabled"
	}
	return inspect.Node{
		Kind:       "button",
		Bounds:     inspect.Bounds{Width: m.width, Height: m.height},
		Label:      m.label,
		Focused:    m.focused,
		Status:     status,
		Actions:    m.Actions(),
		Attributes: attributes,
	}
}
