package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ishan5ain/tuiweave/palette"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func sized(t *testing.T) model {
	t.Helper()
	m := newModel()
	next, _ := m.Update(tea.WindowSizeMsg{Width: 64, Height: 12})
	return next.(model)
}

func TestPaletteExampleGolden(t *testing.T) {
	m := sized(t)
	snaptest.Snap(t, m.render())
}

func TestPaletteExampleFiltersAndActivates(t *testing.T) {
	m := sized(t)
	for _, r := range "fo" {
		next, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = next.(model)
	}
	if got := m.commands.SelectedID(); got != "format" {
		t.Fatalf("filtered selection = %q, want format", got)
	}
	next, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = next.(model)
	if cmd == nil {
		t.Fatal("enter returned no command")
	}
	msg, ok := cmd().(palette.SelectedMsg)
	if !ok || msg.ID != "format" {
		t.Fatalf("activation message = %#v, want format", msg)
	}
}

func TestPaletteExampleScenarioGolden(t *testing.T) {
	result := snaptest.RunScenario(sized(t),
		snaptest.ScenarioStep{Name: "search format", Msg: tea.KeyPressMsg{Code: 'f', Text: "f"}},
		snaptest.ScenarioStep{Name: "search formatter", Msg: tea.KeyPressMsg{Code: 'o', Text: "o"}},
		snaptest.ScenarioStep{Name: "activate", Msg: tea.KeyPressMsg{Code: tea.KeyEnter}},
	)
	snaptest.SnapScenario(t, result)
}

func TestPaletteExamplePaste(t *testing.T) {
	m := sized(t)
	next, cmd := m.Update(tea.PasteMsg{Content: "format"})
	m = next.(model)
	if cmd != nil || m.commands.Query() != "format" || m.commands.SelectedID() != "format" {
		t.Fatalf("paste query=%q", m.commands.Query())
	}
	snaptest.Snap(t, m.render())
}
