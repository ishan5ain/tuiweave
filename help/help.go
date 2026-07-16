// Package help provides a one-line key-hint bar, e.g.:
//
//	tab focus • enter send • q quit
package help

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/layout"
)

// Binding is one key hint.
type Binding struct {
	Key  string
	Desc string
}

// Model is a help bar component. Create one with New.
type Model struct {
	width, height int
	bindings      []Binding

	keyStyle  lipgloss.Style
	descStyle lipgloss.Style
	sepStyle  lipgloss.Style
}

// New returns a help bar styled from the theme's roles.
func New(theme tuiweave.Theme) Model {
	return Model{
		keyStyle:  lipgloss.NewStyle().Foreground(theme.TextMuted).Bold(true),
		descStyle: lipgloss.NewStyle().Foreground(theme.TextFaint),
		sepStyle:  lipgloss.NewStyle().Foreground(theme.TextFaint),
	}
}

// SetSize sets the box the bar renders in; the bar is one line tall.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
}

// SizeMode reports that the help bar is width-constrained and renders at
// most one natural row.
func (m Model) SizeMode() layout.SizeMode { return layout.SizeWidthBounded }

// SetBindings replaces the displayed hints.
func (m *Model) SetBindings(bindings ...Binding) {
	m.bindings = bindings
}

// Update implements the component contract. The help bar handles no messages.
func (m Model) Update(_ tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// View renders as many hints as fit, whole hints only.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 || len(m.bindings) == 0 {
		return ""
	}
	sep := m.sepStyle.Render(" • ")
	var parts []string
	used := 0
	for i, b := range m.bindings {
		hint := m.keyStyle.Render(b.Key) + m.descStyle.Render(" "+b.Desc)
		w := ansi.StringWidth(hint)
		if i > 0 {
			w += 3 // separator
		}
		if used+w > m.width {
			break
		}
		used += w
		parts = append(parts, hint)
	}
	return strings.Join(parts, sep)
}
