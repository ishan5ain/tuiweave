package toolbar

import (
	"strconv"

	"github.com/ishan5ain/tuiweave/action"
	"github.com/ishan5ain/tuiweave/inspect"
)

// Actions reports stable local intents for toolbar focus, navigation,
// selection, and activation. Disabled items remain discoverable with
// Enabled=false.
func (m Model) Actions() []inspect.Action {
	actions := []inspect.Action{
		{ID: ActionFocus, Label: "Focus toolbar", Enabled: !m.focused},
		{ID: ActionBlur, Label: "Blur toolbar", Enabled: m.focused},
		{ID: ActionNext, Label: "Next toolbar action", Enabled: m.nextEnabled() >= 0},
		{ID: ActionPrevious, Label: "Previous toolbar action", Enabled: m.previousEnabled() >= 0},
		{ID: ActionFirst, Label: "First toolbar action", Enabled: m.firstEnabled() >= 0},
		{ID: ActionLast, Label: "Last toolbar action", Enabled: m.lastEnabled() >= 0},
		{ID: ActionActivate, Label: "Activate toolbar action", Enabled: m.selected >= 0},
	}
	for _, item := range m.items {
		if item.ID == "" {
			continue
		}
		actions = append(actions, inspect.Action{
			ID:          action.SelectID(item.ID),
			Label:       "Select " + item.Label,
			Description: item.Description,
			Enabled:     !item.Disabled,
		})
	}
	return actions
}

// Inspect reports toolbar selection, disabled state counts, focus, and the
// horizontal visible window.
func (m Model) Inspect() inspect.Node {
	var selected *inspect.Selection
	if m.selected >= 0 && m.selected < len(m.items) {
		selected = &inspect.Selection{
			Index: m.selected,
			Count: len(m.items),
			Label: m.items[m.selected].Label,
		}
	}
	attributes := map[string]string{
		"item_count":    strconv.Itoa(len(m.items)),
		"enabled_count": strconv.Itoa(action.EnabledCount(m.items)),
	}
	if id := m.SelectedID(); id != "" {
		attributes["selected_id"] = id
	}
	return inspect.Node{
		Kind:       "toolbar",
		Bounds:     inspect.Bounds{Width: m.width, Height: m.height},
		Label:      m.SelectedItem().Label,
		Focused:    m.focused,
		Selected:   selected,
		Scroll:     &inspect.Scroll{Total: len(m.items), Visible: m.visibleCount(), Offset: m.off},
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

func (m Model) visibleCount() int {
	if m.width <= 0 || m.off >= len(m.items) {
		return 0
	}
	used := 0
	count := 0
	for i := m.off; i < len(m.items); i++ {
		separator := 0
		if count > 0 {
			separator = 1
		}
		width := m.itemWidth(i)
		if used+separator+width > m.width {
			break
		}
		used += separator + width
		count++
	}
	return count
}
