package chat

import (
	"strconv"

	"github.com/ishansain/gotui/inspect"
)

// Inspect reports transcript focus, follow state, cell count, and scroll
// state. Cell contents remain owned by the application and are not copied into
// the semantic tree.
func (m Model) Inspect() inspect.Node {
	return inspect.Node{
		Kind:    "chat",
		Bounds:  inspect.Bounds{Width: m.width, Height: m.height},
		Focused: m.Focused(),
		Scroll:  &inspect.Scroll{Total: m.TotalLines(), Visible: m.VisibleLines(), Offset: m.YOffset()},
		Attributes: map[string]string{
			"cell_count": strconv.Itoa(len(m.cells)),
			"following":  strconv.FormatBool(m.follow),
		},
	}
}
