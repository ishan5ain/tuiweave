package diffview

import (
	"github.com/ishan5ain/tuiweave/inspect"
	"github.com/ishan5ain/tuiweave/viewport"
)

const (
	// ActionFocus focuses the diff pane.
	ActionFocus = viewport.ActionFocus
	// ActionBlur blurs the diff pane.
	ActionBlur = viewport.ActionBlur
	// ActionScrollUp scrolls the diff pane up one line.
	ActionScrollUp = viewport.ActionScrollUp
	// ActionScrollDown scrolls the diff pane down one line.
	ActionScrollDown = viewport.ActionScrollDown
	// ActionTop scrolls the diff pane to its first line.
	ActionTop = viewport.ActionTop
	// ActionBottom scrolls the diff pane to its last visible window.
	ActionBottom = viewport.ActionBottom
)

// Actions reports stable local intents for diff focus and scrolling.
func (m Model) Actions() []inspect.Action {
	return []inspect.Action{
		{ID: ActionFocus, Label: "Focus diff", Enabled: !m.Focused()},
		{ID: ActionBlur, Label: "Blur diff", Enabled: m.Focused()},
		{ID: ActionScrollUp, Label: "Scroll diff up", Enabled: m.YOffset() > 0},
		{ID: ActionScrollDown, Label: "Scroll diff down", Enabled: m.YOffset() < max(0, m.TotalLines()-m.VisibleLines())},
		{ID: ActionTop, Label: "Scroll diff to top", Enabled: m.YOffset() > 0},
		{ID: ActionBottom, Label: "Scroll diff to bottom", Enabled: m.YOffset() < max(0, m.TotalLines()-m.VisibleLines())},
	}
}

// Inspect reports diff focus and scroll state without copying diff contents
// into the semantic tree. Applications decide whether source text is safe to
// expose separately.
func (m Model) Inspect() inspect.Node {
	node := m.vp.Inspect()
	node.Kind = "diffview"
	node.Actions = m.Actions()
	return node
}
