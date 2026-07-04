package textarea

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/snaptest"
)

func newFocused(w, h int) Model {
	ta := New(gotui.Dark())
	ta.SetSize(w, h)
	ta.Focus()
	return ta
}

func typeString(m Model, s string) Model {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

func key(name string) tea.KeyPressMsg {
	switch name {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "backspace":
		return tea.KeyPressMsg{Code: tea.KeyBackspace}
	case "delete":
		return tea.KeyPressMsg{Code: tea.KeyDelete}
	case "left":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "home":
		return tea.KeyPressMsg{Code: tea.KeyHome}
	case "end":
		return tea.KeyPressMsg{Code: tea.KeyEnd}
	case "ctrl+k":
		return tea.KeyPressMsg{Code: 'k', Mod: tea.ModCtrl}
	case "ctrl+w":
		return tea.KeyPressMsg{Code: 'w', Mod: tea.ModCtrl}
	}
	panic("unsupported key in test helper: " + name)
}

func TestTypeAndValue(t *testing.T) {
	ta := newFocused(20, 3)
	ta = typeString(ta, "hello")
	ta, _ = ta.Update(key("enter"))
	ta = typeString(ta, "world")
	if got := ta.Value(); got != "hello\nworld" {
		t.Fatalf("Value() = %q", got)
	}
	snaptest.Snap(t, ta.View())
}

func TestSoftWrapGolden(t *testing.T) {
	ta := newFocused(12, 4) // wrap width 10
	ta = typeString(ta, "abcdefghij0123456789xyz")
	if got := ta.ContentHeight(); got != 3 {
		t.Fatalf("ContentHeight = %d, want 3 (23 runes at wrap 10)", got)
	}
	snaptest.Snap(t, ta.View())
}

func TestCursorAtExactWrapBoundary(t *testing.T) {
	ta := newFocused(12, 4) // wrap width 10
	ta = typeString(ta, "0123456789")
	// 10 runes = exact multiple: cursor needs a 3rd... 2nd visual row.
	if got := ta.ContentHeight(); got != 2 {
		t.Fatalf("ContentHeight = %d, want 2 (full row + cursor row)", got)
	}
	snaptest.Snap(t, ta.View())

	// Typing continues on the wrapped row.
	ta = typeString(ta, "X")
	if got := ta.Value(); got != "0123456789X" {
		t.Fatalf("Value() = %q", got)
	}
}

func TestPasteWithNewlines(t *testing.T) {
	ta := newFocused(20, 5)
	// bubbletea v2 delivers a paste as one KeyPressMsg.Text
	ta, _ = ta.Update(tea.KeyPressMsg{Text: "func main() {\n\tgo()\n}"})
	if got := ta.Value(); got != "func main() {\n\tgo()\n}" {
		t.Fatalf("Value() = %q", got)
	}
	if got := ta.ContentHeight(); got != 3 {
		t.Errorf("ContentHeight = %d, want 3", got)
	}
}

func TestBackspaceJoinsLines(t *testing.T) {
	ta := newFocused(20, 4)
	ta = typeString(ta, "ab")
	ta, _ = ta.Update(key("enter"))
	ta = typeString(ta, "cd")
	ta, _ = ta.Update(key("home"))
	ta, _ = ta.Update(key("backspace"))
	if got := ta.Value(); got != "abcd" {
		t.Fatalf("Value() = %q, want abcd", got)
	}
}

func TestDeleteJoinsForward(t *testing.T) {
	ta := newFocused(20, 4)
	ta = typeString(ta, "ab")
	ta, _ = ta.Update(key("enter"))
	ta = typeString(ta, "cd")
	ta, _ = ta.Update(key("up"))
	ta, _ = ta.Update(key("end"))
	ta, _ = ta.Update(key("delete"))
	if got := ta.Value(); got != "abcd" {
		t.Fatalf("Value() = %q, want abcd", got)
	}
}

func TestVerticalMovesByVisualRow(t *testing.T) {
	ta := newFocused(12, 5) // wrap width 10
	ta = typeString(ta, "0123456789ABCDE")
	// cursor at end (visual row 2, vcol 5); up should land inside row 1
	ta, _ = ta.Update(key("up"))
	ta = typeString(ta, "!")
	if got := ta.Value(); got != "01234!56789ABCDE" {
		t.Fatalf("visual up landed wrong: %q", got)
	}
}

func TestArrowsCrossLineBoundaries(t *testing.T) {
	ta := newFocused(20, 4)
	ta = typeString(ta, "ab")
	ta, _ = ta.Update(key("enter"))
	ta = typeString(ta, "cd")
	ta, _ = ta.Update(key("home"))
	ta, _ = ta.Update(key("left")) // wraps to end of line 1
	ta = typeString(ta, "X")
	if got := ta.Value(); got != "abX\ncd" {
		t.Fatalf("left across boundary: %q", got)
	}
	ta, _ = ta.Update(key("right")) // back to start of line 2
	ta = typeString(ta, "Y")
	if got := ta.Value(); got != "abX\nYcd" {
		t.Fatalf("right across boundary: %q", got)
	}
}

func TestCtrlKKillsAndJoins(t *testing.T) {
	ta := newFocused(20, 4)
	ta.SetValue("hello world\nnext")
	ta, _ = ta.Update(key("up"))
	ta, _ = ta.Update(key("home"))
	ta, _ = ta.Update(tea.KeyPressMsg{Code: 'k', Mod: tea.ModCtrl})
	if got := ta.Value(); got != "\nnext" {
		t.Fatalf("ctrl+k kill: %q", got)
	}
	ta, _ = ta.Update(tea.KeyPressMsg{Code: 'k', Mod: tea.ModCtrl})
	if got := ta.Value(); got != "next" {
		t.Fatalf("ctrl+k at EOL should join: %q", got)
	}
}

func TestCtrlWAcrossBoundary(t *testing.T) {
	ta := newFocused(20, 4)
	ta = typeString(ta, "one two")
	ta, _ = ta.Update(key("ctrl+w"))
	if got := ta.Value(); got != "one " {
		t.Fatalf("ctrl+w: %q", got)
	}
}

func TestGrowthAndScrollKeepsCursorVisible(t *testing.T) {
	ta := newFocused(20, 2)
	for i := range 5 {
		ta = typeString(ta, strings.Repeat("x", 3))
		if i < 4 {
			ta, _ = ta.Update(key("enter"))
		}
	}
	if got := ta.ContentHeight(); got != 5 {
		t.Fatalf("ContentHeight = %d, want 5", got)
	}
	// Box is 2 rows; cursor is on the last logical line and must be visible.
	view := ta.View()
	if got := len(strings.Split(view, "\n")); got != 2 {
		t.Fatalf("rendered %d rows, want 2", got)
	}
	snaptest.Snap(t, view)
}

func TestPlaceholderGolden(t *testing.T) {
	ta := New(gotui.Dark())
	ta.SetSize(24, 2)
	ta.Placeholder = "ask anything…"
	snaptest.SnapCells(t, ta.View(), snaptest.WithRoles(gotui.Dark()))
}

func TestBlurredIgnoresKeys(t *testing.T) {
	ta := New(gotui.Dark())
	ta.SetSize(20, 3)
	ta = typeString(ta, "ignored")
	if !ta.Empty() {
		t.Errorf("blurred textarea accepted text: %q", ta.Value())
	}
}

func TestResetAndInsertString(t *testing.T) {
	ta := newFocused(20, 3)
	ta.InsertString("a\nb")
	if got := ta.Value(); got != "a\nb" {
		t.Fatalf("InsertString: %q", got)
	}
	ta.Reset()
	if !ta.Empty() {
		t.Errorf("Reset left content: %q", ta.Value())
	}
}
