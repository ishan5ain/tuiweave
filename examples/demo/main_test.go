package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ishansain/gotui/snaptest"
)

// sized returns the demo model laid out at 80×24, as after a WindowSizeMsg.
func sized(t *testing.T) model {
	t.Helper()
	m := newModel()
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	sm, ok := next.(model)
	if !ok {
		t.Fatalf("Update returned %T, want model", next)
	}
	return sm
}

func TestDemoScreenGolden(t *testing.T) {
	m := sized(t)
	snaptest.Snap(t, m.render())
}

func TestDemoDialogOverlayGolden(t *testing.T) {
	m := sized(t)
	next, _ := m.Update(tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl})
	m = next.(model)
	if !m.showDialog {
		t.Fatal("ctrl+d did not open the dialog")
	}
	snaptest.Snap(t, m.render())
}

func TestDemoFocusCycleReachesInput(t *testing.T) {
	m := sized(t)
	for range 2 {
		next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		m = next.(model)
	}
	next, _ := m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	m = next.(model)
	if got := m.input.Value(); got != "x" {
		t.Fatalf("after tab·tab + x: input value = %q, want %q", got, "x")
	}
	if !strings.Contains(m.render(), " input ") {
		t.Error("statusbar does not show input as the focused pane")
	}
}

func TestDemoAddItemViaEnter(t *testing.T) {
	m := sized(t)
	for range 2 {
		next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		m = next.(model)
	}
	for _, r := range "docs" {
		next, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = next.(model)
	}
	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = next.(model)

	if got := m.list.Len(); got != 5 {
		t.Fatalf("list has %d items after add, want 5", got)
	}
	if got := m.list.SelectedItem(); got != "docs" {
		t.Fatalf("selected item = %q, want %q", got, "docs")
	}
	if got := m.input.Value(); got != "" {
		t.Fatalf("input not cleared after enter: %q", got)
	}
}
