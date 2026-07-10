package textarea

import (
	"strconv"

	tea "charm.land/bubbletea/v2"

	"github.com/ishansain/gotui/inspect"
)

// Actions reports stable local intents for textarea focus, selection, and
// history.
func (m Model) Actions() []inspect.Action {
	return []inspect.Action{
		{ID: ActionFocus, Label: "Focus textarea", Enabled: !m.focused},
		{ID: ActionBlur, Label: "Blur textarea", Enabled: m.focused},
		{ID: ActionClear, Label: "Clear textarea", Enabled: !m.Empty()},
		{ID: ActionUndo, Label: "Undo edit", Enabled: m.CanUndo()},
		{ID: ActionRedo, Label: "Redo edit", Enabled: m.CanRedo()},
		{ID: ActionSelectAll, Label: "Select all text", Enabled: !m.Empty()},
		{ID: ActionClearSelection, Label: "Clear selection", Enabled: m.HasSelection()},
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
	case ActionClear:
		m.Reset()
	case ActionUndo:
		m.Undo()
	case ActionRedo:
		m.Redo()
	case ActionSelectAll:
		m.SelectAll()
	case ActionClearSelection:
		m.ClearSelection()
	default:
		return m, false
	}
	return m, true
}

// Inspect reports textarea focus, cursor position, and content shape without
// copying the user's text into the semantic tree.
func (m Model) Inspect() inspect.Node {
	return inspect.Node{
		Kind:    "textarea",
		Bounds:  inspect.Bounds{Width: m.width, Height: m.height},
		Label:   m.Prompt,
		Focused: m.focused,
		Actions: m.Actions(),
		Attributes: map[string]string{
			"empty":          strconv.FormatBool(m.Empty()),
			"logical_lines":  strconv.Itoa(len(m.lines)),
			"content_height": strconv.Itoa(m.ContentHeight()),
			"cursor_row":     strconv.Itoa(m.row),
			"cursor_column":  strconv.Itoa(m.col),
			"has_selection":  strconv.FormatBool(m.HasSelection()),
			"selected_runes": strconv.Itoa(len([]rune(m.SelectedText()))),
			"can_undo":       strconv.FormatBool(m.CanUndo()),
			"can_redo":       strconv.FormatBool(m.CanRedo()),
		},
	}
}
