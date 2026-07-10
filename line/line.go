// Package line provides small, style-preserving helpers for composing one
// terminal row. It handles visible cell widths, ANSI styling, truncation,
// alignment, repeated fill patterns, and left/right fill zones.
package line

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Align controls how Fit places content within its requested width.
type Align int

const (
	// AlignLeft places content at the left edge.
	AlignLeft Align = iota
	// AlignCenter centers content, putting an odd extra cell on the right.
	AlignCenter
	// AlignRight places content at the right edge.
	AlignRight
)

// JoinOptions configures Join.
type JoinOptions struct {
	// Gap is the minimum number of cells kept between left and right content.
	// Additional space becomes the fill zone. Negative values are treated as
	// zero.
	Gap int
	// NoEllipsis clips overlong zones without adding a truncation marker. Use
	// this for decorative rules or fill patterns where an ellipsis is not
	// meaningful.
	NoEllipsis bool
}

// Truncate reduces value to at most width cells, preserving ANSI styling and
// using an ellipsis when content is removed.
func Truncate(value string, width int) string {
	if width <= 0 {
		return ""
	}
	return ansi.Truncate(value, width, "…")
}

// Fit truncates value when necessary and pads it to exactly width cells using
// the requested alignment. Padding is unstyled so callers can wrap the result
// in their own surface style when a fill zone needs a background.
func Fit(value string, width int, align Align) string {
	if width <= 0 {
		return ""
	}
	value = Truncate(value, width)
	padding := max(0, width-ansi.StringWidth(value))
	switch align {
	case AlignRight:
		return strings.Repeat(" ", padding) + value
	case AlignCenter:
		left := padding / 2
		return strings.Repeat(" ", left) + value + strings.Repeat(" ", padding-left)
	default:
		return value + strings.Repeat(" ", padding)
	}
}

// Fill repeats pattern until width cells are occupied. A partial final
// pattern is omitted rather than split, then any remaining cells are padded
// with spaces. An empty or zero-width pattern produces spaces.
func Fill(width int, pattern string) string {
	if width <= 0 {
		return ""
	}
	patternWidth := ansi.StringWidth(pattern)
	if patternWidth <= 0 {
		return strings.Repeat(" ", width)
	}

	var b strings.Builder
	used := 0
	for used+patternWidth <= width {
		b.WriteString(pattern)
		used += patternWidth
	}
	return b.String() + strings.Repeat(" ", width-used)
}

// Join composes left and right content into exactly width cells. When space is
// tight, the left side is truncated first so the right side remains readable;
// the gap is dropped only when the right side itself needs the full width.
func Join(width int, left, right string, opts JoinOptions) string {
	if width <= 0 {
		return ""
	}

	gap := max(0, opts.Gap)
	truncate := Truncate
	if opts.NoEllipsis {
		truncate = func(value string, width int) string {
			if width <= 0 {
				return ""
			}
			return ansi.Truncate(value, width, "")
		}
	}
	left = truncate(left, width)
	right = truncate(right, width)
	lw, rw := ansi.StringWidth(left), ansi.StringWidth(right)
	if lw+rw+gap > width {
		left = truncate(left, max(0, width-rw-gap))
		lw = ansi.StringWidth(left)
	}
	if lw+rw+gap > width {
		gap = 0
		left = truncate(left, max(0, width-rw))
		lw = ansi.StringWidth(left)
	}
	if lw+rw > width {
		right = truncate(right, width)
		rw = ansi.StringWidth(right)
		left = ""
		lw = 0
	}
	return left + strings.Repeat(" ", max(0, width-lw-rw)) + right
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
