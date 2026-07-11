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
	next, _ := m.Update(tea.WindowSizeMsg{Width: 72, Height: 22})
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

// answer sends a key to the open permission prompt and delivers the
// ResultMsg its command produces, as the runtime would.
func answer(t *testing.T, m model, key tea.KeyPressMsg) model {
	t.Helper()
	next, cmd := m.Update(key)
	m = next.(model)
	if cmd == nil {
		t.Fatal("permission prompt produced no result command")
	}
	return send(t, m, cmd())
}

// runUntil ticks the state machine until pred holds (or fails the test).
func runUntil(t *testing.T, m model, pred func(model) bool) model {
	t.Helper()
	for range 200 {
		if pred(m) {
			return m
		}
		m = send(t, m, tickMsg{})
	}
	t.Fatal("state machine did not reach expected state in 200 ticks")
	return m
}

// startSession types a question and presses enter.
func startSession(t *testing.T, m model) model {
	t.Helper()
	for _, r := range "how does layout work?" {
		m = send(t, m, tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return send(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
}

func TestMultilineInputGrowsAndSends(t *testing.T) {
	m := sized(t)
	for _, r := range "line one" {
		m = send(t, m, tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	m = send(t, m, tea.KeyPressMsg{Code: tea.KeyEnter, Mod: tea.ModAlt})
	for _, r := range "line two" {
		m = send(t, m, tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	if got := m.input.Value(); got != "line one\nline two" {
		t.Fatalf("input value = %q", got)
	}
	if got := m.input.ContentHeight(); got != 2 {
		t.Fatalf("ContentHeight = %d, want 2", got)
	}
	// The frame must stay the terminal height: transcript shrank instead.
	if got := len(strings.Split(m.render(), "\n")); got != 22 {
		t.Fatalf("frame is %d rows, want 22", got)
	}

	m = send(t, m, tea.KeyPressMsg{Code: tea.KeyEnter}) // enter sends
	if m.step != stepStreamingIntro {
		t.Fatal("enter did not start the session")
	}
	if !m.input.Empty() {
		t.Errorf("input not cleared after send: %q", m.input.Value())
	}
}

func TestSessionReachesPermissionGolden(t *testing.T) {
	m := startSession(t, sized(t))
	m = runUntil(t, m, func(m model) bool { return m.step == stepAwaitPermission })
	if !m.showPerm {
		t.Fatal("permission overlay not shown at stepAwaitPermission")
	}
	snaptest.Snap(t, m.render())
}

func TestFullSessionGolden(t *testing.T) {
	m := startSession(t, sized(t))
	m = runUntil(t, m, func(m model) bool { return m.step == stepAwaitPermission })
	m = answer(t, m, tea.KeyPressMsg{Code: '1', Text: "1"}) // allow once
	m = runUntil(t, m, func(m model) bool { return m.step == stepIdle })

	// The visible window follows the bottom; assert against the full
	// transcript by rendering every cell.
	var full strings.Builder
	for _, cell := range m.transcript.Cells() {
		full.WriteString(ansi.Strip(cell.Render(72)) + "\n")
	}
	for _, want := range []string{"❯ you", "✦ assistant", "✓ Bash", "+"} {
		if !strings.Contains(full.String(), want) {
			t.Errorf("transcript missing %q", want)
		}
	}
	if !strings.Contains(ansi.Strip(m.render()), "ctx") {
		t.Error("usage bar missing from final screen")
	}
	snaptest.Snap(t, m.render())
}

func TestDenyStopsSession(t *testing.T) {
	m := startSession(t, sized(t))
	m = runUntil(t, m, func(m model) bool { return m.step == stepAwaitPermission })
	m = answer(t, m, tea.KeyPressMsg{Code: tea.KeyEscape}) // safe default = Deny
	if m.step != stepIdle {
		t.Fatalf("after deny: step = %d, want stepIdle", m.step)
	}
	if !strings.Contains(ansi.Strip(m.render()), "denied") {
		t.Error("transcript missing denial note")
	}
}
