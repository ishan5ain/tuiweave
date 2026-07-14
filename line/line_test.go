package line

import (
	"strings"
	"testing"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave/snaptest"
)

func TestLineHelpersGolden(t *testing.T) {
	accent := lipgloss.NewStyle().Bold(true).Render("main.go")
	view := strings.Join([]string{
		Truncate("a very long title", 10),
		Fit(accent, 16, AlignLeft),
		Fit("center", 16, AlignCenter),
		Fit("12:45", 16, AlignRight),
		Join(28, "main.go", "Ln 12  Col 4", JoinOptions{Gap: 1}),
		Fill(12, "─"),
	}, "\n")
	snaptest.Snap(t, view)
}

func TestLineHelpersPreserveWidth(t *testing.T) {
	styled := lipgloss.NewStyle().Bold(true).Render("wide text")
	for width := 1; width <= 32; width++ {
		for _, value := range []string{
			Truncate(styled, width),
			Fit(styled, width, AlignLeft),
			Fit(styled, width, AlignCenter),
			Fit(styled, width, AlignRight),
			Fill(width, "─"),
			Join(width, styled, "right", JoinOptions{Gap: 1}),
		} {
			if got := ansi.StringWidth(value); got != width && value != Truncate(styled, width) {
				t.Fatalf("width %d rendered as %d: %q", width, got, value)
			}
		}
	}
}

func TestLineJoinKeepsRightSideWhenNarrow(t *testing.T) {
	if got := ansi.Strip(Join(12, "a long filename", "Ln 9", JoinOptions{Gap: 1})); got != "a long… Ln 9" {
		t.Fatalf("narrow join = %q, want %q", got, "a long… Ln 9")
	}
	if got := Join(4, "left", "right", JoinOptions{Gap: 1}); ansi.StringWidth(got) != 4 {
		t.Fatalf("fully narrow join width = %d, want 4", ansi.StringWidth(got))
	}
}

func TestLineEmptyAndInvalidWidths(t *testing.T) {
	for _, width := range []int{0, -1} {
		if got := Truncate("value", width); got != "" {
			t.Errorf("Truncate(%d) = %q, want empty", width, got)
		}
		if got := Fit("value", width, AlignLeft); got != "" {
			t.Errorf("Fit(%d) = %q, want empty", width, got)
		}
		if got := Fill(width, "─"); got != "" {
			t.Errorf("Fill(%d) = %q, want empty", width, got)
		}
		if got := Join(width, "left", "right", JoinOptions{}); got != "" {
			t.Errorf("Join(%d) = %q, want empty", width, got)
		}
	}
}

func TestLineUnicodeNarrowBoundaries(t *testing.T) {
	values := []string{
		"界e\u0301",
		"👩‍💻 developer",
		"a very long wide 名称",
	}
	for _, width := range []int{0, 1, 2, 3, 4, 8, 16} {
		for _, value := range values {
			truncated := Truncate(value, width)
			if got := ansi.StringWidth(truncated); got > width {
				t.Fatalf("Truncate(%q, %d) width = %d", value, width, got)
			}

			for _, align := range []Align{AlignLeft, AlignCenter, AlignRight} {
				fitted := Fit(value, width, align)
				if got := ansi.StringWidth(fitted); got != width {
					t.Fatalf("Fit(%q, %d, %d) width = %d, want %d", value, width, align, got, width)
				}
			}

			joined := Join(width, value, "終", JoinOptions{Gap: 1})
			if got := ansi.StringWidth(joined); got != width {
				t.Fatalf("Join(%q, %d) width = %d, want %d", value, width, got, width)
			}
		}

		for _, pattern := range []string{"界", "e\u0301", "👩‍💻"} {
			filled := Fill(width, pattern)
			if got := ansi.StringWidth(filled); got != width {
				t.Fatalf("Fill(%q, %d) width = %d, want %d", pattern, width, got, width)
			}
		}
	}
}

func TestLineUnicodeTruncationGolden(t *testing.T) {
	view := strings.Join([]string{
		Truncate("界e\u0301👩‍💻 wide", 9),
		Fit("界e\u0301", 8, AlignCenter),
		Join(14, "👩‍💻 developer", "終", JoinOptions{Gap: 1}),
		Fill(8, "界"),
	}, "\n")
	snaptest.Snap(t, view)
}

func TestLineNarrowEmojiDoesNotSplit(t *testing.T) {
	for _, test := range []struct {
		width int
		want  string
	}{
		{width: 1, want: "…"},
		{width: 2, want: "👩‍💻"},
	} {
		if got := ansi.Strip(Truncate("👩‍💻", test.width)); got != test.want {
			t.Fatalf("Truncate emoji at width %d = %q, want %q", test.width, got, test.want)
		}
		if got := ansi.Strip(Fit("👩‍💻", test.width, AlignLeft)); ansi.StringWidth(got) != test.width {
			t.Fatalf("Fit emoji at width %d rendered width %d", test.width, ansi.StringWidth(got))
		}
	}
}
