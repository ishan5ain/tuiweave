package tuiweave

import (
	"fmt"
	"image/color"
	"strings"
	"testing"
)

func TestPresetCatalog(t *testing.T) {
	want := []Preset{
		{ID: "dark", Name: "Dark", Dark: true},
		{ID: "light", Name: "Light"},
		{ID: "dracula", Name: "Dracula", Dark: true},
		{ID: "nord", Name: "Nord", Dark: true},
		{ID: "catppuccin-mocha", Name: "Catppuccin Mocha", Dark: true},
		{ID: "catppuccin-latte", Name: "Catppuccin Latte"},
		{ID: "gruvbox-dark", Name: "Gruvbox Dark", Dark: true},
		{ID: "gruvbox-light", Name: "Gruvbox Light"},
	}

	got := Presets()
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("Presets() = %#v, want %#v", got, want)
	}
	seen := make(map[string]bool, len(got))
	for _, preset := range got {
		if preset.ID == "" || preset.Name == "" {
			t.Errorf("preset has empty metadata: %#v", preset)
		}
		if seen[preset.ID] {
			t.Errorf("duplicate preset ID %q", preset.ID)
		}
		seen[preset.ID] = true
		if _, ok := ThemeForPreset(preset.ID); !ok {
			t.Errorf("ThemeForPreset(%q) was not found", preset.ID)
		}
	}

	got[0].Name = "changed"
	if Presets()[0].Name != "Dark" {
		t.Fatal("Presets returned shared mutable metadata")
	}
	if theme, ok := ThemeForPreset("missing"); ok || theme != (Theme{}) {
		t.Fatalf("unknown preset = (%#v, %v), want zero theme and false", theme, ok)
	}
}

func TestPresetColors(t *testing.T) {
	tests := []struct {
		id   string
		make func() Theme
		want string
	}{
		{"dark", Dark, "#16161d #1f1f28 #101014 #e6e6f0 #9a9ab0 #5c5c70 #16161d #7aa2f7 #3d59a1 #9ece6a #e0af68 #f7768e #7dcfff #3b3b4d #7aa2f7 #2a2a38 #2e3c64 #e6e6f0"},
		{"light", Light, "#fafafa #ffffff #f0f0f4 #2a2a33 #6e6e80 #a0a0b0 #fafafa #3760bf #99a7df #587539 #8f5e15 #c64343 #188092 #d0d0da #3760bf #e2e2ea #c4d1f5 #2a2a33"},
		{"dracula", Dracula, "#282a36 #44475a #282a36 #f8f8f2 #6272a4 #6272a4 #282a36 #bd93f9 #6272a4 #50fa7b #f1fa8c #ff5555 #8be9fd #44475a #bd93f9 #6272a4 #44475a #f8f8f2"},
		{"nord", Nord, "#2e3440 #3b4252 #2e3440 #eceff4 #d8dee9 #4c566a #2e3440 #88c0d0 #5e81ac #a3be8c #ebcb8b #bf616a #81a1c1 #4c566a #88c0d0 #434c5e #434c5e #eceff4"},
		{"catppuccin-mocha", CatppuccinMocha, "#1e1e2e #313244 #181825 #cdd6f4 #a6adc8 #6c7086 #1e1e2e #cba6f7 #45475a #a6e3a1 #f9e2af #f38ba8 #89dceb #585b70 #cba6f7 #45475a #45475a #cdd6f4"},
		{"catppuccin-latte", CatppuccinLatte, "#eff1f5 #eff1f5 #e6e9ef #4c4f69 #6c6f85 #9ca0b0 #eff1f5 #8839ef #bcc0cc #40a02b #df8e1d #d20f39 #209fb5 #acb0be #8839ef #ccd0da #bcc0cc #4c4f69"},
		{"gruvbox-dark", GruvboxDark, "#282828 #3c3836 #1d2021 #ebdbb2 #a89984 #7c6f64 #282828 #83a598 #504945 #b8bb26 #fabd2f #fb4934 #8ec07c #665c54 #83a598 #504945 #504945 #fbf1c7"},
		{"gruvbox-light", GruvboxLight, "#fbf1c7 #f9f5d7 #ebdbb2 #3c3836 #665c54 #928374 #fbf1c7 #458588 #d5c4a1 #79740e #b57614 #9d0006 #427b58 #bdae93 #458588 #d5c4a1 #d5c4a1 #3c3836"},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			theme := tt.make()
			if got := themeColors(theme); got != tt.want {
				t.Fatalf("colors = %q\nwant   = %q", got, tt.want)
			}
			resolved, ok := ThemeForPreset(tt.id)
			if !ok || themeColors(resolved) != tt.want {
				t.Fatalf("catalog theme = (%q, %v), want (%q, true)", themeColors(resolved), ok, tt.want)
			}
		})
	}
}

func themeColors(theme Theme) string {
	colors := []color.Color{
		theme.Surface, theme.SurfaceRaised, theme.SurfaceSunken,
		theme.Text, theme.TextMuted, theme.TextFaint, theme.TextInverted,
		theme.Accent, theme.AccentMuted, theme.Success, theme.Warning, theme.Danger, theme.Info,
		theme.Border, theme.BorderFocused, theme.BorderMuted,
		theme.SelectionBg, theme.SelectionFg,
	}
	values := make([]string, len(colors))
	for i, c := range colors {
		if c == nil {
			values[i] = "<nil>"
			continue
		}
		r, g, b, a := c.RGBA()
		if a != 0xffff {
			values[i] = fmt.Sprintf("#%02x%02x%02x/%04x", r>>8, g>>8, b>>8, a)
			continue
		}
		values[i] = fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
	}
	return strings.Join(values, " ")
}
