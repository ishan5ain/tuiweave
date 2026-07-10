package tabs

import (
	"strconv"

	"github.com/ishansain/gotui/inspect"
)

// Actions reports stable local intents for tab focus and selection.
func (m Model) Actions() []inspect.Action {
	actions := []inspect.Action{
		{ID: ActionFocus, Label: "Focus tabs", Enabled: !m.focused},
		{ID: ActionBlur, Label: "Blur tabs", Enabled: m.focused},
		{ID: ActionNext, Label: "Next tab", Description: "Select the next tab", Enabled: m.selected >= 0 && m.selected < len(m.tabs)-1},
		{ID: ActionPrevious, Label: "Previous tab", Description: "Select the previous tab", Enabled: m.selected > 0},
		{ID: ActionFirst, Label: "First tab", Enabled: m.selected > 0},
		{ID: ActionLast, Label: "Last tab", Enabled: len(m.tabs) > 0 && m.selected < len(m.tabs)-1},
	}
	for _, tab := range m.tabs {
		if tab.ID == "" {
			continue
		}
		actions = append(actions, inspect.Action{
			ID:          ActionSelectPrefix + tab.ID,
			Label:       "Select " + tab.Label,
			Description: "Select this tab",
			Enabled:     true,
		})
	}
	return actions
}

// Inspect reports tab selection, focus, and stable selection actions.
func (m Model) Inspect() inspect.Node {
	var selected *inspect.Selection
	if m.selected >= 0 && m.selected < len(m.tabs) {
		selected = &inspect.Selection{
			Index: m.selected,
			Count: len(m.tabs),
			Label: m.tabs[m.selected].Label,
		}
	}
	attributes := map[string]string{
		"tab_count":     strconv.Itoa(len(m.tabs)),
		"visible_start": strconv.Itoa(m.off),
	}
	if id := m.SelectedID(); id != "" {
		attributes["selected_id"] = id
	}
	return inspect.Node{
		Kind:       "tabs",
		Bounds:     inspect.Bounds{Width: m.width, Height: m.height},
		Label:      m.SelectedTab().Label,
		Focused:    m.focused,
		Selected:   selected,
		Actions:    m.Actions(),
		Attributes: attributes,
	}
}
