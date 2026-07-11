package tuiweave

import lipgloss "charm.land/lipgloss/v2"

// Preset describes one built-in theme. IDs are stable and suitable for
// application configuration and persistence.
type Preset struct {
	ID   string
	Name string
	Dark bool
}

type presetDefinition struct {
	Preset
	theme func() Theme
}

var presetDefinitions = [...]presetDefinition{
	{Preset: Preset{ID: "dark", Name: "Dark", Dark: true}, theme: Dark},
	{Preset: Preset{ID: "light", Name: "Light"}, theme: Light},
	{Preset: Preset{ID: "dracula", Name: "Dracula", Dark: true}, theme: Dracula},
	{Preset: Preset{ID: "nord", Name: "Nord", Dark: true}, theme: Nord},
	{Preset: Preset{ID: "catppuccin-mocha", Name: "Catppuccin Mocha", Dark: true}, theme: CatppuccinMocha},
	{Preset: Preset{ID: "catppuccin-latte", Name: "Catppuccin Latte"}, theme: CatppuccinLatte},
	{Preset: Preset{ID: "gruvbox-dark", Name: "Gruvbox Dark", Dark: true}, theme: GruvboxDark},
	{Preset: Preset{ID: "gruvbox-light", Name: "Gruvbox Light"}, theme: GruvboxLight},
}

// Presets returns the built-in themes in their stable display order. The
// returned slice is a copy and may be modified by the caller.
func Presets() []Preset {
	presets := make([]Preset, len(presetDefinitions))
	for i, definition := range presetDefinitions {
		presets[i] = definition.Preset
	}
	return presets
}

// ThemeForPreset resolves a built-in theme by stable preset ID.
func ThemeForPreset(id string) (Theme, bool) {
	for _, definition := range presetDefinitions {
		if definition.ID == id {
			return definition.theme(), true
		}
	}
	return Theme{}, false
}

// Dracula returns a theme based on the official Dracula color palette.
// Palette source: https://spec.draculatheme.com/
func Dracula() Theme {
	return Theme{
		Surface:       lipgloss.Color("#282a36"),
		SurfaceRaised: lipgloss.Color("#44475a"),
		SurfaceSunken: lipgloss.Color("#282a36"),

		Text:         lipgloss.Color("#f8f8f2"),
		TextMuted:    lipgloss.Color("#6272a4"),
		TextFaint:    lipgloss.Color("#6272a4"),
		TextInverted: lipgloss.Color("#282a36"),

		Accent:      lipgloss.Color("#bd93f9"),
		AccentMuted: lipgloss.Color("#6272a4"),
		Success:     lipgloss.Color("#50fa7b"),
		Warning:     lipgloss.Color("#f1fa8c"),
		Danger:      lipgloss.Color("#ff5555"),
		Info:        lipgloss.Color("#8be9fd"),

		Border:        lipgloss.Color("#44475a"),
		BorderFocused: lipgloss.Color("#bd93f9"),
		BorderMuted:   lipgloss.Color("#6272a4"),

		SelectionBg: lipgloss.Color("#44475a"),
		SelectionFg: lipgloss.Color("#f8f8f2"),
	}
}

// Nord returns a theme based on the official Nord color palette.
// Palette source: https://www.nordtheme.com/docs/colors-and-palettes
func Nord() Theme {
	return Theme{
		Surface:       lipgloss.Color("#2e3440"),
		SurfaceRaised: lipgloss.Color("#3b4252"),
		SurfaceSunken: lipgloss.Color("#2e3440"),

		Text:         lipgloss.Color("#eceff4"),
		TextMuted:    lipgloss.Color("#d8dee9"),
		TextFaint:    lipgloss.Color("#4c566a"),
		TextInverted: lipgloss.Color("#2e3440"),

		Accent:      lipgloss.Color("#88c0d0"),
		AccentMuted: lipgloss.Color("#5e81ac"),
		Success:     lipgloss.Color("#a3be8c"),
		Warning:     lipgloss.Color("#ebcb8b"),
		Danger:      lipgloss.Color("#bf616a"),
		Info:        lipgloss.Color("#81a1c1"),

		Border:        lipgloss.Color("#4c566a"),
		BorderFocused: lipgloss.Color("#88c0d0"),
		BorderMuted:   lipgloss.Color("#434c5e"),

		SelectionBg: lipgloss.Color("#434c5e"),
		SelectionFg: lipgloss.Color("#eceff4"),
	}
}

// CatppuccinMocha returns the darkest official Catppuccin flavor.
// Palette source: https://catppuccin.com/palette/
func CatppuccinMocha() Theme {
	return Theme{
		Surface:       lipgloss.Color("#1e1e2e"),
		SurfaceRaised: lipgloss.Color("#313244"),
		SurfaceSunken: lipgloss.Color("#181825"),

		Text:         lipgloss.Color("#cdd6f4"),
		TextMuted:    lipgloss.Color("#a6adc8"),
		TextFaint:    lipgloss.Color("#6c7086"),
		TextInverted: lipgloss.Color("#1e1e2e"),

		Accent:      lipgloss.Color("#cba6f7"),
		AccentMuted: lipgloss.Color("#45475a"),
		Success:     lipgloss.Color("#a6e3a1"),
		Warning:     lipgloss.Color("#f9e2af"),
		Danger:      lipgloss.Color("#f38ba8"),
		Info:        lipgloss.Color("#89dceb"),

		Border:        lipgloss.Color("#585b70"),
		BorderFocused: lipgloss.Color("#cba6f7"),
		BorderMuted:   lipgloss.Color("#45475a"),

		SelectionBg: lipgloss.Color("#45475a"),
		SelectionFg: lipgloss.Color("#cdd6f4"),
	}
}

// CatppuccinLatte returns the official light Catppuccin flavor.
// Palette source: https://catppuccin.com/palette/
func CatppuccinLatte() Theme {
	return Theme{
		Surface:       lipgloss.Color("#eff1f5"),
		SurfaceRaised: lipgloss.Color("#eff1f5"),
		SurfaceSunken: lipgloss.Color("#e6e9ef"),

		Text:         lipgloss.Color("#4c4f69"),
		TextMuted:    lipgloss.Color("#6c6f85"),
		TextFaint:    lipgloss.Color("#9ca0b0"),
		TextInverted: lipgloss.Color("#eff1f5"),

		Accent:      lipgloss.Color("#8839ef"),
		AccentMuted: lipgloss.Color("#bcc0cc"),
		Success:     lipgloss.Color("#40a02b"),
		Warning:     lipgloss.Color("#df8e1d"),
		Danger:      lipgloss.Color("#d20f39"),
		Info:        lipgloss.Color("#209fb5"),

		Border:        lipgloss.Color("#acb0be"),
		BorderFocused: lipgloss.Color("#8839ef"),
		BorderMuted:   lipgloss.Color("#ccd0da"),

		SelectionBg: lipgloss.Color("#bcc0cc"),
		SelectionFg: lipgloss.Color("#4c4f69"),
	}
}

// GruvboxDark returns the dark variant of the original Gruvbox palette.
// Palette source: https://github.com/morhetz/gruvbox
func GruvboxDark() Theme {
	return Theme{
		Surface:       lipgloss.Color("#282828"),
		SurfaceRaised: lipgloss.Color("#3c3836"),
		SurfaceSunken: lipgloss.Color("#1d2021"),

		Text:         lipgloss.Color("#ebdbb2"),
		TextMuted:    lipgloss.Color("#a89984"),
		TextFaint:    lipgloss.Color("#7c6f64"),
		TextInverted: lipgloss.Color("#282828"),

		Accent:      lipgloss.Color("#83a598"),
		AccentMuted: lipgloss.Color("#504945"),
		Success:     lipgloss.Color("#b8bb26"),
		Warning:     lipgloss.Color("#fabd2f"),
		Danger:      lipgloss.Color("#fb4934"),
		Info:        lipgloss.Color("#8ec07c"),

		Border:        lipgloss.Color("#665c54"),
		BorderFocused: lipgloss.Color("#83a598"),
		BorderMuted:   lipgloss.Color("#504945"),

		SelectionBg: lipgloss.Color("#504945"),
		SelectionFg: lipgloss.Color("#fbf1c7"),
	}
}

// GruvboxLight returns the light variant of the original Gruvbox palette.
// Palette source: https://github.com/morhetz/gruvbox
func GruvboxLight() Theme {
	return Theme{
		Surface:       lipgloss.Color("#fbf1c7"),
		SurfaceRaised: lipgloss.Color("#f9f5d7"),
		SurfaceSunken: lipgloss.Color("#ebdbb2"),

		Text:         lipgloss.Color("#3c3836"),
		TextMuted:    lipgloss.Color("#665c54"),
		TextFaint:    lipgloss.Color("#928374"),
		TextInverted: lipgloss.Color("#fbf1c7"),

		Accent:      lipgloss.Color("#458588"),
		AccentMuted: lipgloss.Color("#d5c4a1"),
		Success:     lipgloss.Color("#79740e"),
		Warning:     lipgloss.Color("#b57614"),
		Danger:      lipgloss.Color("#9d0006"),
		Info:        lipgloss.Color("#427b58"),

		Border:        lipgloss.Color("#bdae93"),
		BorderFocused: lipgloss.Color("#458588"),
		BorderMuted:   lipgloss.Color("#d5c4a1"),

		SelectionBg: lipgloss.Color("#d5c4a1"),
		SelectionFg: lipgloss.Color("#3c3836"),
	}
}
