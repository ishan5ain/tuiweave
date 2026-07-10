package dialog

import "github.com/ishansain/gotui/inspect"

// Inspect reports the dialog's prompt and selected answer. Visibility remains
// application-owned, so an app should include this node only while rendering
// the dialog.
func (m Model) Inspect() inspect.Node {
	label := m.ConfirmLabel
	if m.sel == 1 {
		label = m.CancelLabel
	}
	return inspect.Node{
		Kind:   "dialog",
		Bounds: inspect.Bounds{Width: m.width, Height: m.height},
		Label:  m.Title,
		Status: "awaiting_input",
		Selected: &inspect.Selection{
			Index: m.sel,
			Count: 2,
			Label: label,
		},
		Attributes: map[string]string{"body": m.Body},
	}
}
