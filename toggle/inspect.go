package toggle

import (
	"strconv"

	"github.com/ishansain/gotui/inspect"
)

// Actions reports stable local intents for focus and changing the toggle.
func (m Model) Actions() []inspect.Action {
	return []inspect.Action{
		{ID: ActionFocus, Label: "Focus toggle", Enabled: !m.focused && !m.disabled},
		{ID: ActionBlur, Label: "Blur toggle", Enabled: m.focused},
		{ID: ActionToggle, Label: "Toggle " + m.label, Enabled: !m.disabled},
		{ID: ActionOn, Label: "Turn on " + m.label, Enabled: !m.disabled && !m.checked},
		{ID: ActionOff, Label: "Turn off " + m.label, Enabled: !m.disabled && m.checked},
	}
}

// Inspect reports the toggle's state, focus, and stable semantic actions.
func (m Model) Inspect() inspect.Node {
	attributes := map[string]string{
		"checked":  strconv.FormatBool(m.checked),
		"disabled": strconv.FormatBool(m.disabled),
	}
	if m.ID != "" {
		attributes["id"] = m.ID
	}
	status := "off"
	if m.checked {
		status = "on"
	}
	return inspect.Node{
		Kind:       "toggle",
		Bounds:     inspect.Bounds{Width: m.width, Height: m.height},
		Label:      m.label,
		Focused:    m.focused,
		Status:     status,
		Actions:    m.Actions(),
		Attributes: attributes,
	}
}
