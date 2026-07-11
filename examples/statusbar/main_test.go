package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func TestStatusbarExampleCyclesPresets(t *testing.T) {
	m := sizedModel(t)
	if m.preset != 0 || !contains(m, "Dark") {
		t.Fatalf("initial preset = %d, want Dark", m.preset)
	}

	for i := 1; i < len(tuiweave.Presets()); i++ {
		m = updateModel(t, m, tea.KeyPressMsg{Code: 't', Text: "t"})
		if m.preset != i || !contains(m, tuiweave.Presets()[i].Name) {
			t.Fatalf("cycled preset = %d, want %d (%s)", m.preset, i, tuiweave.Presets()[i].Name)
		}
	}
	m = updateModel(t, m, tea.KeyPressMsg{Code: 't', Text: "t"})
	if m.preset != 0 || !contains(m, "Dark") {
		t.Fatalf("wrapped preset = %d, want Dark", m.preset)
	}
}

func TestStatusbarExampleGolden(t *testing.T) {
	m := sizedModel(t)
	t.Run("dark", func(t *testing.T) {
		snaptest.Snap(t, m.View().Content)
		snaptest.SnapCells(t, m.View().Content, snaptest.WithRoles(m.theme()))
	})

	m = updateModel(t, m, tea.KeyPressMsg{Code: 't', Text: "t"})
	t.Run("light", func(t *testing.T) {
		snaptest.Snap(t, m.View().Content)
		snaptest.SnapCells(t, m.View().Content, snaptest.WithRoles(m.theme()))
	})
}

func sizedModel(t *testing.T) model {
	t.Helper()
	m := newModel()
	return updateModel(t, m, tea.WindowSizeMsg{Width: 64, Height: 8})
}

func updateModel(t *testing.T, m model, msg tea.Msg) model {
	t.Helper()
	next, cmd := m.Update(msg)
	if cmd != nil {
		t.Fatalf("Update(%T) emitted unexpected command", msg)
	}
	return next.(model)
}

func contains(m model, text string) bool {
	return strings.Contains(ansi.Strip(m.View().Content), text)
}
