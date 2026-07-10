package table

import (
	"strconv"
	"strings"

	"github.com/ishansain/gotui/inspect"
)

// Inspect reports table selection, focus, and scroll state. Row contents are
// summarized as a label so the application can decide whether exposing them
// to an inspection consumer is appropriate.
func (m Model) Inspect() inspect.Node {
	selected := m.Selected()
	var selection *inspect.Selection
	if selected >= 0 {
		selection = &inspect.Selection{
			Index: selected,
			Count: len(m.rows),
			Label: strings.Join(m.SelectedRow(), " | "),
		}
	}
	return inspect.Node{
		Kind:     "table",
		Bounds:   inspect.Bounds{Width: m.width, Height: m.height},
		Focused:  m.focused,
		Selected: selection,
		Scroll:   &inspect.Scroll{Total: m.TotalLines(), Visible: m.VisibleLines(), Offset: m.YOffset()},
		Attributes: map[string]string{
			"column_count": strconv.Itoa(len(m.cols)),
			"row_count":    strconv.Itoa(len(m.rows)),
		},
	}
}
