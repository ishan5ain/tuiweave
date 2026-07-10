package stack

import (
	"strings"
	"testing"

	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/frame"
	"github.com/ishansain/gotui/snaptest"
)

func TestVertical(t *testing.T) {
	view := Vertical(gotui.Dark(), 28, Options{Gap: 1, Divider: true},
		func(width int) string { return "header" },
		func(width int) string { return "body\nsecond line" },
		func(width int) string { return "footer" },
	)
	assertWidth(t, view, 28)
	snaptest.Snap(t, view)
	snaptest.SnapCells(t, view, snaptest.WithRoles(gotui.Dark()))
}

func TestVerticalSkipsEmptySections(t *testing.T) {
	view := Vertical(gotui.Light(), 20, Options{},
		nil,
		func(width int) string { return "header" },
		func(width int) string { return "" },
		func(width int) string { return frame.Divider(gotui.Light(), width) },
		func(width int) string { return "footer" },
	)
	assertWidth(t, view, 20)
	snaptest.Snap(t, view)
}

func TestVerticalWidthAndEmpty(t *testing.T) {
	if got := Vertical(gotui.Dark(), 0, Options{}, func(int) string { return "content" }); got != "" {
		t.Fatalf("zero-width stack = %q, want empty", got)
	}
	if got := Vertical(gotui.Dark(), 20, Options{}, nil, func(int) string { return "" }); got != "" {
		t.Fatalf("empty stack = %q, want empty", got)
	}
	for width := 1; width <= 24; width++ {
		view := Vertical(gotui.Dark(), width, Options{},
			func(int) string { return "a" },
			func(int) string { return "b\nc" },
		)
		assertWidth(t, view, width)
	}
}

func assertWidth(t *testing.T, view string, want int) {
	t.Helper()
	for i, line := range strings.Split(view, "\n") {
		if got := lipgloss.Width(line); got != want {
			t.Fatalf("line %d width = %d, want %d: %q", i+1, got, want, line)
		}
	}
}
