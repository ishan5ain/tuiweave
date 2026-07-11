package viewport

import (
	tea "charm.land/bubbletea/v2"

	"github.com/ishan5ain/tuiweave/inspect"
)

const (
	ActionFocus      = "focus"
	ActionBlur       = "blur"
	ActionScrollUp   = "scroll_up"
	ActionScrollDown = "scroll_down"
	ActionTop        = "scroll_top"
	ActionBottom     = "scroll_bottom"
)

// Actions reports stable local intents for scrolling and focus.
func (m Model) Actions() []inspect.Action {
	return []inspect.Action{
		{ID: ActionFocus, Label: "Focus viewport", Enabled: !m.focused},
		{ID: ActionBlur, Label: "Blur viewport", Enabled: m.focused},
		{ID: ActionScrollUp, Label: "Scroll up", Enabled: m.yoff > 0},
		{ID: ActionScrollDown, Label: "Scroll down", Enabled: m.yoff < m.maxYOffset()},
		{ID: ActionTop, Label: "Scroll to top", Enabled: m.yoff > 0},
		{ID: ActionBottom, Label: "Scroll to bottom", Enabled: m.yoff < m.maxYOffset()},
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
		m.ScrollBy(-1)
	case ActionScrollDown:
		m.ScrollBy(1)
	case ActionTop:
		m.GotoTop()
	case ActionBottom:
		m.GotoBottom()
	default:
		return m, false
	}
	return m, true
}

// Inspect reports focus and scroll state. The rendered content is deliberately
// not copied into the semantic tree; callers can inspect source content through
// their own application model when needed.
func (m Model) Inspect() inspect.Node {
	return inspect.Node{
		Kind:    "viewport",
		Bounds:  inspect.Bounds{Width: m.width, Height: m.height},
		Focused: m.focused,
		Scroll:  &inspect.Scroll{Total: m.TotalLines(), Visible: m.VisibleLines(), Offset: m.YOffset()},
		Actions: m.Actions(),
	}
}
