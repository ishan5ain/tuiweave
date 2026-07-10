// Package button provides a compact, focusable action button. It owns only
// presentation and activation; applications own the action's side effects.
package button

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/inspect"
)

const (
	// ActionFocus focuses the button.
	ActionFocus = "focus"
	// ActionBlur blurs the button.
	ActionBlur = "blur"
	// ActionActivate presses the button.
	ActionActivate = "activate"
)

// PressedMsg is emitted when the button is activated by keyboard or semantic
// action input. ID lets an application route one message type from many
// buttons.
type PressedMsg struct {
	ID    string
	Label string
}

// Model is a one-line focusable button. Create one with New.
type Model struct {
	width, height int
	label         string
	disabled      bool
	focused       bool

	// ID identifies this button in PressedMsg and application inspection trees.
	ID string

	style         lipgloss.Style
	focusedStyle  lipgloss.Style
	disabledStyle lipgloss.Style
}

// New returns a button styled from the theme's semantic roles.
func New(theme gotui.Theme) Model {
	return Model{
		style: lipgloss.NewStyle().
			Foreground(theme.Text).
			Background(theme.SurfaceRaised),
		focusedStyle: lipgloss.NewStyle().
			Foreground(theme.TextInverted).
			Background(theme.Accent).
			Bold(true),
		disabledStyle: lipgloss.NewStyle().
			Foreground(theme.TextFaint).
			Background(theme.SurfaceSunken),
	}
}

// SetSize sets the box the button renders in. The button renders one row
// whenever both dimensions are positive.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
}

// SetLabel sets the button's visible label.
func (m *Model) SetLabel(label string) { m.label = label }

// Label returns the button's visible label.
func (m Model) Label() string { return m.label }

// SetDisabled controls whether the button can be focused or activated.
func (m *Model) SetDisabled(disabled bool) {
	m.disabled = disabled
	if disabled {
		m.focused = false
	}
}

// Disabled reports whether the button is unavailable.
func (m Model) Disabled() bool { return m.disabled }

// Focus makes the button respond to enter and space.
func (m *Model) Focus() {
	if !m.disabled {
		m.focused = true
	}
}

// Blur stops the button from responding to keys.
func (m *Model) Blur() { m.focused = false }

// Focused reports whether the button accepts keyboard input.
func (m Model) Focused() bool { return m.focused }

// Activate returns a command that emits PressedMsg, or nil when the button is
// disabled.
func (m Model) Activate() tea.Cmd {
	if m.disabled {
		return nil
	}
	id, label := m.ID, m.label
	return func() tea.Msg { return PressedMsg{ID: id, Label: label} }
}

// Update handles semantic actions and, while focused, enter and space.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if next, cmd, handled := m.applyAction(msg); handled {
		return next, cmd
	}
	if !m.focused || m.disabled {
		return m, nil
	}
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "enter", "space":
		return m, m.Activate()
	}
	return m, nil
}

// View renders one exact-width button row. Labels are truncated inside the
// brackets at narrow widths.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	style := m.style
	if m.disabled {
		style = m.disabledStyle
	} else if m.focused {
		style = m.focusedStyle
	}
	if m.width < 4 {
		return style.Render(ansi.Truncate("[ "+m.label+" ]", m.width, "…"))
	}

	labelWidth := m.width - 4
	label := ansi.Truncate(m.label, labelWidth, "…")
	return style.Render("[ " + label + strings.Repeat(" ", labelWidth-ansi.StringWidth(label)) + " ]")
}

func (m Model) applyAction(msg tea.Msg) (Model, tea.Cmd, bool) {
	action, ok := msg.(inspect.ActionMsg)
	if !ok {
		return m, nil, false
	}
	switch action.ID {
	case ActionFocus:
		m.Focus()
	case ActionBlur:
		m.Blur()
	case ActionActivate:
		return m, m.Activate(), true
	default:
		return m, nil, false
	}
	return m, nil, true
}
