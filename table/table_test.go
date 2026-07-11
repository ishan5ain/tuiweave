package table

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func newTestTable(w, h, rows int) Model {
	tb := New(tuiweave.Dark())
	tb.SetSize(w, h)
	tb.SetColumns(
		Column{Title: "ID", Width: 4},
		Column{Title: "Name"}, // flex
		Column{Title: "State", Width: 7},
	)
	rr := make([][]string, rows)
	for i := range rr {
		rr[i] = []string{fmt.Sprintf("%d", i+1), fmt.Sprintf("service-%02d", i+1), "ok"}
	}
	tb.SetRows(rr...)
	return tb
}

func keyPress(s string) tea.KeyPressMsg {
	if len(s) == 1 {
		return tea.KeyPressMsg{Code: rune(s[0]), Text: s}
	}
	panic("unsupported key in test helper: " + s)
}

func TestTableGolden(t *testing.T) {
	tb := newTestTable(30, 6, 8)
	tb.Focus()
	tb.Select(1)
	view := tb.View()

	if got := len(strings.Split(view, "\n")); got != 6 {
		t.Fatalf("rendered %d lines, want 6", got)
	}
	snaptest.Snap(t, view)
	snaptest.SnapCells(t, view, snaptest.WithRoles(tuiweave.Dark()))
}

func TestTableWindowScrollsToSelection(t *testing.T) {
	tb := newTestTable(30, 5, 10) // 3 visible rows
	tb.Focus()
	tb, _ = tb.Update(keyPress("G"))
	if got := tb.Selected(); got != 9 {
		t.Fatalf("after G: selected = %d, want 9", got)
	}
	if !strings.Contains(tb.View(), "service-10") {
		t.Error("selected row not visible after G")
	}
	snaptest.Snap(t, tb.View())
}

func TestTableFlexColumnFillsWidth(t *testing.T) {
	tb := newTestTable(40, 4, 2)
	for i, line := range strings.Split(tb.View(), "\n") {
		if got := lipgloss.Width(line); got != 40 {
			t.Errorf("line %d width = %d, want 40", i, got)
		}
	}
}

func TestTableEmpty(t *testing.T) {
	tb := New(tuiweave.Dark())
	tb.SetSize(20, 4)
	if got := tb.View(); got != "" {
		t.Errorf("no-columns View() = %q, want empty", got)
	}
	tb.SetColumns(Column{Title: "A"})
	if got := tb.Selected(); got != -1 {
		t.Errorf("empty Selected() = %d, want -1", got)
	}
	tb.Focus()
	tb, _ = tb.Update(keyPress("j")) // must not panic
	_ = tb.View()
}

func TestTableNarrowWideContentStaysWithinBox(t *testing.T) {
	for width := 1; width <= 32; width++ {
		tb := New(tuiweave.Dark())
		tb.SetSize(width, 5)
		tb.SetColumns(
			Column{Title: "識別子", Width: 8},
			Column{Title: "説明"},
			Column{Title: "状態", Width: 8},
			Column{Title: "メモ"},
		)
		tb.SetRows(
			[]string{"界界界", "サービスの説明", "运行中", "👍👍"},
			[]string{"é", "combining text", "ready", "ok"},
		)

		for i, line := range strings.Split(tb.View(), "\n") {
			if got := lipgloss.Width(line); got != width {
				t.Fatalf("width %d line %d rendered as %d: %q", width, i+1, got, line)
			}
		}
	}

	tb := New(tuiweave.Dark())
	tb.SetSize(8, 5)
	tb.SetColumns(
		Column{Title: "識別子", Width: 8},
		Column{Title: "説明"},
		Column{Title: "状態", Width: 8},
		Column{Title: "メモ"},
	)
	tb.SetRows(
		[]string{"界界界", "サービスの説明", "运行中", "👍👍"},
		[]string{"é", "combining text", "ready", "ok"},
	)
	snaptest.Snap(t, tb.View())
	snaptest.SnapCells(t, tb.View(), snaptest.WithRoles(tuiweave.Dark()))
}
