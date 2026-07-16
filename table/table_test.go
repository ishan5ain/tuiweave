package table

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/inspect"
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

func TestTableUnicodeNarrowRowsStayWithinBox(t *testing.T) {
	for _, width := range []int{0, 1, 2, 3, 4, 8, 16} {
		for _, focused := range []bool{false, true} {
			tb := New(tuiweave.Dark())
			tb.SetSize(width, 5)
			tb.SetColumns(
				Column{Title: "識別子", Width: 4},
				Column{Title: "説明"},
				Column{Title: "状態", Width: 4},
			)
			tb.SetRows(
				[]string{"界界", "サービスの説明", "运行中"},
				[]string{"e\u0301", "combining", "👩‍💻"},
			)
			if focused {
				tb.Focus()
				tb.Select(1)
			}

			view := tb.View()
			if width == 0 {
				if view != "" {
					t.Fatalf("width=0 focused=%v rendered %q", focused, view)
				}
				continue
			}
			for row, line := range strings.Split(view, "\n") {
				if got := ansi.StringWidth(ansi.Strip(line)); got != width {
					t.Fatalf("width=%d focused=%v row=%d rendered width=%d: %q", width, focused, row, got, line)
				}
			}
		}
	}
}

func TestTableUnicodeTruncationPreservesClusters(t *testing.T) {
	tb := New(tuiweave.Dark())
	tb.SetSize(8, 4)
	tb.SetColumns(Column{Title: "Name"}, Column{Title: "State", Width: 3})
	tb.SetRows([]string{"👩‍💻", "e\u0301"})
	tb.Focus()
	tb.Select(0)
	view := ansi.Strip(tb.View())
	if !strings.Contains(view, "👩‍💻") {
		t.Fatalf("selected emoji was truncated or split: %q", view)
	}
	if strings.Contains(view, "👩") != strings.Contains(view, "💻") {
		t.Fatalf("emoji grapheme was split: %q", view)
	}
}

func TestTableNarrowEmojiDoesNotSplit(t *testing.T) {
	tb := New(tuiweave.Dark())
	tb.SetSize(3, 3) // one cell for each column, one gap
	tb.SetColumns(Column{Title: "A", Width: 1}, Column{Title: "B"})
	tb.SetRows([]string{"👩‍💻", "ok"})
	view := ansi.Strip(tb.View())
	if strings.Contains(view, "👩") || strings.Contains(view, "💻") {
		t.Fatalf("narrow emoji cell was split instead of truncated: %q", view)
	}
	for row, line := range strings.Split(view, "\n") {
		if got := ansi.StringWidth(line); got != 3 {
			t.Fatalf("row %d width = %d, want 3: %q", row, got, line)
		}
	}
}

func TestTableUnicodeSelectionScroll(t *testing.T) {
	tb := New(tuiweave.Dark())
	tb.SetSize(14, 4) // two visible data rows
	tb.SetColumns(Column{Title: "Name"}, Column{Title: "State", Width: 4})
	tb.SetRows(
		[]string{"界 one", "ok"},
		[]string{"e\u0301 two", "ok"},
		[]string{"👩‍💻 three", "ok"},
		[]string{"終 four", "ok"},
	)
	tb.Focus()
	tb, _ = tb.Update(keyPress("G"))
	if got := tb.Selected(); got != 3 {
		t.Fatalf("selected = %d, want 3", got)
	}
	if got := tb.YOffset(); got != 2 {
		t.Fatalf("YOffset = %d, want 2", got)
	}
	for row, line := range strings.Split(ansi.Strip(tb.View()), "\n") {
		if got := ansi.StringWidth(line); got != 14 {
			t.Fatalf("row %d width = %d, want 14", row, got)
		}
	}
}

func TestTableScenarioGolden(t *testing.T) {
	tb := newTestTable(24, 5, 8)
	result := snaptest.RunScenario(tableScenarioModel{table: tb},
		snaptest.ScenarioStep{Name: "blurred navigation is ignored", Msg: keyPress("G")},
		snaptest.ScenarioStep{Name: "focus table", Msg: inspect.Invoke(ActionFocus)},
		snaptest.ScenarioStep{Name: "select last row", Msg: keyPress("G")},
		snaptest.ScenarioStep{Name: "resize narrow", Msg: tea.WindowSizeMsg{Width: 12, Height: 4}},
	)
	snaptest.SnapScenario(t, result)

	final := result.Model.(tableScenarioModel)
	if got := final.table.Selected(); got != 7 {
		t.Fatalf("final selected row = %d, want 7", got)
	}
	if got := final.table.YOffset(); got != 6 {
		t.Fatalf("final offset = %d, want 6", got)
	}
}

type tableScenarioModel struct {
	table Model
}

func (m tableScenarioModel) Init() tea.Cmd { return nil }

func (m tableScenarioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.table.SetSize(size.Width, size.Height)
		return m, nil
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m tableScenarioModel) View() tea.View { return tea.NewView(m.table.View()) }
