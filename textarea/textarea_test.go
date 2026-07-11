package textarea

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/inspect"
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
	case "ctrl+delete":
		return tea.KeyPressMsg{Code: tea.KeyDelete, Mod: tea.ModCtrl}
	case "alt+backspace":
		return tea.KeyPressMsg{Code: tea.KeyBackspace, Mod: tea.ModAlt}
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
	case "shift+left":
		return tea.KeyPressMsg{Code: tea.KeyLeft, Mod: tea.ModShift}
	case "shift+up":
		return tea.KeyPressMsg{Code: tea.KeyUp, Mod: tea.ModShift}
	case "ctrl+left":
		return tea.KeyPressMsg{Code: tea.KeyLeft, Mod: tea.ModCtrl}
	case "ctrl+shift+left":
		return tea.KeyPressMsg{Code: tea.KeyLeft, Mod: tea.ModCtrl | tea.ModShift}
	case "ctrl+right":
		return tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModCtrl}
	case "ctrl+shift+right":
		return tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModCtrl | tea.ModShift}
	case "alt+y":
		return tea.KeyPressMsg{Code: 'y', Mod: tea.ModAlt}
	case "ctrl+a":
		return tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl}
	case "ctrl+z":
		return tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl}
	case "ctrl+y":
		return tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl}
	case "end":
		return tea.KeyPressMsg{Code: tea.KeyEnd}
	case "ctrl+k":
		return tea.KeyPressMsg{Code: 'k', Mod: tea.ModCtrl}
	case "ctrl+w":
		return tea.KeyPressMsg{Code: 'w', Mod: tea.ModCtrl}
	case "ctrl+u":
		return tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl}
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

func TestCellWidthWrapAndCursor(t *testing.T) {
	ta := newFocused(8, 4) // cell wrap width 6 after the prompt
	ta.SetValue("界abcd")
	if got := ta.ContentHeight(); got != 2 {
		t.Fatalf("wide ContentHeight = %d, want 2 for an exact cell-width row", got)
	}
	rows := ta.visualRows()
	if len(rows) != 2 || rows[0].width != 6 || rows[1].width != 0 {
		t.Fatalf("wide visual rows = %#v, want widths 6, 0", rows)
	}

	ta.SetValue("e\u0301abcd")
	if got := ta.ContentHeight(); got != 1 {
		t.Fatalf("combining ContentHeight = %d, want 1", got)
	}
	for _, line := range strings.Split(ta.View(), "\n") {
		if got := ansi.StringWidth(line); got != 8 {
			t.Fatalf("rendered cell width = %d, want 8 for %q", got, line)
		}
	}
	ta.SetValue("界abc\nabcd")
	ta, _ = ta.Update(key("up"))
	if ta.row != 0 || ta.col != 3 {
		t.Fatalf("vertical cell mapping = (%d,%d), want (0,3)", ta.row, ta.col)
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

func TestSelectionAcrossLinesAndReplacement(t *testing.T) {
	ta := newFocused(20, 4)
	ta.SetValue("hello\nworld")
	ta, _ = ta.Update(key("shift+left"))
	ta, _ = ta.Update(key("shift+up"))
	if !ta.HasSelection() || ta.SelectedText() != "o\nworld" {
		t.Fatalf("selection = %q, has=%v; want o\\nworld", ta.SelectedText(), ta.HasSelection())
	}
	snaptest.Snap(t, ta.View())
	snaptest.SnapCells(t, ta.View(), snaptest.WithRoles(gotui.Dark()))

	ta, _ = ta.Update(tea.KeyPressMsg{Text: "there"})
	if got := ta.Value(); got != "hellthere" {
		t.Fatalf("replacement = %q, want hellthere", got)
	}
	if ta.HasSelection() {
		t.Fatal("replacement left a selection active")
	}
}

func TestSelectAllUndoRedo(t *testing.T) {
	ta := newFocused(20, 4)
	ta.SetValue("hello world")
	ta, _ = ta.Update(key("ctrl+a"))
	if ta.SelectedText() != "hello world" {
		t.Fatalf("select all = %q", ta.SelectedText())
	}
	ta, _ = ta.Update(tea.KeyPressMsg{Text: "replaced"})
	if got := ta.Value(); got != "replaced" || !ta.CanUndo() || ta.CanRedo() {
		t.Fatalf("after replacement: value=%q undo=%v redo=%v", got, ta.CanUndo(), ta.CanRedo())
	}
	ta, _ = ta.Update(key("ctrl+z"))
	if got := ta.Value(); got != "hello world" || !ta.CanRedo() {
		t.Fatalf("after undo: value=%q redo=%v", got, ta.CanRedo())
	}
	ta, _ = ta.Update(key("ctrl+y"))
	if got := ta.Value(); got != "replaced" || !ta.CanUndo() || ta.CanRedo() {
		t.Fatalf("after redo: value=%q undo=%v redo=%v", got, ta.CanUndo(), ta.CanRedo())
	}
	ta.InsertString("!")
	if ta.CanRedo() {
		t.Fatal("new edit did not clear redo history")
	}
}

func TestSelectionCollapsesBeforeTyping(t *testing.T) {
	ta := newFocused(20, 3)
	ta.SetValue("abc")
	ta, _ = ta.Update(key("ctrl+a"))
	ta, _ = ta.Update(key("left"))
	ta.InsertString("X")
	if got := ta.Value(); got != "Xabc" {
		t.Fatalf("left collapsed selection to %q, want Xabc", got)
	}

	ta.SetValue("abc")
	ta, _ = ta.Update(key("ctrl+a"))
	ta, _ = ta.Update(key("right"))
	ta.InsertString("X")
	if got := ta.Value(); got != "abcX" {
		t.Fatalf("right collapsed selection to %q, want abcX", got)
	}
}

func TestWordMovementAndSelection(t *testing.T) {
	ta := newFocused(24, 3)
	ta.SetValue("one two three")
	ta, _ = ta.Update(key("ctrl+left"))
	if ta.row != 0 || ta.col != 8 {
		t.Fatalf("ctrl+left cursor = (%d,%d), want (0,8)", ta.row, ta.col)
	}
	ta, _ = ta.Update(key("ctrl+shift+left"))
	if ta.SelectedText() != "two " {
		t.Fatalf("word selection = %q, want %q", ta.SelectedText(), "two ")
	}

	ta.SetValue("one two\nthree")
	ta.row, ta.col = 0, 0
	ta, _ = ta.Update(key("ctrl+right"))
	if ta.row != 0 || ta.col != 4 {
		t.Fatalf("first ctrl+right cursor = (%d,%d), want (0,4)", ta.row, ta.col)
	}
	ta, _ = ta.Update(key("ctrl+right"))
	if ta.row != 1 || ta.col != 0 {
		t.Fatalf("cross-line ctrl+right cursor = (%d,%d), want (1,0)", ta.row, ta.col)
	}
	ta, _ = ta.Update(key("ctrl+shift+right"))
	if ta.SelectedText() != "three" {
		t.Fatalf("right word selection = %q, want %q", ta.SelectedText(), "three")
	}
}

func TestKillRingAndYank(t *testing.T) {
	ta := newFocused(24, 3)
	ta.SetValue("one two")
	previous := ta
	next, _ := ta.Update(key("ctrl+w"))
	if previous.Value() != "one two" || previous.CanYank() {
		t.Fatalf("prior MVU model changed after kill: value=%q canYank=%v", previous.Value(), previous.CanYank())
	}
	ta = next
	if got := ta.Value(); got != "one " || !ta.CanYank() {
		t.Fatalf("kill = %q, canYank=%v; want one space and true", got, ta.CanYank())
	}
	ta, _ = ta.Update(key("alt+y"))
	if got := ta.Value(); got != "one two" {
		t.Fatalf("yank = %q, want one two", got)
	}
	ta, _ = ta.Update(key("ctrl+z"))
	if got := ta.Value(); got != "one " {
		t.Fatalf("undo yank = %q, want one space", got)
	}
	ta, _ = ta.Update(inspect.Invoke(ActionYank))
	if got := ta.Value(); got != "one two" {
		t.Fatalf("semantic yank = %q, want one two", got)
	}

	ta.SetValue("one two")
	if ta.CanYank() {
		t.Fatal("SetValue did not clear the kill ring")
	}
	ta, _ = ta.Update(key("ctrl+u"))
	if got := ta.Value(); got != "" || !ta.CanYank() {
		t.Fatalf("line kill = %q, canYank=%v; want empty value and true", got, ta.CanYank())
	}
	ta, _ = ta.Update(key("alt+y"))
	if got := ta.Value(); got != "one two" {
		t.Fatalf("line yank = %q, want one two", got)
	}

	ta.SetValue("one two\nthree")
	ta, _ = ta.Update(key("home"))
	ta, _ = ta.Update(key("ctrl+k"))
	if got := ta.Value(); got != "one two\n" || !ta.CanYank() {
		t.Fatalf("line-end kill = %q, canYank=%v; want trailing newline and true", got, ta.CanYank())
	}
	ta, _ = ta.Update(key("alt+y"))
	if got := ta.Value(); got != "one two\nthree" {
		t.Fatalf("line-end yank = %q, want original value", got)
	}
}

func TestKillRingCoalescesAndRotates(t *testing.T) {
	ta := newFocused(24, 3)
	ta.SetValue("one two three")
	ta, _ = ta.Update(key("ctrl+w"))
	ta, _ = ta.Update(key("ctrl+w"))
	if got := ta.killRing[0]; got != "two three" {
		t.Fatalf("backward kill coalescing = %q, want %q", got, "two three")
	}

	ta.SetValue("ab\ncd")
	ta.row, ta.col = 0, 1
	ta, _ = ta.Update(key("ctrl+k"))
	ta, _ = ta.Update(key("ctrl+k"))
	if got := ta.killRing[0]; got != "b\n" {
		t.Fatalf("forward kill coalescing = %q, want %q", got, "b\n")
	}

	ta.SetValue("one two")
	ta, _ = ta.Update(key("ctrl+w"))
	ta, _ = ta.Update(key("left")) // break the consecutive-kill chain
	ta, _ = ta.Update(key("ctrl+w"))
	if len(ta.killRing) != 2 || ta.killRing[0] != "one" || ta.killRing[1] != "two" {
		t.Fatalf("kill ring = %#v, want [one two]", ta.killRing)
	}
	ta, _ = ta.Update(key("alt+y"))
	if got := ta.Value(); got != "one " {
		t.Fatalf("first yank = %q, want %q", got, "one ")
	}
	ta, _ = ta.Update(key("alt+y"))
	if got := ta.Value(); got != "two " {
		t.Fatalf("rotated yank = %q, want %q", got, "two ")
	}
}

func TestRicherWordDeletionCommands(t *testing.T) {
	ta := newFocused(24, 3)
	ta.SetValue("one two three")
	ta.row, ta.col = 0, 0
	ta, _ = ta.Update(key("ctrl+delete"))
	if got := ta.Value(); got != "two three" || ta.killRing[0] != "one " {
		t.Fatalf("forward word kill = %q, ring=%q; want two three and one space", got, ta.killRing[0])
	}
	ta, _ = ta.Update(key("alt+y"))
	if got := ta.Value(); got != "one two three" {
		t.Fatalf("forward word yank = %q, want original value", got)
	}

	ta.SetValue("one two")
	ta, _ = ta.Update(key("alt+backspace"))
	if got := ta.Value(); got != "one " || ta.killRing[0] != "two" {
		t.Fatalf("alt+backspace = %q, ring=%q; want one space and two", got, ta.killRing[0])
	}
	ta, _ = ta.Update(key("ctrl+z"))
	if got := ta.Value(); got != "one two" {
		t.Fatalf("undo word kill = %q, want original value", got)
	}
}

func TestTextareaSemanticEditingActions(t *testing.T) {
	ta := newFocused(20, 3)
	ta.SetValue("abc")
	next, handled := ta.applyAction(inspect.Invoke(ActionSelectAll))
	if !handled || !next.HasSelection() {
		t.Fatalf("select-all action: handled=%v selection=%v", handled, next.HasSelection())
	}
	next, handled = next.applyAction(inspect.Invoke(ActionClearSelection))
	if !handled || next.HasSelection() {
		t.Fatalf("clear-selection action: handled=%v selection=%v", handled, next.HasSelection())
	}
	next.InsertString("d")
	next, handled = next.applyAction(inspect.Invoke(ActionUndo))
	if !handled || next.Value() != "abc" {
		t.Fatalf("undo action: handled=%v value=%q", handled, next.Value())
	}
	next, handled = next.applyAction(inspect.Invoke(ActionRedo))
	if !handled || next.Value() != "abcd" {
		t.Fatalf("redo action: handled=%v value=%q", handled, next.Value())
	}
	if got := next.Inspect().Attributes["selected_runes"]; got != "0" {
		t.Fatalf("selected_runes = %q, want 0", got)
	}
}

func TestTextareaInspectionGolden(t *testing.T) {
	ta := newFocused(20, 3)
	ta.SetValue("abc")
	ta.SelectAll()
	data, err := inspect.Marshal(ta.Inspect())
	if err != nil {
		t.Fatal(err)
	}
	snaptest.Snap(t, string(data))
}

func TestTextareaSelectionScenarioGolden(t *testing.T) {
	m := scenarioModel{textarea: newFocused(20, 4)}
	m.textarea.SetValue("hello world")
	result := snaptest.RunScenario(m,
		snaptest.ScenarioStep{Name: "select all", Msg: key("ctrl+a")},
		snaptest.ScenarioStep{Name: "replace selection", Msg: tea.KeyPressMsg{Text: "hi"}},
		snaptest.ScenarioStep{Name: "undo", Msg: key("ctrl+z")},
	)
	snaptest.SnapScenario(t, result)
}

func TestTextareaEditingDepthScenarioGolden(t *testing.T) {
	m := scenarioModel{textarea: newFocused(20, 3)}
	m.textarea.SetValue("one two")
	result := snaptest.RunScenario(m,
		snaptest.ScenarioStep{Name: "kill previous word", Msg: key("ctrl+w")},
		snaptest.ScenarioStep{Name: "break kill chain", Msg: key("left")},
		snaptest.ScenarioStep{Name: "kill another word", Msg: key("ctrl+w")},
		snaptest.ScenarioStep{Name: "yank", Msg: key("alt+y")},
		snaptest.ScenarioStep{Name: "rotate yank", Msg: key("alt+y")},
		snaptest.ScenarioStep{Name: "undo yank", Msg: key("ctrl+z")},
	)
	snaptest.SnapScenario(t, result)
}

func TestTextareaWordDeletionScenarioGolden(t *testing.T) {
	m := scenarioModel{textarea: newFocused(20, 3)}
	m.textarea.SetValue("one two")
	result := snaptest.RunScenario(m,
		snaptest.ScenarioStep{Name: "move to start", Msg: key("home")},
		snaptest.ScenarioStep{Name: "delete next word", Msg: key("ctrl+delete")},
		snaptest.ScenarioStep{Name: "yank deleted word", Msg: key("alt+y")},
	)
	snaptest.SnapScenario(t, result)
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

type scenarioModel struct {
	textarea Model
}

func (m scenarioModel) Init() tea.Cmd { return nil }

func (m scenarioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m scenarioModel) View() tea.View { return tea.NewView(m.textarea.View()) }
