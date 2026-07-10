// Package stack provides small helpers for composing width-aware vertical
// sections such as headers, body regions, and footers. Applications still own
// section state, layout policy, and message routing.
package stack

import (
	"strings"

	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/frame"
)

// View renders one section for the width supplied by Vertical.
type View func(width int) string

// Options configures Vertical.
type Options struct {
	// Gap is the number of blank rows between adjacent non-empty sections.
	// Negative values are treated as zero.
	Gap int
	// Divider inserts a BorderMuted divider between adjacent sections after
	// the configured gap rows.
	Divider bool
}

// Vertical renders non-empty sections at the requested width, aligns every
// line to that width, and joins them from top to bottom. Empty or nil sections
// are omitted, which lets applications conditionally include a header or
// footer without manual newline handling.
func Vertical(theme gotui.Theme, width int, opts Options, views ...View) string {
	if width <= 0 {
		return ""
	}

	sections := make([]string, 0, len(views))
	for _, view := range views {
		if view == nil {
			continue
		}
		rendered := view(width)
		if rendered == "" {
			continue
		}
		sections = append(sections, fit(rendered, width))
	}
	if len(sections) == 0 {
		return ""
	}

	gap := max(0, opts.Gap)
	parts := []string{sections[0]}
	for _, section := range sections[1:] {
		for range gap {
			parts = append(parts, strings.Repeat(" ", width))
		}
		if opts.Divider {
			parts = append(parts, frame.Divider(theme, width))
		}
		parts = append(parts, section)
	}
	return strings.Join(parts, "\n")
}

func fit(view string, width int) string {
	return lipgloss.NewStyle().Width(width).Render(view)
}
