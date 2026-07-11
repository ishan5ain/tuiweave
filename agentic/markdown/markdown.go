// Package markdown renders markdown to styled terminal text using the
// theme's roles.
//
// Renderer is the seam that keeps the rest of the library (and apps)
// independent of the rendering engine: today's implementation wraps glamour
// and re-renders the full source per call, a purpose-built streaming
// renderer can replace it later without touching consumers. Streaming
// callers re-render as deltas arrive; cache accordingly (chat's assistant
// cell does).
package markdown

import (
	"fmt"
	"image/color"
	"strings"

	glamour "charm.land/glamour/v2"
	"charm.land/glamour/v2/ansi"

	"github.com/ishan5ain/tuiweave"
)

// Renderer renders markdown source to styled terminal text wrapped at width.
// Implementations must be deterministic: same source and width, same output.
type Renderer interface {
	Render(source string, width int) (string, error)
}

// Sprint renders markdown, falling back to the raw source when rendering
// fails — transcript UIs should degrade, not error.
func Sprint(r Renderer, source string, width int) string {
	out, err := r.Render(source, width)
	if err != nil {
		return source
	}
	return strings.TrimRight(out, "\n")
}

// NewRenderer returns the default glamour-backed Renderer, styled from the
// theme's roles.
func NewRenderer(theme tuiweave.Theme) Renderer {
	return &glamourRenderer{
		styles:    styleConfig(theme),
		renderers: map[int]*glamour.TermRenderer{},
	}
}

type glamourRenderer struct {
	styles    ansi.StyleConfig
	renderers map[int]*glamour.TermRenderer // keyed by wrap width
}

func (g *glamourRenderer) Render(source string, width int) (string, error) {
	if width <= 0 {
		return "", fmt.Errorf("markdown: width must be positive, got %d", width)
	}
	tr, ok := g.renderers[width]
	if !ok {
		var err error
		tr, err = glamour.NewTermRenderer(
			glamour.WithStyles(g.styles),
			glamour.WithWordWrap(width),
		)
		if err != nil {
			return "", fmt.Errorf("markdown: creating renderer: %w", err)
		}
		g.renderers[width] = tr
	}
	return tr.Render(source)
}

func hex(c color.Color) *string {
	r, g, b, _ := c.RGBA()
	s := fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
	return &s
}

func ptr[T any](v T) *T { return &v }

// styleConfig maps theme roles onto glamour's stylesheet. Restrained on
// purpose: color and weight from roles, structure left to glamour.
func styleConfig(t tuiweave.Theme) ansi.StyleConfig {
	return ansi.StyleConfig{
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{Color: hex(t.Text)},
			Margin:         ptr(uint(0)),
		},
		Paragraph: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{Color: hex(t.Text)},
		},
		Text: ansi.StylePrimitive{Color: hex(t.Text)},

		Heading: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{Color: hex(t.Accent), Bold: ptr(true)},
		},
		H1: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{Prefix: "# "},
		},
		H2: ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Prefix: "## "}},
		H3: ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Prefix: "### "}},

		Emph:           ansi.StylePrimitive{Italic: ptr(true)},
		Strong:         ansi.StylePrimitive{Bold: ptr(true)},
		Strikethrough:  ansi.StylePrimitive{CrossedOut: ptr(true)},
		HorizontalRule: ansi.StylePrimitive{Color: hex(t.BorderMuted)},

		BlockQuote: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{Color: hex(t.TextMuted), Italic: ptr(true)},
			Indent:         ptr(uint(1)),
			IndentToken:    ptr("│ "),
		},

		List: ansi.StyleList{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{Color: hex(t.Text)},
			},
			LevelIndent: 2,
		},
		Item:        ansi.StylePrimitive{Color: hex(t.Text), BlockPrefix: "• "},
		Enumeration: ansi.StylePrimitive{Color: hex(t.TextMuted), BlockPrefix: ". "},

		Link:     ansi.StylePrimitive{Color: hex(t.Info), Underline: ptr(true)},
		LinkText: ansi.StylePrimitive{Color: hex(t.Info)},

		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{Color: hex(t.Accent)},
		},
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{Color: hex(t.Text)},
				Margin:         ptr(uint(1)),
			},
			Chroma: &ansi.Chroma{
				Text:            ansi.StylePrimitive{Color: hex(t.Text)},
				Comment:         ansi.StylePrimitive{Color: hex(t.TextFaint), Italic: ptr(true)},
				Keyword:         ansi.StylePrimitive{Color: hex(t.Accent)},
				KeywordType:     ansi.StylePrimitive{Color: hex(t.Info)},
				NameFunction:    ansi.StylePrimitive{Color: hex(t.Info)},
				NameBuiltin:     ansi.StylePrimitive{Color: hex(t.Info)},
				LiteralString:   ansi.StylePrimitive{Color: hex(t.Success)},
				LiteralNumber:   ansi.StylePrimitive{Color: hex(t.Warning)},
				Operator:        ansi.StylePrimitive{Color: hex(t.TextMuted)},
				Punctuation:     ansi.StylePrimitive{Color: hex(t.TextMuted)},
				Error:           ansi.StylePrimitive{Color: hex(t.Danger)},
				GenericDeleted:  ansi.StylePrimitive{Color: hex(t.Danger)},
				GenericInserted: ansi.StylePrimitive{Color: hex(t.Success)},
			},
		},
	}
}
