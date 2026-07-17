package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/inspect"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func TestConsumerPublicAPIView(t *testing.T) {
	m := sized(t)
	snaptest.Snap(t, m.render())
	snaptest.SnapCells(t, m.render(), snaptest.WithRoles(tuiweave.Dark()))
}

func TestConsumerPublicAPIInteraction(t *testing.T) {
	m := sized(t)
	next, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if cmd != nil {
		t.Fatal("navigation emitted an unexpected command")
	}
	m = next.(model)
	if got := m.items.SelectedItem(); got != "focus" {
		t.Fatalf("selected item = %q, want focus", got)
	}
}

func TestConsumerInspectionGolden(t *testing.T) {
	m := sized(t)
	data, err := inspect.Marshal(m.Inspect())
	if err != nil {
		t.Fatal(err)
	}
	snaptest.Snap(t, string(data))
}

func TestConsumerQualifiedSemanticActionRouting(t *testing.T) {
	m := sized(t)
	before := m.items.Selected()
	next, cmd := m.Update(inspect.Invoke("items.previous"))
	if cmd != nil {
		t.Fatal("disabled previous action emitted a command")
	}
	m = next.(model)
	if m.items.Selected() != before {
		t.Fatalf("disabled previous action changed selection to %d", m.items.Selected())
	}

	next, cmd = m.Update(inspect.Invoke("items.next"))
	if cmd != nil {
		t.Fatal("list next action emitted a command")
	}
	m = next.(model)
	if m.items.SelectedItem() != "focus" {
		t.Fatalf("qualified next selected %q", m.items.SelectedItem())
	}

	data, err := inspect.Marshal(m.Inspect())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"id": "items.next"`) {
		t.Fatalf("updated inspection missing qualified next action:\n%s", data)
	}
}

func sized(t *testing.T) model {
	t.Helper()
	m := newModel()
	next, cmd := m.Update(tea.WindowSizeMsg{Width: 48, Height: 8})
	if cmd != nil {
		t.Fatal("window sizing emitted an unexpected command")
	}
	return next.(model)
}
