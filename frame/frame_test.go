package frame

import (
	"strings"
	"testing"

	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/layout"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func TestPanel(t *testing.T) {
	view := Panel(tuiweave.Dark(), "name: tuiweave\nstatus: ready", 28, PanelOptions{
		Title:   "Session",
		Padding: 1,
	})
	assertWidth(t, view, 28)
	snaptest.Snap(t, view)
	snaptest.SnapCells(t, view, snaptest.WithRoles(tuiweave.Dark()))
}

func TestPanelFocusedNarrow(t *testing.T) {
	view := Panel(tuiweave.Light(), "running", 14, PanelOptions{
		Title:   "Long title",
		Focused: true,
		Padding: 1,
	})
	assertWidth(t, view, 14)
	snaptest.Snap(t, view)
	snaptest.SnapCells(t, view, snaptest.WithRoles(tuiweave.Light()))
}

func TestDecorations(t *testing.T) {
	view := strings.Join([]string{
		Badge(tuiweave.Dark(), "edit", BadgeAccent),
		Badge(tuiweave.Dark(), "ready", BadgeSuccess),
		Badge(tuiweave.Dark(), "3 files", BadgeMuted),
		Divider(tuiweave.Dark(), 24),
	}, "\n")
	snaptest.Snap(t, view)
	snaptest.SnapCells(t, view, snaptest.WithRoles(tuiweave.Dark()))
}

func TestPanelNarrowAndEmpty(t *testing.T) {
	if got := Panel(tuiweave.Dark(), "ignored", 1, PanelOptions{}); got != "" {
		t.Fatalf("Panel width 1 = %q, want empty", got)
	}
	if got := Divider(tuiweave.Dark(), 0); got != "" {
		t.Fatalf("Divider width 0 = %q, want empty", got)
	}
	if got := lipgloss.Width(Badge(tuiweave.Dark(), "ok", BadgeInfo)); got != 4 {
		t.Fatalf("Badge width = %d, want 4", got)
	}
	for _, content := range []string{"", "a long line that must wrap"} {
		for width := 2; width <= 24; width++ {
			view := Panel(tuiweave.Dark(), content, width, PanelOptions{
				Title:   "title",
				Padding: 2,
			})
			assertWidth(t, view, width)
		}
	}
}

func TestPanelContentRect(t *testing.T) {
	area := layout.NewRect(10, 20, 40, 12)
	got := PanelContentRect(area, PanelOptions{Padding: 1})
	want := layout.NewRect(12, 22, 36, 8)
	if got != want {
		t.Fatalf("content rect = %#v, want %#v", got, want)
	}

	narrow := PanelContentRect(layout.NewRect(0, 0, 3, 3), PanelOptions{Padding: 4})
	if narrow.Dx() != 1 || narrow.Dy() != 1 || narrow.Min.X != 1 || narrow.Min.Y != 1 {
		t.Fatalf("narrow content rect = %#v, want one-cell inset", narrow)
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
