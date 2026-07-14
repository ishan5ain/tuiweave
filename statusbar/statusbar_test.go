package statusbar

import (
	"strings"
	"testing"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func newTestBar(width int) Model {
	sb := New(tuiweave.Dark())
	sb.SetSize(width, 1)
	sb.SetLeft(
		Segment{Text: "tuiweave", Kind: KindAccent},
		Segment{Text: "main.go", Kind: KindNormal},
	)
	sb.SetRight(
		Segment{Text: "ok", Kind: KindSuccess},
		Segment{Text: "12:4", Kind: KindMuted},
	)
	return sb
}

func TestStatusbarSegments(t *testing.T) {
	sb := newTestBar(60)
	view := sb.View()

	if got := lipgloss.Width(view); got != 60 {
		t.Fatalf("rendered width = %d, want 60", got)
	}
	snaptest.Snap(t, view)
	snaptest.SnapCells(t, view, snaptest.WithRoles(tuiweave.Dark()))
}

func TestStatusbarDropsRightWhenTight(t *testing.T) {
	sb := newTestBar(18)
	view := sb.View()

	if got := lipgloss.Width(view); got != 18 {
		t.Fatalf("rendered width = %d, want 18", got)
	}
	snaptest.Snap(t, view)
}

func TestStatusbarTruncatesLeft(t *testing.T) {
	sb := New(tuiweave.Dark())
	sb.SetSize(12, 1)
	sb.SetLeft(Segment{Text: "a very long segment", Kind: KindNormal})
	view := sb.View()

	if got := lipgloss.Width(view); got != 12 {
		t.Fatalf("rendered width = %d, want 12", got)
	}
	snaptest.Snap(t, view)
}

func TestStatusbarZeroSizeRendersNothing(t *testing.T) {
	sb := New(tuiweave.Dark())
	if got := sb.View(); got != "" {
		t.Errorf("zero-size View() = %q, want empty", got)
	}
	sb.SetSize(40, 0)
	if got := sb.View(); got != "" {
		t.Errorf("zero-height View() = %q, want empty", got)
	}
}

func TestStatusbarLightTheme(t *testing.T) {
	sb := New(tuiweave.Light())
	sb.SetSize(40, 1)
	sb.SetLeft(Segment{Text: "light", Kind: KindAccent})
	sb.SetRight(Segment{Text: "warn", Kind: KindWarning})
	snaptest.SnapCells(t, sb.View(), snaptest.WithRoles(tuiweave.Light()))
}

func TestStatusbarUnicodeNarrowRowsStayWithinBox(t *testing.T) {
	for _, width := range []int{0, 1, 2, 3, 4, 8, 16} {
		sb := New(tuiweave.Dark())
		sb.SetSize(width, 1)
		sb.SetLeft(
			Segment{Text: "界e\u0301", Kind: KindNormal},
			Segment{Text: "👩‍💻 developer", Kind: KindAccent},
		)
		sb.SetRight(Segment{Text: "終", Kind: KindMuted})
		view := sb.View()
		if width == 0 {
			if view != "" {
				t.Fatalf("width=0 rendered %q", view)
			}
			continue
		}
		if got := ansi.StringWidth(ansi.Strip(view)); got != width {
			t.Fatalf("width=%d rendered width=%d: %q", width, got, view)
		}
	}
}

func TestStatusbarUnicodeTruncationGolden(t *testing.T) {
	sb := New(tuiweave.Dark())
	sb.SetSize(12, 1)
	sb.SetLeft(Segment{Text: "👩‍💻 developer", Kind: KindNormal})
	snap := sb.View()
	snaptest.Snap(t, snap)
	snaptest.SnapCells(t, snap, snaptest.WithRoles(tuiweave.Dark()))
}

func TestStatusbarNarrowEmojiDoesNotSplit(t *testing.T) {
	for _, width := range []int{1, 2, 3} {
		sb := New(tuiweave.Dark())
		sb.SetSize(width, 1)
		sb.SetLeft(Segment{Text: "👩‍💻", Kind: KindNormal})
		view := ansi.Strip(sb.View())
		if strings.Contains(view, "👩") || strings.Contains(view, "💻") {
			t.Fatalf("width=%d split emoji segment: %q", width, view)
		}
		if got := ansi.StringWidth(view); got != width {
			t.Fatalf("width=%d rendered width=%d: %q", width, got, view)
		}
	}
}
