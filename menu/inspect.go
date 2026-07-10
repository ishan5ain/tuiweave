package menu

import (
	"strconv"

	"github.com/ishansain/gotui/inspect"
)

// Actions reports stable local intents for menu focus, navigation, selection,
// and activation. Disabled items remain discoverable with Enabled=false.
func (m Model) Actions() []inspect.Action {
	actions := []inspect.Action{
		{ID: ActionFocus, Label: "Focus menu", Enabled: !m.focused},
		{ID: ActionBlur, Label: "Blur menu", Enabled: m.focused},
		{ID: ActionNext, Label: "Next menu item", Enabled: m.nextEnabled() >= 0},
		{ID: ActionPrevious, Label: "Previous menu item", Enabled: m.previousEnabled() >= 0},
		{ID: ActionFirst, Label: "First menu item", Enabled: m.firstEnabled() >= 0},
		{ID: ActionLast, Label: "Last menu item", Enabled: m.lastEnabled() >= 0},
		{ID: ActionActivate, Label: "Activate menu item", Enabled: m.selected >= 0},
	}
	for _, item := range m.items {
		if item.ID == "" {
			continue
		}
		actions = append(actions, inspect.Action{
			ID:          ActionSelectPrefix + item.ID,
			Label:       "Select " + item.Label,
			Description: item.Description,
			Enabled:     !item.Disabled,
		})
	}
	return actions
}

// Inspect reports menu selection, disabled state counts, focus, and scroll
// position. It intentionally does not include application command details.
func (m Model) Inspect() inspect.Node {
	var selected *inspect.Selection
	if m.selected >= 0 && m.selected < len(m.items) {
		selected = &inspect.Selection{
			Index: m.selected,
			Count: len(m.items),
			Label: m.items[m.selected].Label,
		}
	}
	enabled := 0
	for _, item := range m.items {
		if !item.Disabled {
			enabled++
		}
	}
	attributes := map[string]string{
		"item_count":    strconv.Itoa(len(m.items)),
		"enabled_count": strconv.Itoa(enabled),
	}
	if id := m.SelectedID(); id != "" {
		attributes["selected_id"] = id
	}
	return inspect.Node{
		Kind:       "menu",
		Bounds:     inspect.Bounds{Width: m.width, Height: m.height},
		Label:      m.SelectedItem().Label,
		Focused:    m.focused,
		Selected:   selected,
		Scroll:     &inspect.Scroll{Total: m.TotalLines(), Visible: m.VisibleLines(), Offset: m.YOffset()},
		Actions:    m.Actions(),
		Attributes: attributes,
	}
}

func (m Model) nextEnabled() int {
	if m.selected < 0 {
		return m.firstEnabled()
	}
	for i := m.selected + 1; i < len(m.items); i++ {
		if !m.items[i].Disabled {
			return i
		}
	}
	return -1
}

func (m Model) previousEnabled() int {
	if m.selected < 0 {
		return m.lastEnabled()
	}
	for i := m.selected - 1; i >= 0; i-- {
		if !m.items[i].Disabled {
			return i
		}
	}
	return -1
}
