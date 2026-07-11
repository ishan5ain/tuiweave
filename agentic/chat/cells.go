package chat

import (
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/agentic/markdown"
)

func fitHeader(style lipgloss.Style, text string, width int) string {
	if width <= 0 {
		return ""
	}
	return style.Render(ansi.Truncate(text, width, ""))
}

func fitCell(view string, width int) string {
	if width <= 0 {
		return ""
	}
	rows := strings.Split(view, "\n")
	for i, row := range rows {
		rows[i] = ansi.Truncate(row, width, "")
	}
	return strings.Join(rows, "\n")
}

func fitPaddedBody(style lipgloss.Style, text string, width int) string {
	if width <= 0 {
		return ""
	}
	padding := min(2, width)
	contentWidth := width - padding
	if contentWidth == 0 {
		return strings.Repeat(" ", padding)
	}
	rendered := style.PaddingLeft(padding).Width(contentWidth).Render(text)
	rows := strings.Split(rendered, "\n")
	for i, row := range rows {
		row = ansi.Truncate(row, width, "")
		if pad := width - ansi.StringWidth(row); pad > 0 {
			row += strings.Repeat(" ", pad)
		}
		rows[i] = row
	}
	return strings.Join(rows, "\n")
}

// User is a user-message cell: an accent "❯ you" header over plain text.
type User struct {
	id          string
	text        string
	headerStyle lipgloss.Style
	bodyStyle   lipgloss.Style
}

// NewUser returns a user message cell.
func NewUser(theme tuiweave.Theme, text string) *User {
	return &User{
		text:        text,
		headerStyle: lipgloss.NewStyle().Foreground(theme.Accent).Bold(true),
		bodyStyle:   lipgloss.NewStyle().Foreground(theme.Text).PaddingLeft(2),
	}
}

// SetID assigns the application-owned stable identity for this cell.
func (u *User) SetID(id string) { u.id = id }

// CellID implements CellIdentity.
func (u *User) CellID() string { return u.id }

// CellKind implements CellKind.
func (u *User) CellKind() string { return "user" }

// Lifecycle implements CellLifecycle. User messages are complete when added.
func (u *User) Lifecycle() State { return StateComplete }

// Render implements Cell.
func (u *User) Render(width int) string {
	return fitCell(fitHeader(u.headerStyle, "❯ you", width)+"\n"+
		fitPaddedBody(u.bodyStyle, u.text, width), width)
}

// Assistant is a streaming assistant-message cell: markdown-rendered text
// under a "✦ assistant" header. Stream into it with Append; the render is
// cached and recomputed only when the source revision or width changes.
type Assistant struct {
	id          string
	renderer    markdown.Renderer
	source      string
	state       State
	headerStyle lipgloss.Style

	cachedWidth    int
	sourceRevision uint64
	cachedRevision uint64
	cached         string
	cachedValid    bool
}

// NewAssistant returns an empty assistant message cell rendering through r.
func NewAssistant(theme tuiweave.Theme, r markdown.Renderer) *Assistant {
	return &Assistant{
		renderer:    r,
		state:       StateStreaming,
		headerStyle: lipgloss.NewStyle().Foreground(theme.TextMuted).Bold(true),
	}
}

// SetID assigns the application-owned stable identity for this cell.
func (a *Assistant) SetID(id string) { a.id = id }

// CellID implements CellIdentity.
func (a *Assistant) CellID() string { return a.id }

// CellKind implements CellKind.
func (a *Assistant) CellKind() string { return "assistant" }

// SetLifecycle changes the assistant message lifecycle state.
func (a *Assistant) SetLifecycle(state State) { a.state = state }

// Lifecycle implements CellLifecycle.
func (a *Assistant) Lifecycle() State { return a.state }

// Append adds a streamed delta to the message source.
func (a *Assistant) Append(delta string) {
	if delta == "" {
		return
	}
	a.source += delta
	a.sourceRevision++
}

// SetSource replaces the message source.
func (a *Assistant) SetSource(s string) {
	if a.source == s {
		return
	}
	a.source = s
	a.sourceRevision++
}

// Source returns the raw markdown accumulated so far.
func (a *Assistant) Source() string { return a.source }

// Render implements Cell.
func (a *Assistant) Render(width int) string {
	if !a.cachedValid || width != a.cachedWidth || a.sourceRevision != a.cachedRevision {
		a.cached = markdown.Sprint(a.renderer, a.source, width)
		a.cachedWidth, a.cachedRevision, a.cachedValid = width, a.sourceRevision, true
	}
	return fitCell(fitHeader(a.headerStyle, "✦ assistant", width)+"\n"+a.cached, width)
}

// Text is a plain one-off cell for session notes ("compacted history",
// errors), rendered faint.
type Text struct {
	id    string
	text  string
	style lipgloss.Style
}

// NewText returns a faint note cell.
func NewText(theme tuiweave.Theme, text string) *Text {
	return &Text{
		text:  text,
		style: lipgloss.NewStyle().Foreground(theme.TextFaint).Italic(true),
	}
}

// SetID assigns the application-owned stable identity for this cell.
func (t *Text) SetID(id string) { t.id = id }

// CellID implements CellIdentity.
func (t *Text) CellID() string { return t.id }

// CellKind implements CellKind.
func (t *Text) CellKind() string { return "text" }

// Lifecycle implements CellLifecycle. Notes are complete when added.
func (t *Text) Lifecycle() State { return StateComplete }

// Render implements Cell.
func (t *Text) Render(width int) string {
	return fitCell(t.style.Width(width).Render(t.text), width)
}
