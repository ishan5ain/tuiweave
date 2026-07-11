package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ishansain/gotui/autocomplete"
	"github.com/ishansain/gotui/snaptest"
)

func sized(t *testing.T) model {
	t.Helper()
	m := newModel()
	next, _ := m.Update(tea.WindowSizeMsg{Width: 64, Height: 12})
	return next.(model)
}

func TestAutocompleteExampleGolden(t *testing.T) {
	m := sized(t)
	snaptest.Snap(t, m.render())
}

func TestAutocompleteExampleFiltersAndAccepts(t *testing.T) {
	m := sized(t)
	for _, r := range "git" {
		next, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = next.(model)
	}
	if got := m.suggestions.SelectedID(); got != "git-checkout" {
		t.Fatalf("filtered selection = %q, want git-checkout", got)
	}
	next, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = next.(model)
	if got := m.suggestions.SelectedID(); got != "git-status" {
		t.Fatalf("after down selection = %q, want git-status", got)
	}
	if cmd != nil {
		t.Fatal("down returned an unexpected command")
	}
	next, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = next.(model)
	if cmd == nil {
		t.Fatal("enter returned no command")
	}
	msg, ok := cmd().(autocomplete.SelectedMsg)
	if !ok || msg.ID != "git-status" || msg.Value != "git status" {
		t.Fatalf("acceptance message = %#v, want git-status/git status", msg)
	}
}

func TestAutocompleteExampleScenarioGolden(t *testing.T) {
	result := snaptest.RunScenario(sized(t),
		snaptest.ScenarioStep{Name: "type git", Msg: tea.KeyPressMsg{Code: 'g', Text: "g"}},
		snaptest.ScenarioStep{Name: "type git status", Msg: tea.KeyPressMsg{Code: 'i', Text: "i"}},
		snaptest.ScenarioStep{Name: "choose next", Msg: tea.KeyPressMsg{Code: tea.KeyDown}},
		snaptest.ScenarioStep{Name: "accept", Msg: tea.KeyPressMsg{Code: tea.KeyEnter}},
	)
	snaptest.SnapScenario(t, result)
}
