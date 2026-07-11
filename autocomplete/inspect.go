package autocomplete

import (
	"strconv"

	"github.com/ishan5ain/tuiweave/action"
	"github.com/ishan5ain/tuiweave/inspect"
)

// Actions reports stable local intents for focus, navigation, selection, and
// activation.
func (m Model) Actions() []inspect.Action {
	actions := []inspect.Action{
		{ID: ActionFocus, Label: "Focus suggestions", Enabled: !m.focused},
		{ID: ActionBlur, Label: "Blur suggestions", Enabled: m.focused},
		{ID: ActionNext, Label: "Next suggestion", Enabled: m.nextEnabledPosition() >= 0},
		{ID: ActionPrevious, Label: "Previous suggestion", Enabled: m.previousEnabledPosition() >= 0},
		{ID: ActionFirst, Label: "First suggestion", Enabled: m.firstEnabledPosition() >= 0},
		{ID: ActionLast, Label: "Last suggestion", Enabled: m.lastEnabledPosition() >= 0},
		{ID: ActionClear, Label: "Clear completion query", Enabled: m.query != ""},
		{ID: ActionActivate, Label: "Accept suggestion", Enabled: m.Selected() >= 0},
	}
	for _, item := range m.items {
		if item.ID == "" {
			continue
		}
		actions = append(actions, inspect.Action{
			ID:          action.SelectID(item.ID),
			Label:       "Select " + displayLabel(item),
			Description: item.Description,
			Enabled:     !item.Disabled,
		})
	}
	return actions
}

// Inspect reports the query, filtered candidates, selection, focus, and
// visible suggestion window for tools and tests.
func (m Model) Inspect() inspect.Node {
	var selected *inspect.Selection
	if original := m.Selected(); original >= 0 {
		selected = &inspect.Selection{Index: original, Count: len(m.matches), Label: displayLabel(m.items[original])}
	}
	attributes := map[string]string{
		"item_count":     strconv.Itoa(len(m.items)),
		"filtered_count": strconv.Itoa(len(m.matches)),
		"query":          m.query,
	}
	if id := m.SelectedID(); id != "" {
		attributes["selected_id"] = id
	}
	return inspect.Node{
		Kind:       "autocomplete",
		Bounds:     inspect.Bounds{Width: m.width, Height: m.height},
		Label:      displayLabel(m.SelectedItem()),
		Focused:    m.focused,
		Selected:   selected,
		Scroll:     &inspect.Scroll{Total: len(m.matches), Visible: m.VisibleLines(), Offset: m.off},
		Actions:    m.Actions(),
		Attributes: attributes,
	}
}

func (m Model) nextEnabledPosition() int {
	if m.pos < 0 {
		return m.firstEnabledPosition()
	}
	for position := m.pos + 1; position < len(m.matches); position++ {
		if !m.items[m.matches[position]].Disabled {
			return position
		}
	}
	return -1
}

func (m Model) previousEnabledPosition() int {
	if m.pos < 0 {
		return m.lastEnabledPosition()
	}
	for position := m.pos - 1; position >= 0; position-- {
		if !m.items[m.matches[position]].Disabled {
			return position
		}
	}
	return -1
}

func displayLabel(item Item) string {
	if item.Label != "" {
		return item.Label
	}
	return item.Value
}
