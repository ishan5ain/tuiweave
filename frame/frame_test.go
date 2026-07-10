package frame

import (
	"strings"
	"testing"

	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/snaptest"
)

func TestPanel(t *testing.T) {
	view := Panel(gotui.Dark(), "name: gotui\nstatus: ready", 28, PanelOptions{
		Title:   "Session",
		Padding: 1,
	})
	assertWidth(t, view, 28)
	snaptest.Snap(t, view)
	snaptest.SnapCells(t, view, snaptest.WithRoles(gotui.Dark()))
}

func TestPanelFocusedNarrow(t *testing.T) {
	view := Panel(gotui.Light(), "running", 14, PanelOptions{
		Title:   "Long title",
		Focused: true,
		Padding: 1,
	})
	assertWidth(t, view, 14)
	snaptest.Snap(t, view)
	snaptest.SnapCells(t, view, snaptest.WithRoles(gotui.Light()))
}

func TestDecorations(t *testing.T) {
	view := strings.Join([]string{
		Badge(gotui.Dark(), "edit", BadgeAccent),
		Badge(gotui.Dark(), "ready", BadgeSuccess),
		Badge(gotui.Dark(), "3 files", BadgeMuted),
		Divider(gotui.Dark(), 24),
	}, "\n")
	snaptest.Snap(t, view)
	snaptest.SnapCells(t, view, snaptest.WithRoles(gotui.Dark()))
}

func TestPanelNarrowAndEmpty(t *testing.T) {
	if got := Panel(gotui.Dark(), "ignored", 1, PanelOptions{}); got != "" {
		t.Fatalf("Panel width 1 = %q, want empty", got)
	}
	if got := Divider(gotui.Dark(), 0); got != "" {
		t.Fatalf("Divider width 0 = %q, want empty", got)
	}
	if got := lipgloss.Width(Badge(gotui.Dark(), "ok", BadgeInfo)); got != 4 {
		t.Fatalf("Badge width = %d, want 4", got)
	}
	for _, content := range []string{"", "a long line that must wrap"} {
		for width := 2; width <= 24; width++ {
			view := Panel(gotui.Dark(), content, width, PanelOptions{
				Title:   "title",
				Padding: 2,
			})
			assertWidth(t, view, width)
		}
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
