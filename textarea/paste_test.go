package textarea

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func TestPaste(t *testing.T) {
	for _, tc := range []struct{ name, content, want string }{
		{"plain", "hello", "hello"},
		{"unicode", "界e\u0301👩‍💻", "界e\u0301👩‍💻"},
		{"multiline", "a\r\nb\tc", "a\nb\tc"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newFocused(16, 4)
			m.SetValue("old")
			m.SelectAll()
			var cmd tea.Cmd
			m, cmd = m.Update(tea.PasteMsg{Content: tc.content})
			if cmd != nil || m.Value() != tc.want || m.HasSelection() {
				t.Fatalf("paste = %q, selection=%v, cmd=%v", m.Value(), m.HasSelection(), cmd != nil)
			}
			lines := strings.Split(tc.want, "\n")
			if got := m.CursorPosition(); got != (Position{Row: len(lines) - 1, Column: len([]rune(lines[len(lines)-1]))}) {
				t.Fatalf("cursor=%#v", got)
			}
			m.Undo()
			if m.Value() != "old" || m.CanUndo() {
				t.Fatal("paste was not one undoable edit")
			}
			m.Redo()
			if m.Value() != tc.want {
				t.Fatal("redo did not restore paste")
			}
			m.Blur()
			m, cmd = m.Update(tea.PasteMsg{Content: "ignored"})
			if m.Value() != tc.want || cmd != nil {
				t.Fatal("blurred paste changed input")
			}
		})
	}
}

func TestPasteNoopPreservesSelectionAndHistory(t *testing.T) {
	for _, msg := range []tea.Msg{tea.PasteMsg{}, tea.PasteMsg{Content: "\x1b[31m\x00\x7f"}, tea.PasteStartMsg{}, tea.PasteEndMsg{}} {
		m := newFocused(20, 3)
		m.SetValue("old")
		m.InsertString("!")
		m.SelectAll()
		before := m.CursorPosition()
		next, cmd := m.Update(msg)
		if cmd != nil || next.Value() != m.Value() || next.SelectedText() != m.SelectedText() || next.CursorPosition() != before || !next.CanUndo() || next.CanRedo() {
			t.Fatal("no-op changed editing state")
		}
		next.Undo()
		if next.Value() != "old" || next.CanUndo() {
			t.Fatal("no-op changed history")
		}
	}
}

func TestPasteInMiddle(t *testing.T) {
	m := newFocused(20, 4)
	m.SetValue("ab")
	m, _ = m.Update(key("left"))
	m, cmd := m.Update(tea.PasteMsg{Content: "界\r\ne\u0301"})
	if cmd != nil || m.Value() != "a界\ne\u0301b" || m.CursorPosition() != (Position{Row: 1, Column: 2}) {
		t.Fatalf("middle paste = %q at %#v", m.Value(), m.CursorPosition())
	}
}

func TestPasteGolden(t *testing.T) {
	for _, width := range []int{18, 5, 3} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			m := newFocused(width, 4)
			m, _ = m.Update(tea.PasteMsg{Content: "界e\u0301\r\n👩‍💻 end"})
			view := m.View()
			if strings.Count(view, "\n") != 3 {
				t.Fatalf("height changed: %q", view)
			}
			for _, line := range strings.Split(view, "\n") {
				if ansi.StringWidth(line) > width {
					t.Fatalf("width exceeded: %q", line)
				}
			}
			snaptest.Snap(t, view)
			snaptest.SnapCells(t, view, snaptest.WithRoles(tuiweave.Dark()))
		})
	}
}

func TestPasteScenario(t *testing.T) {
	m := scenarioModel{textarea: newFocused(12, 4)}
	m.textarea.SetValue("old")
	result := snaptest.RunScenario(m,
		snaptest.ScenarioStep{Name: "select all", Msg: key("ctrl+a")},
		snaptest.ScenarioStep{Name: "paste starts", Msg: tea.PasteStartMsg{}},
		snaptest.ScenarioStep{Name: "paste replaces selection", Msg: tea.PasteMsg{Content: "界e\u0301\r\nx"}},
		snaptest.ScenarioStep{Name: "paste ends", Msg: tea.PasteEndMsg{}},
		snaptest.ScenarioStep{Name: "undo paste", Msg: key("ctrl+z")},
		snaptest.ScenarioStep{Name: "redo paste", Msg: key("ctrl+y")},
	)
	snaptest.SnapScenario(t, result)
}

func TestPasteResetsTransientEditingAndRedo(t *testing.T) {
	m := newFocused(20, 4)
	m.SetValue("one two")
	m, _ = m.Update(key("ctrl+w"))
	if !m.canCoalesceKill || !m.CanYank() {
		t.Fatal("kill did not establish transient state")
	}
	m, _ = m.Update(tea.PasteMsg{Content: "X"})
	if m.canCoalesceKill || m.yankActive || !m.CanYank() {
		t.Fatal("paste did not break kill chain while preserving ring")
	}
	m.Yank()
	if !m.yankActive {
		t.Fatal("yank was not active")
	}
	next, cmd := m.Update(tea.PasteMsg{Content: "!"})
	if cmd != nil || next.yankActive {
		t.Fatal("paste did not end yank rotation")
	}
	next.Undo()
	if !next.CanRedo() {
		t.Fatal("undo did not establish redo")
	}
	next, cmd = next.Update(tea.PasteMsg{Content: "?"})
	if cmd != nil || next.CanRedo() {
		t.Fatal("paste did not invalidate redo branch")
	}
}

func TestTabPasteGeometry(t *testing.T) {
	m := newFocused(12, 4)
	m, cmd := m.Update(tea.PasteMsg{Content: "\tabcdefghij"})
	if cmd != nil || m.ContentHeight() != 2 {
		t.Fatalf("height=%d", m.ContentHeight())
	}
	plain := ansi.Strip(m.View())
	if !strings.Contains(plain, "ghij") {
		t.Fatalf("tab clipped suffix: %q", plain)
	}
	if got := cellWidth([]rune("\ta"), 10); got != 5 {
		t.Fatalf("tab width=%d, want 5", got)
	}
}

func TestTabPasteGolden(t *testing.T) {
	for _, width := range []int{12, 6, 5, 3} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			m := newFocused(width, 16)
			m, cmd := m.Update(tea.PasteMsg{Content: "\tabcdefghij"})
			if cmd != nil || m.Value() != "\tabcdefghij" {
				t.Fatal("tab paste altered logical value")
			}
			plain := ansi.Strip(m.View())
			for _, char := range "abcdefghij" {
				if !strings.ContainsRune(plain, char) {
					t.Fatalf("lost %c in %q", char, plain)
				}
			}
			for _, row := range strings.Split(m.View(), "\n") {
				if ansi.StringWidth(row) != width {
					t.Fatalf("row is not exact width: %q", row)
				}
			}
			snaptest.Snap(t, m.View())
			snaptest.SnapCells(t, m.View(), snaptest.WithRoles(tuiweave.Dark()))
		})
	}
}

func TestTabCursorAndSelectionGolden(t *testing.T) {
	m := newFocused(12, 4)
	m.SetValue("\tabcdefghij")
	m.SetCursor(0, 1)
	rows := m.visualRows()
	row, col := m.cursorVisual(rows)
	if row != 0 || col != 4 {
		t.Fatalf("cursor after tab=(%d,%d), want (0,4)", row, col)
	}
	m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if cmd != nil || m.CursorPosition() != (Position{Row: 0, Column: 11}) {
		t.Fatalf("down from tab=%#v", m.CursorPosition())
	}
	m.SetCursor(0, 0)
	m, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModShift})
	if cmd != nil || m.SelectedText() != "\t" {
		t.Fatalf("selected tab=%q", m.SelectedText())
	}
	snaptest.Snap(t, m.View())
	snaptest.SnapCells(t, m.View(), snaptest.WithRoles(tuiweave.Dark()))
	m, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if cmd != nil || m.CursorPosition() != (Position{Row: 0, Column: 1}) {
		t.Fatalf("right across tab=%#v", m.CursorPosition())
	}
	m, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	if cmd != nil || m.Value() != "abcdefghij" {
		t.Fatal("backspace failed to remove one logical tab")
	}
}

func TestTabGeometryAtNarrowWidth(t *testing.T) {
	for _, width := range []int{3, 5, 6, 12} {
		m := newFocused(width, 16)
		m.SetValue("\t\tx")
		rows := m.visualRows()
		m.SetCursor(0, 1)
		row, col := m.cursorVisual(rows)
		wantRow, wantCol := 0, tabWidth
		if m.wrapWidth() <= tabWidth {
			wantRow, wantCol = 1, 0
		}
		if row != wantRow || col != wantCol {
			t.Fatalf("width %d cursor after full-row tab=(%d,%d)", width, row, col)
		}
		if got := runeOffsetAtCell(rows[0].text, min(tabWidth, m.wrapWidth()), m.wrapWidth()); got != 1 {
			t.Fatalf("width %d offset=%d", width, got)
		}
	}
}
