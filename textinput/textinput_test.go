package textinput

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/snaptest"
)

func typeString(m Model, s string) Model {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

func key(name string) tea.KeyPressMsg {
	switch name {
	case "backspace":
		return tea.KeyPressMsg{Code: tea.KeyBackspace}
	case "left":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "home":
		return tea.KeyPressMsg{Code: tea.KeyHome}
	case "ctrl+w":
		return tea.KeyPressMsg{Code: 'w', Mod: tea.ModCtrl}
	case "ctrl+u":
		return tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl}
	}
	panic("unsupported key in test helper: " + name)
}

func newFocused(width int) Model {
	ti := New(gotui.Dark())
	ti.SetSize(width, 1)
	ti.Focus()
	return ti
}

func TestTypingAndValue(t *testing.T) {
	ti := newFocused(30)
	ti = typeString(ti, "hello world")
	if got := ti.Value(); got != "hello world" {
		t.Fatalf("Value() = %q, want %q", got, "hello world")
	}
	snaptest.SnapCells(t, ti.View(), snaptest.WithRoles(gotui.Dark()))
}

func TestEditingKeys(t *testing.T) {
	ti := newFocused(30)
	ti = typeString(ti, "abc def")

	ti, _ = ti.Update(key("backspace"))
	if got := ti.Value(); got != "abc de" {
		t.Fatalf("after backspace: %q", got)
	}
	ti, _ = ti.Update(key("ctrl+w"))
	if got := ti.Value(); got != "abc " {
		t.Fatalf("after ctrl+w: %q", got)
	}
	ti, _ = ti.Update(key("ctrl+u"))
	if got := ti.Value(); got != "" {
		t.Fatalf("after ctrl+u: %q", got)
	}
}

func TestCursorMidText(t *testing.T) {
	ti := newFocused(30)
	ti = typeString(ti, "abcd")
	ti, _ = ti.Update(key("left"))
	ti, _ = ti.Update(key("left"))
	ti = typeString(ti, "X")
	if got := ti.Value(); got != "abXcd" {
		t.Fatalf("mid-text insert: %q", got)
	}
}

func TestBlurredIgnoresKeys(t *testing.T) {
	ti := New(gotui.Dark())
	ti.SetSize(30, 1)
	ti = typeString(ti, "ignored")
	if got := ti.Value(); got != "" {
		t.Errorf("blurred input accepted text: %q", got)
	}
}

func TestPlaceholderGolden(t *testing.T) {
	ti := New(gotui.Dark())
	ti.SetSize(30, 1)
	ti.Placeholder = "type a message…"
	snaptest.SnapCells(t, ti.View(), snaptest.WithRoles(gotui.Dark()))
}

func TestHorizontalScrollShowsCursor(t *testing.T) {
	ti := newFocused(12) // prompt "> " leaves 10 columns, 9 visible
	ti = typeString(ti, "0123456789ABCDEF")
	snaptest.Snap(t, ti.View())
}

func TestHomeShowsStart(t *testing.T) {
	ti := newFocused(12)
	ti = typeString(ti, "0123456789ABCDEF")
	ti, _ = ti.Update(key("home"))
	snaptest.Snap(t, ti.View())
}

func TestCellWidthRendering(t *testing.T) {
	ti := newFocused(8)
	ti.SetValue("界abcd")
	ti.pos = 0
	if got := ansi.StringWidth(ansi.Strip(ti.View())); got > ti.width {
		t.Fatalf("wide cursor view width = %d, want <= %d", got, ti.width)
	}

	ti.SetSize(3, 1)
	ti.SetValue("界")
	ti.pos = 0
	if got := ansi.StringWidth(ansi.Strip(ti.View())); got != ti.width {
		t.Fatalf("narrow wide-rune view width = %d, want %d", got, ti.width)
	}

	ti.SetSize(8, 1)
	ti.Prompt = "界"
	ti.SetValue("e\u0301abcd")
	ti.pos = 1 // inside the combining grapheme; the whole cluster stays styled
	if got := ansi.StringWidth(ansi.Strip(ti.View())); got > ti.width {
		t.Fatalf("combining cursor view width = %d, want <= %d", got, ti.width)
	}
	snaptest.Snap(t, ti.View())
}
