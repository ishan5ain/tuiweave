package overlay

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui/snaptest"
)

func base() string {
	rows := make([]string, 5)
	for i := range rows {
		rows[i] = strings.Repeat("░", 20)
	}
	return strings.Join(rows, "\n")
}

func TestPlaceGolden(t *testing.T) {
	over := "┌────┐\n│ hi │\n└────┘"
	snaptest.Snap(t, Place(base(), over, 3, 1))
}

func TestCenterGolden(t *testing.T) {
	over := "┌────┐\n│ hi │\n└────┘"
	snaptest.Snap(t, Center(base(), over))
}

func TestPlaceClipsAtEdges(t *testing.T) {
	out := Place(base(), "XXXXXX\nXXXXXX", 17, 4)
	lines := strings.Split(out, "\n")
	if len(lines) != 5 {
		t.Fatalf("composed view has %d lines, want 5 (base size)", len(lines))
	}
	snaptest.Snap(t, out)
}

func TestOverlayPreservesStyles(t *testing.T) {
	styledBase := "\x1b[31maaaaaaaaaa\x1b[0m\n\x1b[31maaaaaaaaaa\x1b[0m"
	styledOver := "\x1b[1;34mBB\x1b[0m"
	out := Place(styledBase, styledOver, 4, 0)
	snaptest.SnapCells(t, out)
}

func TestOverlayPreservesCombiningGraphemes(t *testing.T) {
	view := Place("        ", "e\u0301 ", 0, 0)
	out := ansi.Strip(view)
	if !strings.Contains(out, "e\u0301") {
		t.Fatalf("overlay lost combining grapheme: %q", out)
	}
	snaptest.Snap(t, view)
	snaptest.SnapCells(t, view)
}
