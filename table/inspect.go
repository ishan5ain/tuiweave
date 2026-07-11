package table

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/ishan5ain/tuiweave/inspect"
)

// Actions reports stable local intents for table navigation and focus.
func (m Model) Actions() []inspect.Action {
	return []inspect.Action{
		{ID: ActionFocus, Label: "Focus table", Enabled: !m.focused},
		{ID: ActionBlur, Label: "Blur table", Enabled: m.focused},
		{ID: ActionNext, Label: "Next row", Description: "Select the next row", Enabled: m.sel < len(m.rows)-1},
		{ID: ActionPrevious, Label: "Previous row", Description: "Select the previous row", Enabled: m.sel > 0},
		{ID: ActionFirst, Label: "First row", Enabled: m.sel > 0},
		{ID: ActionLast, Label: "Last row", Enabled: len(m.rows) > 0 && m.sel < len(m.rows)-1},
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
		m.Select(m.sel + 1)
	case ActionPrevious:
		m.Select(m.sel - 1)
	case ActionFirst:
		m.Select(0)
	case ActionLast:
		m.Select(len(m.rows) - 1)
	default:
		return m, false
	}
	return m, true
}

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
		Actions:  m.Actions(),
		Attributes: map[string]string{
			"column_count": strconv.Itoa(len(m.cols)),
			"row_count":    strconv.Itoa(len(m.rows)),
		},
	}
}
