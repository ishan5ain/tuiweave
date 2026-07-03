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

	"github.com/ishansain/gotui"
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

// New returns a dialog styled from the theme's roles.
func New(theme gotui.Theme) Model {
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

// SetSize sets the dialog's outer box, border included.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
}

// Update handles button navigation and confirmation:
// left/right/tab switch buttons, enter answers, esc cancels.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
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
	if m.width <= 4 || m.height <= 4 {
		return ""
	}
	inner := m.width - 2 - 4 // border + horizontal padding

	confirm := m.buttonStyle
	cancel := m.selectedStyle
	if m.sel == 0 {
		confirm, cancel = m.selectedStyle, m.buttonStyle
	}
	buttons := confirm.Render(m.ConfirmLabel) + m.bodyStyle.Render("  ") + cancel.Render(m.CancelLabel)

	sections := []string{
		m.titleStyle.Width(inner).Render(m.Title),
		m.bodyStyle.Width(inner).Render(m.Body),
		"",
		lipgloss.PlaceHorizontal(inner, lipgloss.Right, buttons,
			lipgloss.WithWhitespaceStyle(m.bodyStyle)),
	}
	content := strings.Join(sections, "\n")
	return m.panelStyle.Width(m.width - 2).Render(content)
}
