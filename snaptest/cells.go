package snaptest

import (
	"fmt"
	"image/color"
	"reflect"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui/internal/grapheme"
)

// Option configures SnapCells.
type Option func(*config)

type config struct {
	// roles maps a color (by RGBA) to a human-readable name.
	roles map[rgba]string
}

type rgba [4]uint32

func colorKey(c color.Color) rgba {
	r, g, b, a := c.RGBA()
	return rgba{r, g, b, a}
}

// WithRoles labels colors in the cells golden with the names of matching
// color.Color fields from the given struct, instead of hex values. Pass a
// theme (e.g. gotui.Dark()) so runs read like [fg=Accent bold] rather than
// [fg=#7aa2f7 bold].
func WithRoles(theme any) Option {
	return func(cfg *config) {
		v := reflect.ValueOf(theme)
		for v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			return
		}
		colorType := reflect.TypeFor[color.Color]()
		for i := range v.NumField() {
			f := v.Type().Field(i)
			if !f.IsExported() || !f.Type.Implements(colorType) {
				continue
			}
			c, ok := v.Field(i).Interface().(color.Color)
			if !ok || c == nil {
				continue
			}
			key := colorKey(c)
			// First role wins so duplicate colors label deterministically.
			if _, exists := cfg.roles[key]; !exists {
				cfg.roles[key] = f.Name
			}
		}
	}
}

// SnapCells parses the rendered view into a terminal cell grid and snapshots
// a per-line listing of style runs against testdata/<TestName>.cells.golden.
// Each run is the cell text followed by its style, e.g.:
//
//	1: " main " [fg=TextInverted bg=Accent bold] | " 80x24 " [fg=TextMuted bg=SurfaceRaised]
//
// This is the artifact for asserting *which role* styles what, where the
// plain golden asserts layout and the styled golden asserts exact bytes.
func SnapCells(t *testing.T, view string, opts ...Option) {
	t.Helper()

	cfg := config{roles: map[rgba]string{}}
	for _, opt := range opts {
		opt(&cfg)
	}

	buf := renderToGrid(view)

	var b strings.Builder
	for y := range buf.Height() {
		fmt.Fprintf(&b, "%d: %s\n", y+1, formatRuns(buf, y, cfg))
	}
	compare(t, goldenPath(t, ".cells.golden"), b.String())
}

// renderToGrid parses a rendered ANSI string into a cell grid.
func renderToGrid(view string) uv.ScreenBuffer {
	protected, replacements := grapheme.Protect(view)
	ss := uv.NewStyledString(protected)
	bounds := ss.Bounds()
	buf := uv.NewScreenBuffer(bounds.Dx(), bounds.Dy())
	// Snapshots must retain the complete grapheme in each cell. The buffer's
	// default WcWidth decoder represents combining marks as separate width-zero
	// cells, and its ASCII fast path can overwrite them with following padding.
	buf.Method = ansi.GraphemeWidth
	ss.Draw(buf, buf.Bounds())
	for y := range buf.Height() {
		for x := range buf.Width() {
			cell := buf.CellAt(x, y)
			if cell == nil {
				continue
			}
			if original, ok := replacements[cell.Content]; ok {
				cell.Content = original
			}
		}
	}
	return buf
}

// run is a horizontal stretch of cells sharing one style.
type run struct {
	text  strings.Builder
	style uv.Style
}

func formatRuns(buf uv.ScreenBuffer, y int, cfg config) string {
	var runs []*run
	var cur *run
	for x := range buf.Width() {
		cell := buf.CellAt(x, y)
		if cell == nil {
			continue // out of bounds
		}
		if cell.Width == 0 {
			if cell.Content == "" {
				continue // continuation of a wide cell
			}
			// ultraviolet may store a combining mark as a non-empty
			// width-zero cell after its base cell. Keep it in the style run
			// instead of treating it like a wide-cell continuation.
			if cur == nil || !cur.style.Equal(&cell.Style) {
				cur = &run{style: cell.Style}
				runs = append(runs, cur)
			}
			cur.text.WriteString(cell.Content)
			continue
		}
		content := cell.Content
		if content == "" {
			content = " "
		}
		if cur == nil || !cur.style.Equal(&cell.Style) {
			cur = &run{style: cell.Style}
			runs = append(runs, cur)
		}
		cur.text.WriteString(content)
	}

	parts := make([]string, 0, len(runs))
	for i, r := range runs {
		text := r.text.String()
		// Trailing unstyled whitespace is buffer padding, not UI: trim it
		// from the last run (styled trailing spaces are real, e.g. bar fills).
		if i == len(runs)-1 && r.style.IsZero() {
			text = strings.TrimRight(text, " ")
			if text == "" {
				continue
			}
		}
		s := fmt.Sprintf("%q", text)
		if desc := describeStyle(r.style, cfg); desc != "" {
			s += " [" + desc + "]"
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, " | ")
}

func describeStyle(s uv.Style, cfg config) string {
	var parts []string
	if s.Fg != nil {
		parts = append(parts, "fg="+colorName(s.Fg, cfg))
	}
	if s.Bg != nil {
		parts = append(parts, "bg="+colorName(s.Bg, cfg))
	}
	for _, attr := range []struct {
		bit  uint8
		name string
	}{
		{uv.AttrBold, "bold"},
		{uv.AttrFaint, "faint"},
		{uv.AttrItalic, "italic"},
		{uv.AttrBlink, "blink"},
		{uv.AttrReverse, "reverse"},
		{uv.AttrConceal, "conceal"},
		{uv.AttrStrikethrough, "strikethrough"},
	} {
		if s.Attrs&attr.bit != 0 {
			parts = append(parts, attr.name)
		}
	}
	if s.Underline != 0 {
		parts = append(parts, "underline")
	}
	return strings.Join(parts, " ")
}

func colorName(c color.Color, cfg config) string {
	key := colorKey(c)
	if name, ok := cfg.roles[key]; ok {
		return name
	}
	return fmt.Sprintf("#%02x%02x%02x", uint8(key[0]>>8), uint8(key[1]>>8), uint8(key[2]>>8))
}
