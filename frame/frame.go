// Package frame provides small, domain-neutral decoration helpers for
// composing tuiweave views. The helpers return styled strings; applications still
// own layout, state, and message routing.
package frame

import (
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/layout"
)

// BadgeKind selects the semantic roles used by Badge.
type BadgeKind int

const (
	// BadgeAccent is for a mode, identity, or other primary emphasis.
	BadgeAccent BadgeKind = iota
	// BadgeMuted is for subdued metadata that benefits from a contained fill.
	BadgeMuted
	// BadgeSuccess, BadgeWarning, BadgeDanger, and BadgeInfo communicate intent.
	BadgeSuccess
	BadgeWarning
	BadgeDanger
	BadgeInfo
)

// PanelOptions controls the decoration applied by Panel.
type PanelOptions struct {
	// Title is rendered into the top border. An empty title leaves the border
	// uninterrupted.
	Title string
	// Focused uses the theme's focused border role and accents the title.
	Focused bool
	// Padding is applied to the content on all sides. Negative values are
	// treated as zero.
	Padding int
}

// PanelContentRect returns the rectangle available to content inside a panel
// occupying area. It accounts for the rounded border and symmetric padding,
// using the same inset contract as Panel. The returned rectangle keeps the
// original origin so callers can also use it for overlay or inspection
// coordinates.
func PanelContentRect(area layout.Rect, opts PanelOptions) layout.Rect {
	padding := panelPadding(area.Dx(), opts)
	return layout.NewRect(
		area.Min.X+1+padding,
		area.Min.Y+1+padding,
		max(0, area.Dx()-2-2*padding),
		max(0, area.Dy()-2-2*padding),
	)
}

// Panel returns content enclosed in a rounded, theme-styled frame with the
// requested total width. The returned string is empty for widths below two;
// otherwise every rendered line is exactly width cells wide.
//
// Content may already contain ANSI styling. The panel supplies the surface
// and default text roles while preserving nested styles.
func Panel(theme tuiweave.Theme, content string, width int, opts PanelOptions) string {
	if width < 2 {
		return ""
	}

	borderColor := theme.Border
	if opts.Focused {
		borderColor = theme.BorderFocused
	}

	border := lipgloss.NewStyle().Foreground(borderColor)
	innerWidth := width - 2
	padding := panelPadding(width, opts)
	contentWidth := innerWidth - 2*padding
	if contentWidth <= 0 {
		content = ""
	}

	body := lipgloss.NewStyle().
		Width(innerWidth).
		Padding(padding).
		Background(theme.SurfaceRaised).
		Foreground(theme.Text).
		Render(content)
	bodyLines := strings.Split(body, "\n")
	for i, line := range bodyLines {
		bodyLines[i] = border.Render(string(lipgloss.RoundedBorder().Left)) +
			line +
			border.Render(string(lipgloss.RoundedBorder().Right))
	}

	return strings.Join([]string{
		renderTop(theme, border, opts, innerWidth),
		strings.Join(bodyLines, "\n"),
		border.Render(string(lipgloss.RoundedBorder().BottomLeft)) + border.Render(strings.Repeat(lipgloss.RoundedBorder().Bottom, innerWidth)) + border.Render(string(lipgloss.RoundedBorder().BottomRight)),
	}, "\n")
}

func panelPadding(width int, opts PanelOptions) int {
	innerWidth := max(0, width-2)
	return min(max(0, opts.Padding), innerWidth/2)
}

func renderTop(theme tuiweave.Theme, border lipgloss.Style, opts PanelOptions, innerWidth int) string {
	rounded := lipgloss.RoundedBorder()
	if opts.Title == "" || innerWidth < 3 {
		return border.Render(string(rounded.TopLeft) + strings.Repeat(rounded.Top, innerWidth) + rounded.TopRight)
	}

	maxTitle := max(0, innerWidth-2)
	title := ansi.Truncate(opts.Title, maxTitle, "…")
	title = " " + title + " "
	remaining := innerWidth - lipgloss.Width(title)
	if remaining < 0 {
		remaining = 0
	}
	titleColor := theme.TextMuted
	if opts.Focused {
		titleColor = theme.Accent
	}
	return border.Render(string(rounded.TopLeft)) +
		lipgloss.NewStyle().Foreground(titleColor).Bold(opts.Focused).Render(title) +
		border.Render(strings.Repeat(rounded.Top, remaining)+rounded.TopRight)
}

// Divider returns a width-aware subtle horizontal separator.
func Divider(theme tuiweave.Theme, width int) string {
	if width <= 0 {
		return ""
	}
	return lipgloss.NewStyle().Foreground(theme.BorderMuted).Render(strings.Repeat("─", width))
}

// Badge returns a compact, padded label styled from semantic theme roles.
func Badge(theme tuiweave.Theme, text string, kind BadgeKind) string {
	background := theme.Accent
	foreground := theme.TextInverted
	switch kind {
	case BadgeMuted:
		background = theme.AccentMuted
	case BadgeSuccess:
		background = theme.Success
	case BadgeWarning:
		background = theme.Warning
	case BadgeDanger:
		background = theme.Danger
	case BadgeInfo:
		background = theme.Info
	}
	return lipgloss.NewStyle().
		Background(background).
		Foreground(foreground).
		Bold(true).
		Padding(0, 1).
		Render(text)
}
