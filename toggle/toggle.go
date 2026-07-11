// Package toggle provides a compact, focusable on/off control. It owns only
// the visual state and emits a message when that state changes; applications
// own the setting's meaning and persistence.
package toggle

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/inspect"
)

const (
	// ActionFocus focuses the toggle.
	ActionFocus = "focus"
	// ActionBlur blurs the toggle.
	ActionBlur = "blur"
	// ActionToggle flips the current value.
	ActionToggle = "toggle"
	// ActionOn turns the value on.
	ActionOn = "on"
	// ActionOff turns the value off.
	ActionOff = "off"
)

// ChangedMsg is emitted when the toggle changes through keyboard or semantic
// action input. ID lets an application route one message type from many
// toggles.
type ChangedMsg struct {
	ID      string
	Checked bool
}

// Model is a one-line focusable toggle. Create one with New.
type Model struct {
	width, height int
	label         string
	checked       bool
	disabled      bool
	focused       bool

	// ID identifies this toggle in ChangedMsg and application inspection trees.
	ID string

	labelStyle         lipgloss.Style
	disabledLabelStyle lipgloss.Style
	checkedStyle       lipgloss.Style
	uncheckedStyle     lipgloss.Style
	focusedStyle       lipgloss.Style
	disabledBoxStyle   lipgloss.Style
}

// New returns an unchecked toggle styled from the theme's semantic roles.
func New(theme tuiweave.Theme) Model {
	base := lipgloss.NewStyle().Background(theme.SurfaceRaised)
	return Model{
		labelStyle:         base.Foreground(theme.Text),
		disabledLabelStyle: base.Foreground(theme.TextFaint),
		checkedStyle: lipgloss.NewStyle().
			Foreground(theme.TextInverted).
			Background(theme.Success).
			Bold(true),
		uncheckedStyle: lipgloss.NewStyle().
			Foreground(theme.TextMuted).
			Background(theme.SurfaceSunken).
			Bold(true),
		focusedStyle: lipgloss.NewStyle().
			Foreground(theme.SelectionFg).
			Background(theme.SelectionBg).
			Bold(true),
		disabledBoxStyle: lipgloss.NewStyle().
			Foreground(theme.TextFaint).
			Background(theme.SurfaceSunken),
	}
}

// SetSize sets the box the toggle renders in. The toggle renders one row
// whenever both dimensions are positive.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
}

// SetLabel sets the setting label shown beside the checkbox.
func (m *Model) SetLabel(label string) { m.label = label }

// Label returns the setting label.
func (m Model) Label() string { return m.label }

// SetChecked sets the on/off value without emitting a message.
func (m *Model) SetChecked(checked bool) { m.checked = checked }

// Checked reports whether the toggle is on.
func (m Model) Checked() bool { return m.checked }

// SetDisabled controls whether the toggle can be focused or changed.
func (m *Model) SetDisabled(disabled bool) {
	m.disabled = disabled
	if disabled {
		m.focused = false
	}
}

// Disabled reports whether the toggle is unavailable.
func (m Model) Disabled() bool { return m.disabled }

// Focus makes the toggle respond to space, x, and enter.
func (m *Model) Focus() {
	if !m.disabled {
		m.focused = true
	}
}

// Blur stops the toggle from responding to keys.
func (m *Model) Blur() { m.focused = false }

// Focused reports whether the toggle accepts keyboard input.
func (m Model) Focused() bool { return m.focused }

// Toggle flips the value without emitting a message. Use Update when the app
// needs a ChangedMsg command for the transition.
func (m *Model) Toggle() {
	if !m.disabled {
		m.checked = !m.checked
	}
}

// Update handles semantic actions and, while focused, space, x, and enter.
// State changes return a command that emits ChangedMsg.
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
	case "space", "x", "enter":
		return m, m.setChecked(!m.checked)
	}
	return m, nil
}

// View renders one exact-width row. At narrow widths the label is truncated
// first, then the checkbox is shortened only when the box is narrower than
// its normal three-cell marker.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	box := "[ ]"
	if m.checked {
		box = "[x]"
	}
	boxStyle, labelStyle := m.styles()
	if m.width < 4 {
		return boxStyle.Render(ansi.Truncate(box, m.width, "…"))
	}

	labelWidth := m.width - 4
	label := ansi.Truncate(m.label, labelWidth, "…")
	return boxStyle.Render(box) +
		labelStyle.Render(" "+label+strings.Repeat(" ", labelWidth-ansi.StringWidth(label)))
}

func (m Model) styles() (lipgloss.Style, lipgloss.Style) {
	if m.disabled {
		return m.disabledBoxStyle, m.disabledLabelStyle
	}
	if m.focused {
		return m.focusedStyle, m.focusedStyle
	}
	if m.checked {
		return m.checkedStyle, m.labelStyle
	}
	return m.uncheckedStyle, m.labelStyle
}

func (m *Model) setChecked(checked bool) tea.Cmd {
	if m.disabled || m.checked == checked {
		return nil
	}
	m.checked = checked
	id := m.ID
	return func() tea.Msg { return ChangedMsg{ID: id, Checked: checked} }
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
	case ActionToggle:
		return m, m.setChecked(!m.checked), true
	case ActionOn:
		return m, m.setChecked(true), true
	case ActionOff:
		return m, m.setChecked(false), true
	default:
		return m, nil, false
	}
	return m, nil, true
}
