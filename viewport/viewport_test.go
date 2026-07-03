package viewport

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/snaptest"
)

func newTestViewport(w, h, contentLines int) Model {
	vp := New(gotui.Dark())
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
	if vp.YOffset() != 3 {
		t.Errorf("after wheel down: yoff = %d, want 3", vp.YOffset())
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
