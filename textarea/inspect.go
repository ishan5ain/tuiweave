package textarea

import (
	"strconv"

	"github.com/ishansain/gotui/inspect"
)

// Inspect reports textarea focus, cursor position, and content shape without
// copying the user's text into the semantic tree.
func (m Model) Inspect() inspect.Node {
	return inspect.Node{
		Kind:    "textarea",
		Bounds:  inspect.Bounds{Width: m.width, Height: m.height},
		Label:   m.Prompt,
		Focused: m.focused,
		Attributes: map[string]string{
			"empty":          strconv.FormatBool(m.Empty()),
			"logical_lines":  strconv.Itoa(len(m.lines)),
			"content_height": strconv.Itoa(m.ContentHeight()),
			"cursor_row":     strconv.Itoa(m.row),
			"cursor_column":  strconv.Itoa(m.col),
		},
	}
}
