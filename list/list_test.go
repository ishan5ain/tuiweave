package list

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func newTestList(w, h, n int) Model {
	l := New(tuiweave.Dark())
	l.SetSize(w, h)
	items := make([]string, n)
	for i := range items {
		items[i] = fmt.Sprintf("item %02d", i+1)
	}
	l.SetItems(items...)
	return l
}

func keyPress(s string) tea.KeyPressMsg {
	if len(s) == 1 {
		return tea.KeyPressMsg{Code: rune(s[0]), Text: s}
	}
	panic("unsupported key in test helper: " + s)
}

func TestListGoldenFocused(t *testing.T) {
	l := newTestList(14, 4, 8)
	l.Focus()
	l.Select(1)
	snaptest.Snap(t, l.View())
	snaptest.SnapCells(t, l.View(), snaptest.WithRoles(tuiweave.Dark()))
}

func TestListGoldenBlurred(t *testing.T) {
	l := newTestList(14, 4, 8)
	l.Select(1)
	snaptest.SnapCells(t, l.View(), snaptest.WithRoles(tuiweave.Dark()))
}

func TestListNavigationAndWindowing(t *testing.T) {
	l := newTestList(14, 4, 8)
	l.Focus()

	for range 5 {
		l, _ = l.Update(keyPress("j"))
	}
	if got := l.Selected(); got != 5 {
		t.Fatalf("after 5×j: selected = %d, want 5", got)
	}
	snaptest.Snap(t, l.View()) // window must have scrolled to keep selection visible

	l, _ = l.Update(keyPress("G"))
	if got := l.Selected(); got != 7 {
		t.Fatalf("after G: selected = %d, want 7", got)
	}
	l, _ = l.Update(keyPress("g"))
	if got := l.Selected(); got != 0 {
		t.Fatalf("after g: selected = %d, want 0", got)
	}
}

func TestListEmpty(t *testing.T) {
	l := New(tuiweave.Dark())
	l.SetSize(10, 3)
	if got := l.Selected(); got != -1 {
		t.Errorf("empty Selected() = %d, want -1", got)
	}
	if got := l.SelectedItem(); got != "" {
		t.Errorf("empty SelectedItem() = %q, want empty", got)
	}
	l.Focus()
	l, _ = l.Update(keyPress("j")) // must not panic
	_ = l.View()
}

func TestListTruncatesLongItems(t *testing.T) {
	l := New(tuiweave.Dark())
	l.SetSize(10, 2)
	l.SetItems("a very long item name", "short")
	l.Focus()
	snaptest.Snap(t, l.View())
}

func TestListWidthOneGolden(t *testing.T) {
	views := make([]string, 0, 2)
	for _, focused := range []bool{false, true} {
		l := newTestList(1, 2, 2)
		if focused {
			l.Focus()
		}
		for i, row := range strings.Split(l.View(), "\n") {
			if got := lipgloss.Width(row); got != 1 {
				t.Fatalf("focused=%v row %d width=%d, want 1", focused, i, got)
			}
		}
		views = append(views, l.View())
	}
	snaptest.SnapCells(t, strings.Join(views, "\n"), snaptest.WithRoles(tuiweave.Dark()))
}

func TestItemsAreDefensiveCopies(t *testing.T) {
	items := []string{"alpha", "beta"}
	l := New(tuiweave.Dark())
	l.SetItems(items...)
	items[0] = "changed by caller"
	got := l.Items()
	got[1] = "changed through getter"
	if l.SelectedItem() != "alpha" || l.Items()[1] != "beta" {
		t.Fatalf("list items were mutated through shared storage: %q", l.Items())
	}
}

func TestListBlurredIgnoresKeys(t *testing.T) {
	l := newTestList(14, 4, 8)
	l, _ = l.Update(keyPress("j"))
	if got := l.Selected(); got != 0 {
		t.Errorf("blurred list moved selection to %d", got)
	}
}

func TestFilterNarrowsAndReportsOriginalIndex(t *testing.T) {
	l := New(tuiweave.Dark())
	l.SetSize(14, 4)
	l.SetItems("alpha", "beta", "gamma", "beacon")
	l.Focus()

	l.SetFilter("BE") // case-insensitive
	if got := l.FilteredLen(); got != 2 {
		t.Fatalf("FilteredLen = %d, want 2 (beta, beacon)", got)
	}
	if got := l.Selected(); got != 1 {
		t.Fatalf("first match Selected() = %d, want 1 (beta's original index)", got)
	}
	l, _ = l.Update(keyPress("j"))
	if got, item := l.Selected(), l.SelectedItem(); got != 3 || item != "beacon" {
		t.Fatalf("second match = (%d, %q), want (3, beacon)", got, item)
	}
	snaptest.Snap(t, l.View())
}

func TestFilterKeepsSelectionWhenStillMatching(t *testing.T) {
	l := New(tuiweave.Dark())
	l.SetSize(14, 4)
	l.SetItems("alpha", "beta", "gamma")
	l.Select(2) // gamma
	l.SetFilter("ma")
	if got := l.SelectedItem(); got != "gamma" {
		t.Errorf("selection not kept through filter: %q", got)
	}
	l.SetFilter("")
	if got := l.Selected(); got != 2 {
		t.Errorf("selection lost after clearing filter: %d", got)
	}
}

func TestFilterNoMatches(t *testing.T) {
	l := New(tuiweave.Dark())
	l.SetSize(14, 3)
	l.SetItems("alpha", "beta")
	l.SetFilter("zzz")
	if got := l.Selected(); got != -1 {
		t.Errorf("no-match Selected() = %d, want -1", got)
	}
	if got := l.SelectedItem(); got != "" {
		t.Errorf("no-match SelectedItem() = %q, want empty", got)
	}
	l.Focus()
	l, _ = l.Update(keyPress("j")) // must not panic
	_ = l.View()
}

func TestFilteredWindowingAndScrollable(t *testing.T) {
	l := New(tuiweave.Dark())
	l.SetSize(14, 2)
	l.SetItems("m1", "x", "m2", "x", "m3", "x", "m4")
	l.Focus()
	l.SetFilter("m")

	if got := l.TotalLines(); got != 4 {
		t.Fatalf("TotalLines = %d, want 4 matches", got)
	}
	l, _ = l.Update(keyPress("G"))
	if got := l.SelectedItem(); got != "m4" {
		t.Fatalf("G selected %q, want m4", got)
	}
	if got := l.YOffset(); got != 2 {
		t.Errorf("YOffset = %d, want 2 (window scrolled in filtered space)", got)
	}
}

func TestSelectNonMatchingIndexIsNoop(t *testing.T) {
	l := New(tuiweave.Dark())
	l.SetSize(14, 4)
	l.SetItems("alpha", "beta", "gamma")
	l.SetFilter("al") // only alpha matches
	l.Select(2)       // gamma doesn't match: no-op
	if got := l.SelectedItem(); got != "alpha" {
		t.Errorf("Select(non-matching) changed selection to %q", got)
	}
}
