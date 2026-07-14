package viewport

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/mouse"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func newTestViewport(w, h, contentLines int) Model {
	vp := New(tuiweave.Dark())
	vp.SetSize(w, h)
	lines := make([]string, contentLines)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %02d", i+1)
	}
	vp.SetContent(strings.Join(lines, "\n"))
	return vp
}

func keyPress(s string) tea.KeyPressMsg {
	if len(s) == 1 {
		return tea.KeyPressMsg{Code: rune(s[0]), Text: s}
	}
	switch s {
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "end":
		return tea.KeyPressMsg{Code: tea.KeyEnd}
	}
	panic("unsupported key in test helper: " + s)
}

func TestViewportGolden(t *testing.T) {
	vp := newTestViewport(12, 4, 10)
	snaptest.Snap(t, vp.View())
}

func TestViewportScrollKeys(t *testing.T) {
	vp := newTestViewport(12, 4, 10)
	vp.Focus()

	vp, _ = vp.Update(keyPress("j"))
	if vp.YOffset() != 1 {
		t.Fatalf("after j: yoff = %d, want 1", vp.YOffset())
	}
	vp, _ = vp.Update(keyPress("G"))
	if !vp.AtBottom() {
		t.Fatalf("after G: not at bottom (yoff=%d)", vp.YOffset())
	}
	snaptest.Snap(t, vp.View())
}

func TestViewportIgnoresKeysWhenBlurred(t *testing.T) {
	vp := newTestViewport(12, 4, 10)
	vp, _ = vp.Update(keyPress("j"))
	if vp.YOffset() != 0 {
		t.Errorf("blurred viewport scrolled to %d", vp.YOffset())
	}
}

func TestViewportWheelScrollsEvenBlurred(t *testing.T) {
	vp := newTestViewport(12, 4, 10)
	vp, _ = vp.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if vp.YOffset() != mouse.WheelLines {
		t.Errorf("after wheel down: yoff = %d, want %d", vp.YOffset(), mouse.WheelLines)
	}
	vp, _ = vp.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if vp.YOffset() != 0 {
		t.Errorf("after wheel up: yoff = %d, want 0", vp.YOffset())
	}
	vp, _ = vp.Update(tea.MouseWheelMsg{Button: tea.MouseWheelLeft})
	if vp.YOffset() != 0 {
		t.Errorf("after horizontal wheel: yoff = %d, want 0", vp.YOffset())
	}
}

func TestViewportClampsAndPads(t *testing.T) {
	vp := newTestViewport(10, 6, 3) // content shorter than box
	vp.ScrollBy(99)
	if vp.YOffset() != 0 {
		t.Errorf("yoff = %d, want 0 when content fits", vp.YOffset())
	}
	view := vp.View()
	if got := len(strings.Split(view, "\n")); got != 6 {
		t.Fatalf("rendered %d rows, want 6", got)
	}
	for i, row := range strings.Split(view, "\n") {
		if len(row) != 10 {
			t.Errorf("row %d width = %d, want 10 (%q)", i, len(row), row)
		}
	}
}

func TestViewportScrollPercent(t *testing.T) {
	vp := newTestViewport(12, 4, 10)
	if got := vp.ScrollPercent(); got != 0 {
		t.Errorf("initial percent = %v, want 0", got)
	}
	vp.GotoBottom()
	if got := vp.ScrollPercent(); got != 1 {
		t.Errorf("bottom percent = %v, want 1", got)
	}
}

func TestViewportUnicodeNarrowRowsStayWithinBox(t *testing.T) {
	content := "界界界\ne\u0301 combining\n👩‍💻 developer\nshort"
	for _, width := range []int{0, 1, 2, 3, 4, 8, 16} {
		for _, focused := range []bool{false, true} {
			vp := New(tuiweave.Dark())
			vp.SetSize(width, 3)
			vp.SetContent(content)
			if focused {
				vp.Focus()
			}

			view := vp.View()
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

func TestViewportUnicodeTruncationPreservesClusters(t *testing.T) {
	for _, test := range []struct {
		name  string
		width int
		want  string
	}{
		{name: "combining", width: 1, want: "e\u0301"},
		{name: "emoji too narrow", width: 1, want: " "},
		{name: "emoji", width: 2, want: "👩‍💻"},
		{name: "emoji and following rune", width: 3, want: "👩‍💻x"},
	} {
		t.Run(test.name, func(t *testing.T) {
			vp := New(tuiweave.Dark())
			vp.SetSize(test.width, 1)
			content := test.want
			if test.name == "emoji too narrow" || test.name == "emoji" || test.name == "emoji and following rune" {
				content = "👩‍💻x"
			}
			if test.name == "combining" {
				content = "e\u0301x"
			}
			vp.SetContent(content)
			if got := ansi.Strip(vp.View()); got != test.want {
				t.Fatalf("rendered %q, want %q", got, test.want)
			}
		})
	}
}

func TestViewportUnicodeScrollKeepsRowsBounded(t *testing.T) {
	vp := New(tuiweave.Dark())
	vp.SetSize(12, 2)
	vp.SetContent("界 one\ne\u0301 two\n👩‍💻 three\n終 four")
	vp.Focus()
	vp, _ = vp.Update(keyPress("end"))
	if !vp.AtBottom() {
		t.Fatalf("viewport is not at bottom: offset=%d", vp.YOffset())
	}
	for row, line := range strings.Split(ansi.Strip(vp.View()), "\n") {
		if got := ansi.StringWidth(line); got != 12 {
			t.Fatalf("row %d width = %d, want 12", row, got)
		}
	}
}
