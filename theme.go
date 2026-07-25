package tuiweave

import (
	"image/color"

	lipgloss "charm.land/lipgloss/v2"
)

// Theme is the set of semantic color roles that every tuiweave component
// consumes. Components derive their styles exclusively from these roles and
// never use color literals, so swapping the Theme restyles an entire app
// consistently.
//
// Applications that want a role to inherit the terminal or parent color may
// assign lipgloss.NoColor{} to that role on a copy of a preset Theme. This is
// an absence of emitted color, not a separate semantic role; components still
// derive all other styling from the Theme they receive.
//
// Adding a role is an API event: it requires justification that no existing
// role covers the semantics. This is what keeps the theme surface from
// growing with every new component.
type Theme struct {
	// Surfaces.
	Surface       color.Color // app background
	SurfaceRaised color.Color // panels, dialogs, popovers
	SurfaceSunken color.Color // input wells, code blocks

	// Text.
	Text         color.Color // primary text
	TextMuted    color.Color // secondary text, placeholders, help
	TextFaint    color.Color // tertiary text, timestamps, subtle separators
	TextInverted color.Color // text rendered on accent/intent fills

	// Intent.
	Accent      color.Color // interactive/brand emphasis
	AccentMuted color.Color // subdued accent fills and highlights
	Success     color.Color
	Warning     color.Color
	Danger      color.Color
	Info        color.Color

	// Chrome.
	Border        color.Color // default borders
	BorderFocused color.Color // border of the focused element
	BorderMuted   color.Color // subtle dividers

	// Selection.
	SelectionBg color.Color // background of selected rows/items
	SelectionFg color.Color // foreground of selected rows/items
}

// Dark is the default dark theme.
func Dark() Theme {
	return Theme{
		Surface:       lipgloss.Color("#16161d"),
		SurfaceRaised: lipgloss.Color("#1f1f28"),
		SurfaceSunken: lipgloss.Color("#101014"),

		Text:         lipgloss.Color("#e6e6f0"),
		TextMuted:    lipgloss.Color("#9a9ab0"),
		TextFaint:    lipgloss.Color("#5c5c70"),
		TextInverted: lipgloss.Color("#16161d"),

		Accent:      lipgloss.Color("#7aa2f7"),
		AccentMuted: lipgloss.Color("#3d59a1"),
		Success:     lipgloss.Color("#9ece6a"),
		Warning:     lipgloss.Color("#e0af68"),
		Danger:      lipgloss.Color("#f7768e"),
		Info:        lipgloss.Color("#7dcfff"),

		Border:        lipgloss.Color("#3b3b4d"),
		BorderFocused: lipgloss.Color("#7aa2f7"),
		BorderMuted:   lipgloss.Color("#2a2a38"),

		SelectionBg: lipgloss.Color("#2e3c64"),
		SelectionFg: lipgloss.Color("#e6e6f0"),
	}
}

// Light is the default light theme.
func Light() Theme {
	return Theme{
		Surface:       lipgloss.Color("#fafafa"),
		SurfaceRaised: lipgloss.Color("#ffffff"),
		SurfaceSunken: lipgloss.Color("#f0f0f4"),

		Text:         lipgloss.Color("#2a2a33"),
		TextMuted:    lipgloss.Color("#6e6e80"),
		TextFaint:    lipgloss.Color("#a0a0b0"),
		TextInverted: lipgloss.Color("#fafafa"),

		Accent:      lipgloss.Color("#3760bf"),
		AccentMuted: lipgloss.Color("#99a7df"),
		Success:     lipgloss.Color("#587539"),
		Warning:     lipgloss.Color("#8f5e15"),
		Danger:      lipgloss.Color("#c64343"),
		Info:        lipgloss.Color("#188092"),

		Border:        lipgloss.Color("#d0d0da"),
		BorderFocused: lipgloss.Color("#3760bf"),
		BorderMuted:   lipgloss.Color("#e2e2ea"),

		SelectionBg: lipgloss.Color("#c4d1f5"),
		SelectionFg: lipgloss.Color("#2a2a33"),
	}
}
