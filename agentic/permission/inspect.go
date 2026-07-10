package permission

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/ishansain/gotui/inspect"
)

const actionChoosePrefix = "choose."

// Actions reports one semantic action for each permission option.
func (m Model) Actions() []inspect.Action {
	actions := make([]inspect.Action, 0, len(m.options))
	for i, option := range m.options {
		actions = append(actions, inspect.Action{
			ID:          fmt.Sprintf("%s%d", actionChoosePrefix, i+1),
			Label:       option,
			Description: "Choose this permission option",
			Enabled:     true,
		})
	}
	return actions
}

func (m Model) applyAction(msg tea.Msg) (Model, tea.Cmd, bool) {
	action, ok := msg.(inspect.ActionMsg)
	if !ok || !strings.HasPrefix(action.ID, actionChoosePrefix) {
		return m, nil, false
	}
	choice, err := strconv.Atoi(strings.TrimPrefix(action.ID, actionChoosePrefix))
	if err != nil || choice < 1 || choice > len(m.options) {
		return m, nil, false
	}
	return m, m.result(choice - 1), true
}

// Inspect reports the pending approval, its provenance, and available choices.
func (m Model) Inspect() inspect.Node {
	p := m.provenance
	attributes := map[string]string{
		"body": m.Body,
	}
	for key, value := range map[string]string{
		"tool":          p.Tool,
		"operation":     p.Operation,
		"target":        p.Target,
		"scope":         p.Scope,
		"detail":        p.Detail,
		"impact":        p.Impact,
		"reversibility": p.Reversibility,
		"policy":        p.Policy,
	} {
		if value != "" {
			attributes[key] = value
		}
	}
	var selected *inspect.Selection
	if m.sel >= 0 && m.sel < len(m.options) {
		selected = &inspect.Selection{Index: m.sel, Count: len(m.options), Label: m.options[m.sel]}
	}
	return inspect.Node{
		ID:         m.ID,
		Kind:       "permission",
		Bounds:     inspect.Bounds{Width: m.width, Height: m.height},
		Label:      m.Title,
		Focused:    true,
		Status:     "awaiting_approval",
		Selected:   selected,
		Actions:    m.Actions(),
		Attributes: attributes,
	}
}
