// Package chat provides a scrolling transcript for agentic sessions: a
// stack of cells (user/assistant messages, tool calls, diffs, notes) inside
// a viewport that auto-follows new content.
//
// Cells are pointers the app keeps and mutates as the session progresses —
// streaming text into an assistant message, flipping a tool call's status.
// The transcript re-renders from cells on Append, SetSize, and Invalidate:
// after mutating a cell in place, call Invalidate() to re-render (Append and
// SetSize do it for you).
//
// Auto-follow: the transcript sticks to the bottom while the user is at the
// bottom; scrolling up unsticks it, scrolling back to the bottom re-sticks.
package chat

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/viewport"
)

// Cell is one unit of transcript content, rendered at the transcript width.
// Cell implementations live in this package (Text, User, Assistant) and in
// agentic/toolcall; adapt anything else with CellFunc.
type Cell interface {
	Render(width int) string
}

// CellFunc adapts a plain function to a Cell — e.g. an inline diff:
//
//	chat.Append(chat.CellFunc(func(w int) string {
//	    return diffview.Sprint(theme, diff, w)
//	}))
type CellFunc func(width int) string

// Render implements Cell.
func (f CellFunc) Render(width int) string { return f(width) }

// Model is a chat transcript component. Create one with New.
type Model struct {
	vp            viewport.Model
	cells         []Cell
	follow        bool
	width, height int
}

// New returns an empty transcript that auto-follows new content.
func New(theme gotui.Theme) Model {
	return Model{vp: viewport.New(theme), follow: true}
}

// SetSize sets the box the transcript renders in.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
	m.vp.SetSize(width, height)
	m.Invalidate()
}

// Append adds cells to the end of the transcript.
func (m *Model) Append(cells ...Cell) {
	m.cells = append(m.cells, cells...)
	m.Invalidate()
}

// Cells returns the transcript's cells, oldest first.
func (m Model) Cells() []Cell { return m.cells }

// Invalidate re-renders the transcript from its cells. Call it after
// mutating a cell in place (streaming a delta, changing a tool status).
func (m *Model) Invalidate() {
	if m.width <= 0 {
		return
	}
	blocks := make([]string, len(m.cells))
	for i, c := range m.cells {
		blocks[i] = c.Render(m.width)
	}
	m.vp.SetContent(strings.Join(blocks, "\n\n"))
	if m.follow {
		m.vp.GotoBottom()
	}
}

// Focus makes the transcript respond to scroll keys.
func (m *Model) Focus() { m.vp.Focus() }

// Blur stops the transcript from responding to keys.
func (m *Model) Blur() { m.vp.Blur() }

// Focused reports whether the transcript handles keys.
func (m Model) Focused() bool { return m.vp.Focused() }

// Following reports whether the transcript is stuck to the bottom.
func (m Model) Following() bool { return m.follow }

// TotalLines returns the rendered transcript length (scrollbar.Scrollable).
func (m Model) TotalLines() int { return m.vp.TotalLines() }

// VisibleLines returns how many lines are shown at once (scrollbar.Scrollable).
func (m Model) VisibleLines() int { return m.vp.VisibleLines() }

// YOffset returns the index of the first visible line (scrollbar.Scrollable).
func (m Model) YOffset() int { return m.vp.YOffset() }

// GotoBottom scrolls to the newest content and re-enables auto-follow.
func (m *Model) GotoBottom() {
	m.vp.GotoBottom()
	m.follow = true
}

// Update delegates scrolling to the viewport and maintains auto-follow:
// any scroll that leaves the bottom unsticks; returning to the bottom
// re-sticks.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if next, handled := m.applyAction(msg); handled {
		return next, nil
	}
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	m.follow = m.vp.AtBottom()
	return m, cmd
}

// View renders the visible window of the transcript.
func (m Model) View() string { return m.vp.View() }
