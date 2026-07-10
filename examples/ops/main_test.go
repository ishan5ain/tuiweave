package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ishansain/gotui/inspect"
	"github.com/ishansain/gotui/palette"
	"github.com/ishansain/gotui/snaptest"
)

func sized(t *testing.T) model {
	t.Helper()
	m := newModel()
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 22})
	return next.(model)
}

func TestOpsScreenGolden(t *testing.T) {
	m := sized(t)
	snaptest.Snap(t, m.render())
}

func TestOpsTabAndFocusComposition(t *testing.T) {
	m := sized(t)
	if m.fm.Index() != 0 || !m.tabs.Focused() {
		t.Fatalf("initial focus index=%d tabs=%v", m.fm.Index(), m.tabs.Focused())
	}
	m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyRight})
	if m.tabs.SelectedID() != "jobs" || m.rows.SelectedRow()[0] != "nightly-backup" {
		t.Fatalf("tab selection: tab=%q row=%v", m.tabs.SelectedID(), m.rows.SelectedRow())
	}
	for i := 1; i <= 4; i++ {
		m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
		if m.fm.Index() != i {
			t.Fatalf("tab %d focus index = %d", i, m.fm.Index())
		}
	}
	if !m.openLogs.Focused() {
		t.Fatal("focus cycle did not reach open logs button")
	}
}

func TestOpsCommandPaletteGolden(t *testing.T) {
	m := sized(t)
	m, _ = update(t, m, tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})
	if !m.showPalette {
		t.Fatal("ctrl+p did not open command palette")
	}
	snaptest.Snap(t, m.render())
}

func TestOpsCommandPaletteActivation(t *testing.T) {
	m := sized(t)
	m, _ = update(t, m, tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})
	for _, r := range "restart" {
		m, _ = update(t, m, tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	m, cmd := update(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("palette enter returned no command")
	}
	if m.showPalette == false {
		// The command is intentionally not executed by this test; palette
		// visibility closes when the resulting message is delivered below.
		t.Fatal("palette closed before SelectedMsg was delivered")
	}
	selected, ok := cmd().(palette.SelectedMsg)
	if !ok || selected.ID != "restart" {
		t.Fatalf("palette message = %#v, want restart", selected)
	}
	next, _ := m.Update(selected)
	m = next.(model)
	if m.showPalette || m.notice != "ran Restart service" {
		t.Fatalf("palette result: open=%v notice=%q", m.showPalette, m.notice)
	}
}

func TestOpsInspectionGolden(t *testing.T) {
	m := sized(t)
	m, _ = update(t, m, tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})
	node := m.Inspect()
	data, err := inspect.Marshal(node)
	if err != nil {
		t.Fatal(err)
	}
	snaptest.Snap(t, string(data))
}

func TestOpsScenarioGolden(t *testing.T) {
	result := snaptest.RunScenario(sized(t),
		snaptest.ScenarioStep{Name: "focus actions", Msg: tea.KeyPressMsg{Code: tea.KeyTab}},
		snaptest.ScenarioStep{Name: "navigate action", Msg: tea.KeyPressMsg{Code: tea.KeyDown}},
		snaptest.ScenarioStep{Name: "open commands", Msg: tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl}},
	)
	snaptest.SnapScenario(t, result)
}

func update(t *testing.T, m model, msg tea.Msg) (model, tea.Cmd) {
	t.Helper()
	next, cmd := m.Update(msg)
	return next.(model), cmd
}
