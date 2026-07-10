package viewport

import "github.com/ishansain/gotui/inspect"

// Inspect reports focus and scroll state. The rendered content is deliberately
// not copied into the semantic tree; callers can inspect source content through
// their own application model when needed.
func (m Model) Inspect() inspect.Node {
	return inspect.Node{
		Kind:    "viewport",
		Bounds:  inspect.Bounds{Width: m.width, Height: m.height},
		Focused: m.focused,
		Scroll:  &inspect.Scroll{Total: m.TotalLines(), Visible: m.VisibleLines(), Offset: m.YOffset()},
	}
}
