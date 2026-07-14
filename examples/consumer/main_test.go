package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ishan5ain/tuiweave"
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

func sized(t *testing.T) model {
	t.Helper()
	m := newModel()
	next, cmd := m.Update(tea.WindowSizeMsg{Width: 48, Height: 8})
	if cmd != nil {
		t.Fatal("window sizing emitted an unexpected command")
	}
	return next.(model)
}
