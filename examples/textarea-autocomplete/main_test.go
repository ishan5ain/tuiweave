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
	next, _ := m.Update(tea.WindowSizeMsg{Width: 72, Height: 14})
	return next.(model)
}

func typeText(t *testing.T, m model, text string) model {
	t.Helper()
	for _, r := range text {
		next, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = next.(model)
	}
	return m
}

func TestTextareaAutocompleteExampleGolden(t *testing.T) {
	m := sized(t)
	snaptest.Snap(t, m.render())
}

func TestTextareaAutocompleteReplacesLogicalToken(t *testing.T) {
	m := typeText(t, sized(t), "run git")
	if got := m.input.CursorPosition(); got.Column != 7 {
		t.Fatalf("cursor = %#v, want column 7", got)
	}
	if got := m.suggestions.Query(); got != "git" {
		t.Fatalf("query = %q, want git", got)
	}

	next, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = next.(model)
	if cmd == nil {
		t.Fatal("enter returned no completion command")
	}
	selected, ok := cmd().(autocomplete.SelectedMsg)
	if !ok || selected.ID != "git-status" {
		t.Fatalf("completion = %#v, want git-status", selected)
	}
	next, cmd = m.Update(selected)
	if cmd != nil {
		t.Fatal("completion delivery returned an unexpected command")
	}
	m = next.(model)
	if got := m.input.Value(); got != "run git status" {
		t.Fatalf("completed value = %q, want %q", got, "run git status")
	}
	if got := m.input.CursorPosition(); got.Column != len([]rune("run git status")) {
		t.Fatalf("cursor after completion = %#v", got)
	}
}

func TestTextareaAutocompleteTracksMiddleCursor(t *testing.T) {
	m := sized(t)
	m.input.SetValue("git tail")
	m.syncQuery()
	for range 5 {
		next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
		m = next.(model)
	}
	if got := m.input.CursorPosition(); got.Column != 3 {
		t.Fatalf("middle cursor = %#v, want column 3", got)
	}
	if got := m.suggestions.Query(); got != "git" {
		t.Fatalf("middle query = %q, want git", got)
	}
	m.suggestions.SelectID("git-status")

	next, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = next.(model)
	if cmd == nil {
		t.Fatal("middle completion returned no command")
	}
	selected := cmd().(autocomplete.SelectedMsg)
	next, _ = m.Update(selected)
	m = next.(model)
	if got := m.input.Value(); got != "git status tail" {
		t.Fatalf("middle completed value = %q, want %q", got, "git status tail")
	}
}

func TestTextareaAutocompleteScenarioGolden(t *testing.T) {
	m := sized(t)
	result := snaptest.RunScenario(m,
		snaptest.ScenarioStep{Name: "type command fragment", Msg: tea.KeyPressMsg{Code: 'g', Text: "g"}},
		snaptest.ScenarioStep{Name: "type command prefix", Msg: tea.KeyPressMsg{Code: 'i', Text: "i"}},
		snaptest.ScenarioStep{Name: "accept completion", Msg: tea.KeyPressMsg{Code: tea.KeyEnter}},
		snaptest.ScenarioStep{Name: "deliver completion", Msg: autocomplete.SelectedMsg{ID: "git-status", Value: "git status", Label: "git status"}},
	)
	snaptest.SnapScenario(t, result)
}
