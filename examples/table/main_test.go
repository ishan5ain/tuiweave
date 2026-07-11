package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave/snaptest"
)

func sized(t *testing.T) model {
	t.Helper()
	m := newModel()
	next, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 20})
	sm, ok := next.(model)
	if !ok {
		t.Fatalf("Update returned %T, want model", next)
	}
	return sm
}

func send(t *testing.T, m model, msg tea.Msg) model {
	t.Helper()
	next, _ := m.Update(msg)
	return next.(model)
}

func TestScreenGolden(t *testing.T) {
	m := sized(t)
	snaptest.Snap(t, m.render())
}

func TestSelectionDrivesDiff(t *testing.T) {
	m := sized(t)
	m = send(t, m, tea.KeyPressMsg{Code: 'j', Text: "j"})
	plain := ansi.Strip(m.render())
	if !strings.Contains(plain, "pkg02") {
		t.Errorf("diff pane not synced to selected row:\n%s", plain)
	}
	snaptest.Snap(t, m.render())
}

func TestScrollbarsPresent(t *testing.T) {
	m := sized(t)
	plain := ansi.Strip(m.render())
	if !strings.Contains(plain, "┃") {
		t.Error("no scrollbar thumb rendered (24 rows in an 18-row pane must scroll)")
	}
	if !strings.Contains(plain, "│") {
		t.Error("no scrollbar track rendered")
	}
}

func TestTabMovesFocusToDiff(t *testing.T) {
	m := sized(t)
	// Select the last row first — its diff (~78 lines) overflows the pane,
	// so it can actually scroll.
	m = send(t, m, tea.KeyPressMsg{Code: 'G', Text: "G"})
	sel := m.files.Selected()

	m = send(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	before := m.diff.YOffset()
	m = send(t, m, tea.KeyPressMsg{Code: 'j', Text: "j"})
	if m.diff.YOffset() != before+1 {
		t.Error("focused diff pane did not scroll on j")
	}
	if got := m.files.Selected(); got != sel {
		t.Errorf("blurred table moved selection from %d to %d", sel, got)
	}
}
