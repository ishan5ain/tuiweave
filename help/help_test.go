package help

import (
	"strings"
	"testing"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func newTestHelp(width int) Model {
	h := New(tuiweave.Dark())
	h.SetSize(width, 1)
	h.SetBindings(
		Binding{Key: "tab", Desc: "focus"},
		Binding{Key: "enter", Desc: "send"},
		Binding{Key: "q", Desc: "quit"},
	)
	return h
}

func TestHelpGolden(t *testing.T) {
	h := newTestHelp(60)
	snaptest.Snap(t, h.View())
	snaptest.SnapCells(t, h.View(), snaptest.WithRoles(tuiweave.Dark()))
}

func TestHelpDropsWholeHintsWhenNarrow(t *testing.T) {
	h := newTestHelp(16) // fits "tab focus" but not "• enter send"
	view := h.View()
	if strings.Contains(view, "enter") {
		t.Errorf("narrow help bar shows a hint that does not fit: %q", view)
	}
	if got := lipgloss.Width(view); got > 16 {
		t.Errorf("rendered width = %d, want <= 16", got)
	}
	snaptest.Snap(t, view)
}

func TestHelpEmpty(t *testing.T) {
	h := New(tuiweave.Dark())
	h.SetSize(40, 1)
	if got := h.View(); got != "" {
		t.Errorf("empty help View() = %q, want empty", got)
	}
}

func TestHelpUnicodeNarrowRowsStayWithinBox(t *testing.T) {
	for _, width := range []int{0, 1, 2, 3, 4, 8, 16} {
		h := New(tuiweave.Dark())
		h.SetSize(width, 1)
		h.SetBindings(
			Binding{Key: "界", Desc: "e\u0301"},
			Binding{Key: "👩‍💻", Desc: "developer"},
			Binding{Key: "q", Desc: "quit"},
		)
		view := h.View()
		if width == 0 {
			if view != "" {
				t.Fatalf("width=0 rendered %q", view)
			}
			continue
		}
		if got := ansi.StringWidth(ansi.Strip(view)); got > width {
			t.Fatalf("width=%d rendered width=%d: %q", width, got, view)
		}
	}
}

func TestHelpUnicodeGolden(t *testing.T) {
	h := New(tuiweave.Dark())
	h.SetSize(30, 1)
	h.SetBindings(
		Binding{Key: "界", Desc: "e\u0301"},
		Binding{Key: "👩‍💻", Desc: "developer"},
		Binding{Key: "q", Desc: "quit"},
	)
	snaptest.Snap(t, h.View())
	snaptest.SnapCells(t, h.View(), snaptest.WithRoles(tuiweave.Dark()))
}
