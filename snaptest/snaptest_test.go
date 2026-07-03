package snaptest

import (
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	in := "header   \nbody\t\nfooter"
	want := "header\nbody\t\nfooter\n"
	if got := normalize(in); got != want {
		t.Errorf("normalize(%q) = %q, want %q", in, got, want)
	}
}

func TestGoldenPathSanitizesSubtests(t *testing.T) {
	t.Run("with spaces/and slashes", func(t *testing.T) {
		got := goldenPath(t, ".golden")
		if strings.ContainsAny(got[len("testdata/"):], "/ ") {
			t.Errorf("goldenPath produced unsanitized name: %q", got)
		}
	})
}

func TestDiffLinesReportsMismatches(t *testing.T) {
	out := diffLines("a\nb\nc\n", "a\nX\nc\nextra\n")
	for _, wantFragment := range []string{"line 2:", `want: "b"`, `got:  "X"`, "line 4:", "<missing line>"} {
		if !strings.Contains(out, wantFragment) {
			t.Errorf("diff output missing %q in:\n%s", wantFragment, out)
		}
	}
	if strings.Contains(out, "line 1:") || strings.Contains(out, "line 3:") {
		t.Errorf("diff output reports matching lines:\n%s", out)
	}
}

// TestSnapGolden exercises the real golden workflow end to end; its files in
// testdata/ are generated with -update and checked in.
func TestSnapGolden(t *testing.T) {
	view := "┌ demo ─┐   \n│ hello │\n└───────┘"
	Snap(t, view)
}

func TestSnapStyledGolden(t *testing.T) {
	// Hand-written ANSI (bold red "hi") keeps this deterministic across
	// environments, independent of any color-profile detection.
	view := "\x1b[1;31mhi\x1b[0m there"
	SnapStyled(t, view)
}
