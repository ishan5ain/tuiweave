// Package dialog provides a modal confirmation box with OK/Cancel buttons.
//
// The dialog renders its own bordered panel; composite it over the app with
// overlay.Center. Visibility belongs to the app: show the dialog by
// rendering it, and close it when Update returns a ResultMsg command.
//
//	case dlgMsg := <-... // in the app's Update:
//	case dialog.ResultMsg:
//	    m.showDialog = false
//	    if dlgMsg.OK { ... }
package dialog

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/layout"
)

// ResultMsg is emitted (as a command) when the user confirms or dismisses
// the dialog.
type ResultMsg struct {
	// ID identifies which dialog answered, for apps with more than one.
	ID string
	// OK is true when the user chose the confirm button.
	OK bool
}

// Model is a dialog component. Create one with New.
type Model struct {
	width, height int
	sel           int // 0 = confirm, 1 = cancel

	// ID tags the ResultMsg this dialog emits.
	ID string
	// Title is rendered bold at the top of the panel.
	Title string
	// Body is rendered below the title, wrapped to the panel width.
	Body string
	// ConfirmLabel and CancelLabel name the buttons. Defaults: "OK", "Cancel".
	ConfirmLabel, CancelLabel string

	panelStyle    lipgloss.Style
	titleStyle    lipgloss.Style
	bodyStyle     lipgloss.Style
	buttonStyle   lipgloss.Style
	selectedStyle lipgloss.Style
}

const (
	ActionConfirm  = "confirm"
	ActionCancel   = "cancel"
	ActionNext     = "next"
	ActionPrevious = "previous"
)

// New returns a dialog styled from the theme's roles.
func New(theme tuiweave.Theme) Model {
	return Model{
		ConfirmLabel: "OK",
		CancelLabel:  "Cancel",
		panelStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.BorderFocused).
			Background(theme.SurfaceRaised).
			Padding(1, 2),
		titleStyle: lipgloss.NewStyle().
			Foreground(theme.Text).
			Background(theme.SurfaceRaised).
			Bold(true),
		bodyStyle: lipgloss.NewStyle().
			Foreground(theme.TextMuted).
			Background(theme.SurfaceRaised),
		buttonStyle: lipgloss.NewStyle().
			Foreground(theme.TextMuted).
			Background(theme.SurfaceSunken).
			Padding(0, 2),
		selectedStyle: lipgloss.NewStyle().
			Foreground(theme.TextInverted).
			Background(theme.Accent).
			Bold(true).
			Padding(0, 2),
	}
}

// SetTheme rebuilds all styles from the theme's roles, preserving
// non-style state such as the title, body, and button labels.
func (m *Model) SetTheme(theme tuiweave.Theme) {
	m.panelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.BorderFocused).
		Background(theme.SurfaceRaised).
		Padding(1, 2)
	m.titleStyle = lipgloss.NewStyle().
		Foreground(theme.Text).
		Background(theme.SurfaceRaised).
		Bold(true)
	m.bodyStyle = lipgloss.NewStyle().
		Foreground(theme.TextMuted).
		Background(theme.SurfaceRaised)
	m.buttonStyle = lipgloss.NewStyle().
		Foreground(theme.TextMuted).
		Background(theme.SurfaceSunken).
		Padding(0, 2)
	m.selectedStyle = lipgloss.NewStyle().
		Foreground(theme.TextInverted).
		Background(theme.Accent).
		Bold(true).
		Padding(0, 2)
}

// SetSize sets the dialog's outer box, border included.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
}

// SizeMode reports that dialog width is bounded while content height is
// determined by the title, body, and buttons.
func (m Model) SizeMode() layout.SizeMode { return layout.SizeWidthBounded }

// Update handles button navigation and confirmation:
// left/right/tab switch buttons, enter answers, esc cancels.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if next, cmd, handled := m.applyAction(msg); handled {
		return next, cmd
	}
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "left", "right", "tab", "shift+tab":
		m.sel = 1 - m.sel
	case "enter":
		return m, m.result(m.sel == 0)
	case "esc":
		return m, m.result(false)
	}
	return m, nil
}

func (m Model) result(ok bool) tea.Cmd {
	id := m.ID
	return func() tea.Msg { return ResultMsg{ID: id, OK: ok} }
}

// View renders the bordered panel with title, body, and buttons.
func (m Model) View() string {
	inner := m.width - 2 - 4 // border + horizontal padding
	if inner <= 0 || m.height <= 4 {
		return ""
	}

	confirm := m.buttonStyle
	cancel := m.selectedStyle
	if m.sel == 0 {
		confirm, cancel = m.selectedStyle, m.buttonStyle
	}
	buttons := confirm.Render(m.ConfirmLabel) + m.bodyStyle.Render("  ") + cancel.Render(m.CancelLabel)
	if lipgloss.Width(buttons) > inner {
		buttons = ansi.Truncate(buttons, inner, "")
	}

	sections := []string{
		bounded(m.titleStyle, m.Title, inner),
		bounded(m.bodyStyle, m.Body, inner),
		"",
		lipgloss.PlaceHorizontal(inner, lipgloss.Right, buttons,
			lipgloss.WithWhitespaceStyle(m.bodyStyle)),
	}
	content := strings.Join(sections, "\n")
	return m.panelStyle.Width(m.width).Render(content)
}

func bounded(style lipgloss.Style, text string, width int) string {
	if width <= 0 {
		return ""
	}
	rows := strings.Split(style.Width(width).Render(text), "\n")
	for i, row := range rows {
		if lipgloss.Width(row) > width {
			row = ansi.Truncate(row, width, "")
		}
		rows[i] = row
	}
	return strings.Join(rows, "\n")
}
