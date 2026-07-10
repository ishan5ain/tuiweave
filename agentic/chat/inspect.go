package chat

import (
	"strconv"

	tea "charm.land/bubbletea/v2"

	"github.com/ishansain/gotui/inspect"
)

const (
	ActionFocus      = "focus"
	ActionBlur       = "blur"
	ActionScrollUp   = "scroll_up"
	ActionScrollDown = "scroll_down"
	ActionTop        = "scroll_top"
	ActionBottom     = "scroll_bottom"
)

// Actions reports stable local intents for transcript focus and scrolling.
func (m Model) Actions() []inspect.Action {
	maxOffset := max(0, m.TotalLines()-m.VisibleLines())
	return []inspect.Action{
		{ID: ActionFocus, Label: "Focus transcript", Enabled: !m.Focused()},
		{ID: ActionBlur, Label: "Blur transcript", Enabled: m.Focused()},
		{ID: ActionScrollUp, Label: "Scroll up", Enabled: m.YOffset() > 0},
		{ID: ActionScrollDown, Label: "Scroll down", Enabled: m.YOffset() < maxOffset},
		{ID: ActionTop, Label: "Scroll to top", Enabled: m.YOffset() > 0},
		{ID: ActionBottom, Label: "Scroll to bottom", Enabled: m.YOffset() < maxOffset},
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
	case ActionScrollUp:
		m.vp.ScrollBy(-1)
	case ActionScrollDown:
		m.vp.ScrollBy(1)
	case ActionTop:
		m.vp.GotoTop()
	case ActionBottom:
		m.GotoBottom()
	default:
		return m, false
	}
	m.follow = m.vp.AtBottom()
	return m, true
}

// Inspect reports transcript focus, follow state, cell count, and scroll
// state. Cell contents remain owned by the application and are not copied into
// the semantic tree.
func (m Model) Inspect() inspect.Node {
	return inspect.Node{
		Kind:    "chat",
		Bounds:  inspect.Bounds{Width: m.width, Height: m.height},
		Focused: m.Focused(),
		Scroll:  &inspect.Scroll{Total: m.TotalLines(), Visible: m.VisibleLines(), Offset: m.YOffset()},
		Actions: m.Actions(),
		Attributes: map[string]string{
			"cell_count": strconv.Itoa(len(m.cells)),
			"following":  strconv.FormatBool(m.follow),
		},
	}
}
