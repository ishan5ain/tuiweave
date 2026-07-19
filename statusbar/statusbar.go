// Package statusbar provides a one-line status bar with themed segments
// aligned left and right, in the style of editor status lines.
//
// The bar is passive: it renders state set via SetLeft/SetRight and handles
// no messages itself. Lay it out with layout.Len(1).
package statusbar

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/layout"
)

// Kind selects which theme roles style a segment.
type Kind int

const (
	// KindNormal renders primary text on the bar surface.
	KindNormal Kind = iota
	// KindAccent renders a bold badge on an accent fill — the "mode"
	// segment. Use at most one per side.
	KindAccent
	// KindMuted renders secondary text; the default for informational
	// segments like positions and counts.
	KindMuted
	// KindSuccess, KindWarning, KindDanger, and KindInfo color the segment
	// text with the matching intent role.
	KindSuccess
	KindWarning
	KindDanger
	KindInfo
)

// Segment is one unit of statusbar content.
type Segment struct {
	Text string
	Kind Kind
}

// Model is a statusbar component. Create one with New.
type Model struct {
	width, height int
	left, right   []Segment

	bar   lipgloss.Style
	kinds map[Kind]lipgloss.Style
}

// New returns a statusbar styled from the theme's roles.
func New(theme tuiweave.Theme) Model {
	base := lipgloss.NewStyle().Background(theme.SurfaceRaised)
	return Model{
		bar: base,
		kinds: map[Kind]lipgloss.Style{
			KindNormal: base.Foreground(theme.Text),
			KindAccent: lipgloss.NewStyle().
				Background(theme.Accent).
				Foreground(theme.TextInverted).
				Bold(true),
			KindMuted:   base.Foreground(theme.TextMuted),
			KindSuccess: base.Foreground(theme.Success),
			KindWarning: base.Foreground(theme.Warning),
			KindDanger:  base.Foreground(theme.Danger),
			KindInfo:    base.Foreground(theme.Info),
		},
	}
}

// SetTheme rebuilds all styles from the theme's roles, preserving
// non-style state such as the left/right segments.
func (m *Model) SetTheme(theme tuiweave.Theme) {
	base := lipgloss.NewStyle().Background(theme.SurfaceRaised)
	m.bar = base
	m.kinds = map[Kind]lipgloss.Style{
		KindNormal: base.Foreground(theme.Text),
		KindAccent: lipgloss.NewStyle().
			Background(theme.Accent).
			Foreground(theme.TextInverted).
			Bold(true),
		KindMuted:   base.Foreground(theme.TextMuted),
		KindSuccess: base.Foreground(theme.Success),
		KindWarning: base.Foreground(theme.Warning),
		KindDanger:  base.Foreground(theme.Danger),
		KindInfo:    base.Foreground(theme.Info),
	}
}

// SetSize sets the box the bar renders in. The bar is one line tall; height
// only matters as zero (render nothing) or nonzero.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
}

// SizeMode reports that the status bar is width-constrained with a natural
// height of one row.
func (m Model) SizeMode() layout.SizeMode { return layout.SizeWidthBounded }

// SetLeft replaces the segments aligned to the left edge.
func (m *Model) SetLeft(segments ...Segment) {
	m.left = segments
}

// SetRight replaces the segments aligned to the right edge.
func (m *Model) SetRight(segments ...Segment) {
	m.right = segments
}

// Update implements the tuiweave component contract. The statusbar handles no
// messages.
func (m Model) Update(_ tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// View renders the bar. When space is tight, right segments are dropped
// first, then the left side is truncated with an ellipsis.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	left := m.renderSegments(m.left)
	right := m.renderSegments(m.right)
	lw, rw := lipgloss.Width(left), lipgloss.Width(right)

	if lw+rw > m.width {
		right, rw = "", 0
	}
	if lw > m.width {
		left = ansi.Truncate(left, m.width, "…")
		lw = lipgloss.Width(left)
	}

	fill := m.bar.Render(strings.Repeat(" ", m.width-lw-rw))
	return left + fill + right
}

func (m Model) renderSegments(segments []Segment) string {
	var b strings.Builder
	for _, seg := range segments {
		b.WriteString(m.kinds[seg.Kind].Render(" " + seg.Text + " "))
	}
	return b.String()
}
