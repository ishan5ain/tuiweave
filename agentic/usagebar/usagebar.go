// Package usagebar renders the session status line of an agentic tool:
// model name, token counts, cost, and context usage, formatted onto a
// gotui statusbar.
package usagebar

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/statusbar"
)

// Stats is a session usage snapshot.
type Stats struct {
	// Model is the LLM identifier, e.g. "pi-large".
	Model string
	// TokensIn and TokensOut are cumulative prompt/completion tokens.
	TokensIn, TokensOut int
	// Cost is the session cost in dollars.
	Cost float64
	// ContextUsed is the context-window fill fraction in [0, 1].
	ContextUsed float64
}

// Model is a usage bar component. Create one with New.
type Model struct {
	bar   statusbar.Model
	stats Stats
}

// New returns an empty usage bar styled from the theme's roles.
func New(theme gotui.Theme) Model {
	m := Model{bar: statusbar.New(theme)}
	m.SetStats(Stats{})
	return m
}

// SetStats updates the displayed numbers.
func (m *Model) SetStats(s Stats) {
	m.stats = s
	m.bar.SetLeft(statusbar.Segment{Text: s.Model, Kind: statusbar.KindAccent})

	ctxKind := statusbar.KindMuted
	switch {
	case s.ContextUsed >= 0.95:
		ctxKind = statusbar.KindDanger
	case s.ContextUsed >= 0.8:
		ctxKind = statusbar.KindWarning
	}
	m.bar.SetRight(
		statusbar.Segment{
			Text: fmt.Sprintf("↑%s ↓%s", humanize(s.TokensIn), humanize(s.TokensOut)),
			Kind: statusbar.KindMuted,
		},
		statusbar.Segment{Text: fmt.Sprintf("$%.2f", s.Cost), Kind: statusbar.KindMuted},
		statusbar.Segment{Text: fmt.Sprintf("ctx %.0f%%", s.ContextUsed*100), Kind: ctxKind},
	)
}

// Stats returns the last snapshot.
func (m Model) Stats() Stats { return m.stats }

// SetSize sets the box the bar renders in; the bar is one line tall.
func (m *Model) SetSize(width, height int) { m.bar.SetSize(width, height) }

// Update implements the component contract. The bar handles no messages.
func (m Model) Update(_ tea.Msg) (Model, tea.Cmd) { return m, nil }

// View renders the bar.
func (m Model) View() string { return m.bar.View() }

// humanize renders token counts compactly: 999, 12.3k, 1.2M.
func humanize(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}
