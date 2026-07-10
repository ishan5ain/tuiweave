// Package splitpane provides small helpers for composing sibling views into a
// horizontal split. It owns geometry, natural-height alignment, and the
// themed divider; applications still own the views, state, and routing.
package splitpane

import (
	"strings"

	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/layout"
)

// View renders one pane for the width supplied by Horizontal.
type View func(width int) string

// Options configures Horizontal.
type Options struct {
	// Ratio is the percentage of the available pane width assigned to the
	// left view. Zero means 50%; values outside 1–99 are clamped.
	Ratio int
	// Gap is the total number of cells between panes. Zero means one cell;
	// negative values disable the gap and divider. Positive gaps center a
	// themed vertical divider within the gap.
	Gap int
}

// Horizontal splits width between left and right views, renders each callback
// at its assigned width, aligns them to the taller natural height, and joins
// them with a BorderMuted divider. It returns an empty string when the width
// cannot provide at least one cell for both panes.
func Horizontal(theme gotui.Theme, width int, opts Options, left, right View) string {
	if width <= 0 || left == nil || right == nil {
		return ""
	}

	ratio := opts.Ratio
	if ratio == 0 {
		ratio = 50
	}
	ratio = max(1, min(ratio, 99))

	gap := opts.Gap
	if gap == 0 {
		gap = 1
	}
	if gap < 0 {
		gap = 0
	}
	if width <= gap+1 {
		return ""
	}

	var leftRect, rightRect layout.Rect
	layout.Horizontal(layout.Percent(ratio), layout.Fill(1)).
		WithSpacing(gap).
		Split(layout.NewRect(0, 0, width, 1)).
		Assign(&leftRect, &rightRect)
	if leftRect.Dx() <= 0 || rightRect.Dx() <= 0 {
		return ""
	}

	leftView := left(leftRect.Dx())
	rightView := right(rightRect.Dx())
	height := max(1, max(lipgloss.Height(leftView), lipgloss.Height(rightView)))
	leftView = fit(leftView, leftRect.Dx(), height)
	rightView = fit(rightView, rightRect.Dx(), height)
	return lipgloss.JoinHorizontal(lipgloss.Top,
		leftView,
		divider(theme, gap, height),
		rightView,
	)
}

func fit(view string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	return lipgloss.NewStyle().Width(width).Height(height).Render(view)
}

func divider(theme gotui.Theme, gap, height int) string {
	if gap <= 0 || height <= 0 {
		return ""
	}
	line := strings.Repeat(" ", gap)
	center := gap / 2
	line = line[:center] + "│" + line[center+1:]
	return lipgloss.NewStyle().
		Foreground(theme.BorderMuted).
		Render(strings.TrimSuffix(strings.Repeat(line+"\n", height), "\n"))
}
