// Package scrollbar renders a vertical scroll indicator for any scrolling
// component.
//
// Components stay unaware of scrollbars: they expose scroll stats via the
// Scrollable interface, and the app places the bar as a one-column layout
// segment beside the component:
//
//	layout.Horizontal(layout.Fill(1), layout.Len(1)).
//	    Split(area).Assign(&content, &bar)
//	m.list.SetSize(content.Dx(), content.Dy())
//	...
//	lipgloss.JoinHorizontal(lipgloss.Top, m.list.View(), scrollbar.For(theme, m.list))
package scrollbar

import (
	"strings"

	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
)

// Scrollable is implemented by gotui's scrolling components (viewport, list,
// table, chat, diffview).
type Scrollable interface {
	// TotalLines is the total scrollable extent, in lines/rows.
	TotalLines() int
	// VisibleLines is how many lines are shown at once.
	VisibleLines() int
	// YOffset is the index of the first visible line.
	YOffset() int
}

// For renders a bar matching s, one cell wide and VisibleLines tall.
func For(theme gotui.Theme, s Scrollable) string {
	return Vertical(theme, s.VisibleLines(), s.TotalLines(), s.VisibleLines(), s.YOffset())
}

// Vertical renders a one-column scrollbar of the given height for content of
// total lines, visible of which are shown starting at offset. When
// everything fits, the whole bar renders as track.
func Vertical(theme gotui.Theme, height, total, visible, offset int) string {
	if height <= 0 {
		return ""
	}
	track := lipgloss.NewStyle().Foreground(theme.BorderMuted).Render("│")
	rows := make([]string, height)

	if total <= visible || total <= 0 {
		for i := range rows {
			rows[i] = track
		}
		return strings.Join(rows, "\n")
	}

	thumbLen := max(1, height*visible/total)
	maxOffset := total - visible
	offset = max(0, min(offset, maxOffset))
	thumbStart := (height - thumbLen) * offset / maxOffset

	thumb := lipgloss.NewStyle().Foreground(theme.Border).Render("┃")
	for i := range rows {
		if i >= thumbStart && i < thumbStart+thumbLen {
			rows[i] = thumb
		} else {
			rows[i] = track
		}
	}
	return strings.Join(rows, "\n")
}
