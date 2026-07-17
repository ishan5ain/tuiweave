package dialog

import (
	tea "charm.land/bubbletea/v2"

	"github.com/ishan5ain/tuiweave/inspect"
)

// Actions reports stable local intents for answering the dialog.
func (m Model) Actions() []inspect.Action {
	return []inspect.Action{
		{ID: ActionConfirm, Label: m.ConfirmLabel, Description: "Accept the dialog", Enabled: true},
		{ID: ActionCancel, Label: m.CancelLabel, Description: "Dismiss the dialog", Enabled: true},
		{ID: ActionNext, Label: "Next button", Enabled: m.sel == 0},
		{ID: ActionPrevious, Label: "Previous button", Enabled: m.sel == 1},
	}
}

func (m Model) applyAction(msg tea.Msg) (Model, tea.Cmd, bool) {
	action, ok := msg.(inspect.ActionMsg)
	if !ok {
		return m, nil, false
	}
	switch action.ID {
	case ActionConfirm:
		return m, m.result(true), true
	case ActionCancel:
		return m, m.result(false), true
	case ActionNext:
		m.sel = min(1, m.sel+1)
	case ActionPrevious:
		m.sel = max(0, m.sel-1)
	default:
		return m, nil, false
	}
	return m, nil, true
}

// Inspect reports the dialog's prompt and selected answer. Visibility remains
// application-owned, so an app should include this node only while rendering
// the dialog.
func (m Model) Inspect() inspect.Node {
	label := m.ConfirmLabel
	if m.sel == 1 {
		label = m.CancelLabel
	}
	return inspect.Node{
		Kind:    "dialog",
		Bounds:  inspect.Bounds{Width: m.width, Height: m.height},
		Label:   m.Title,
		Focused: true,
		Status:  "awaiting_input",
		Actions: m.Actions(),
		Selected: &inspect.Selection{
			Index: m.sel,
			Count: 2,
			Label: label,
		},
		Attributes: map[string]string{"body": m.Body},
	}
}
