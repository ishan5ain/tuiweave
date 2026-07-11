package snaptest

import (
	"fmt"
	"image/color"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

type fakeTheme struct {
	Accent  color.Color
	Text    color.Color
	private color.Color //nolint:unused // exercises the exported-fields-only rule
}

func TestWithRolesMapsColors(t *testing.T) {
	accent := color.RGBA{R: 0x7a, G: 0xa2, B: 0xf7, A: 0xff}
	cfg := config{roles: map[rgba]string{}}
	WithRoles(fakeTheme{Accent: accent, Text: color.RGBA{A: 0xff}})(&cfg)

	if got := colorName(accent, cfg); got != "Accent" {
		t.Errorf("colorName(accent) = %q, want Accent", got)
	}
	if got := colorName(color.RGBA{R: 0x01, A: 0xff}, cfg); got != "#010000" {
		t.Errorf("unmapped color = %q, want #010000", got)
	}
}

func TestDescribeStyleAttrs(t *testing.T) {
	s := uv.Style{Attrs: uv.AttrBold | uv.AttrItalic}
	got := describeStyle(s, config{roles: map[rgba]string{}})
	if got != "bold italic" {
		t.Errorf("describeStyle = %q, want %q", got, "bold italic")
	}
}

// TestSnapCellsGolden exercises the full pipeline: hand-written ANSI in,
// role-labeled run listing out. Golden generated with -update and checked in.
func TestSnapCellsGolden(t *testing.T) {
	accent := color.RGBA{R: 0x7a, G: 0xa2, B: 0xf7, A: 0xff}
	// " main " bold with accent fg, then plain " file.go", second line plain.
	view := "\x1b[1;38;2;122;162;247m main \x1b[0m file.go\nplain line"
	SnapCells(t, view, WithRoles(fakeTheme{Accent: accent}))
}

func TestSnapCellsUnstyled(t *testing.T) {
	out := captureCellsGolden(t, "ab\ncd")
	for _, want := range []string{`1: "ab"`, `2: "cd"`} {
		if !strings.Contains(out, want) {
			t.Errorf("cells output missing %q in:\n%s", want, out)
		}
	}
}

func TestSnapCellsPreservesCombiningGraphemes(t *testing.T) {
	out := captureCellsGolden(t, "e\u0301")
	if !strings.Contains(out, "\"e\u0301\"") {
		t.Fatalf("combining grapheme was lost from cells output: %q", out)
	}
}

func TestSnapCellsPreservesStyledCombiningGraphemes(t *testing.T) {
	out := captureCellsGolden(t, "\x1b[1me\u0301\x1b[0m")
	if !strings.Contains(out, "\"e\u0301\" [bold]") {
		t.Fatalf("styled combining grapheme was lost from cells output: %q", out)
	}
}

// captureCellsGolden runs the grid formatting without touching golden files.
func captureCellsGolden(t *testing.T, view string) string {
	t.Helper()
	buf := renderToGrid(view)

	var b strings.Builder
	for y := range buf.Height() {
		fmt.Fprintf(&b, "%d: %s\n", y+1, formatRuns(buf, y, config{roles: map[rgba]string{}}))
	}
	return b.String()
}
