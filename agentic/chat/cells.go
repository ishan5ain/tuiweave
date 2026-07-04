package chat

import (
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/agentic/markdown"
)

// User is a user-message cell: an accent "❯ you" header over plain text.
type User struct {
	text        string
	headerStyle lipgloss.Style
	bodyStyle   lipgloss.Style
}

// NewUser returns a user message cell.
func NewUser(theme gotui.Theme, text string) *User {
	return &User{
		text:        text,
		headerStyle: lipgloss.NewStyle().Foreground(theme.Accent).Bold(true),
		bodyStyle:   lipgloss.NewStyle().Foreground(theme.Text).PaddingLeft(2),
	}
}

// Render implements Cell.
func (u *User) Render(width int) string {
	return u.headerStyle.Render("❯ you") + "\n" +
		u.bodyStyle.Width(width).Render(u.text)
}

// Assistant is a streaming assistant-message cell: markdown-rendered text
// under a "✦ assistant" header. Stream into it with Append; the render is
// cached and recomputed only when the source or width changes.
type Assistant struct {
	renderer    markdown.Renderer
	source      string
	headerStyle lipgloss.Style

	cachedWidth int
	cachedLen   int
	cached      string
}

// NewAssistant returns an empty assistant message cell rendering through r.
func NewAssistant(theme gotui.Theme, r markdown.Renderer) *Assistant {
	return &Assistant{
		renderer:    r,
		headerStyle: lipgloss.NewStyle().Foreground(theme.TextMuted).Bold(true),
	}
}

// Append adds a streamed delta to the message source.
func (a *Assistant) Append(delta string) { a.source += delta }

// SetSource replaces the message source.
func (a *Assistant) SetSource(s string) { a.source = s }

// Source returns the raw markdown accumulated so far.
func (a *Assistant) Source() string { return a.source }

// Render implements Cell.
func (a *Assistant) Render(width int) string {
	if width != a.cachedWidth || len(a.source) != a.cachedLen {
		a.cached = markdown.Sprint(a.renderer, a.source, width)
		a.cachedWidth, a.cachedLen = width, len(a.source)
	}
	return a.headerStyle.Render("✦ assistant") + "\n" + a.cached
}

// Text is a plain one-off cell for session notes ("compacted history",
// errors), rendered faint.
type Text struct {
	text  string
	style lipgloss.Style
}

// NewText returns a faint note cell.
func NewText(theme gotui.Theme, text string) *Text {
	return &Text{
		text:  text,
		style: lipgloss.NewStyle().Foreground(theme.TextFaint).Italic(true),
	}
}

// Render implements Cell.
func (t *Text) Render(width int) string {
	return t.style.Width(width).Render(t.text)
}
