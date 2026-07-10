package list

import (
	"strconv"

	"github.com/ishansain/gotui/inspect"
)

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
		Attributes: attributes,
	}
}
