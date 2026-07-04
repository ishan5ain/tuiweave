package toolcall

import (
	"strings"
	"testing"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/snaptest"
)

func newTestBlock() *Block {
	b := New(gotui.Dark(), "Bash", "go test ./...")
	b.SetStatus(StatusSuccess)
	b.AppendOutput("ok  \tgotui/layout\t0.3s\nok  \tgotui/snaptest\t0.2s")
	return b
}

func TestCollapsedGolden(t *testing.T) {
	b := newTestBlock()
	view := b.Render(50)
	if !strings.Contains(view, "(2 output lines)") {
		t.Errorf("collapsed block missing line-count hint: %q", view)
	}
	snaptest.SnapCells(t, view, snaptest.WithRoles(gotui.Dark()))
}

func TestExpandedGolden(t *testing.T) {
	b := newTestBlock()
	b.Expanded = true
	snaptest.Snap(t, b.Render(50))
	snaptest.SnapCells(t, b.Render(50), snaptest.WithRoles(gotui.Dark()))
}

func TestOutputCapAndHiddenCount(t *testing.T) {
	b := New(gotui.Dark(), "Read", "main.go")
	b.MaxOutputLines = 3
	b.Expanded = true
	b.AppendOutput("l1\nl2\nl3\nl4\nl5")
	view := b.Render(40)
	if !strings.Contains(view, "+2 more lines") {
		t.Errorf("capped output missing hidden count: %q", view)
	}
	if strings.Contains(view, "l4") {
		t.Errorf("capped output shows hidden line: %q", view)
	}
}

func TestStatusIcons(t *testing.T) {
	b := New(gotui.Dark(), "Bash", "")
	for status, icon := range icons {
		b.SetStatus(status)
		if !strings.Contains(b.Render(30), icon) {
			t.Errorf("status %d: icon %q missing", status, icon)
		}
	}
}

func TestNoOutputNoHint(t *testing.T) {
	b := New(gotui.Dark(), "Bash", "ls")
	if view := b.Render(30); strings.Contains(view, "output lines") {
		t.Errorf("block without output shows hint: %q", view)
	}
}
