package splitpane

import (
	"strings"
	"testing"

	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/snaptest"
)

func TestHorizontal(t *testing.T) {
	view := Horizontal(gotui.Dark(), 36, Options{},
		func(width int) string { return "left pane\nsecond line" },
		func(width int) string { return "right pane" },
	)
	assertWidth(t, view, 36)
	snaptest.Snap(t, view)
	snaptest.SnapCells(t, view, snaptest.WithRoles(gotui.Dark()))
}

func TestHorizontalRatioAndGap(t *testing.T) {
	var leftWidth, rightWidth int
	view := Horizontal(gotui.Light(), 40, Options{Ratio: 30, Gap: 3},
		func(width int) string {
			leftWidth = width
			return "left"
		},
		func(width int) string {
			rightWidth = width
			return "right\nsecond"
		},
	)
	assertWidth(t, view, 40)
	if leftWidth >= rightWidth {
		t.Fatalf("ratio widths = %d/%d, want left narrower", leftWidth, rightWidth)
	}
	snaptest.Snap(t, view)
}

func TestHorizontalNarrowAndNil(t *testing.T) {
	if got := Horizontal(gotui.Dark(), 2, Options{}, func(int) string { return "left" }, func(int) string { return "right" }); got != "" {
		t.Fatalf("too-narrow split = %q, want empty", got)
	}
	if got := Horizontal(gotui.Dark(), 20, Options{}, nil, func(int) string { return "right" }); got != "" {
		t.Fatalf("nil left split = %q, want empty", got)
	}
	view := Horizontal(gotui.Dark(), 20, Options{Gap: -1},
		func(int) string { return "left" },
		func(int) string { return "right" },
	)
	assertWidth(t, view, 20)
}

func assertWidth(t *testing.T, view string, want int) {
	t.Helper()
	for i, line := range strings.Split(view, "\n") {
		if got := lipgloss.Width(line); got != want {
			t.Fatalf("line %d width = %d, want %d: %q", i+1, got, want, line)
		}
	}
}
