package textinput

import (
	"strconv"

	"github.com/ishansain/gotui/inspect"
)

// Inspect reports input focus and non-sensitive editing state. The current
// value is not included automatically; applications can expose it separately
// when their inspection surface is trusted to receive user input.
func (m Model) Inspect() inspect.Node {
	return inspect.Node{
		Kind:    "textinput",
		Bounds:  inspect.Bounds{Width: m.width, Height: m.height},
		Label:   m.Prompt,
		Focused: m.focused,
		Attributes: map[string]string{
			"empty":       strconv.FormatBool(len(m.value) == 0),
			"placeholder": m.Placeholder,
			"cursor":      strconv.Itoa(m.pos),
		},
	}
}
