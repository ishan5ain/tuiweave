// Package permission provides the modal prompt agentic tools show before a
// sensitive action: a question with vertically stacked options.
//
// Like dialog, the app owns visibility: render it (composited with
// gotui/overlay) while waiting, and close it when Update returns a
// ResultMsg command.
package permission

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
)

// ResultMsg is emitted (as a command) when the user picks an option.
type ResultMsg struct {
	// ID identifies which prompt answered.
	ID string
	// Choice is the picked option's index; Option is its label.
	Choice int
	Option string
}

// Model is a permission prompt component. Create one with New.
type Model struct {
	width, height int
	sel           int
	options       []string

	// ID tags the ResultMsg this prompt emits.
	ID string
	// Title is the question, e.g. `Run "go test ./..."?`.
	Title string
	// Body adds detail below the title.
	Body string

	panelStyle    lipgloss.Style
	titleStyle    lipgloss.Style
	bodyStyle     lipgloss.Style
	optionStyle   lipgloss.Style
	selectedStyle lipgloss.Style
	numStyle      lipgloss.Style
}

// New returns a prompt with the standard options: Allow once, Allow always,
// Deny. The Warning role frames it — permission requests are caution
// moments, not errors.
func New(theme gotui.Theme) Model {
	return Model{
		options: []string{"Allow once", "Allow always", "Deny"},
		panelStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Warning).
			Background(theme.SurfaceRaised).
			Padding(1, 2),
		titleStyle: lipgloss.NewStyle().
			Foreground(theme.Text).
			Background(theme.SurfaceRaised).
			Bold(true),
		bodyStyle: lipgloss.NewStyle().
			Foreground(theme.TextMuted).
			Background(theme.SurfaceRaised),
		optionStyle: lipgloss.NewStyle().
			Foreground(theme.Text).
			Background(theme.SurfaceRaised),
		selectedStyle: lipgloss.NewStyle().
			Foreground(theme.TextInverted).
			Background(theme.Accent).
			Bold(true),
		numStyle: lipgloss.NewStyle().
			Foreground(theme.TextFaint).
			Background(theme.SurfaceRaised),
	}
}

// SetOptions replaces the options. The last option is treated as the safe
// default: esc picks it.
func (m *Model) SetOptions(options ...string) {
	if len(options) > 0 {
		m.options = options
		m.sel = 0
	}
}

// SetSize sets the prompt's outer box, border included.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
}

// Update handles navigation and answering: up/down/j/k move, number keys
// answer directly, enter answers the selection, esc picks the last option
// (the safe default).
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	s := key.String()
	switch s {
	case "up", "k":
		m.sel = max(0, m.sel-1)
	case "down", "j":
		m.sel = min(len(m.options)-1, m.sel+1)
	case "enter":
		return m, m.result(m.sel)
	case "esc":
		return m, m.result(len(m.options) - 1)
	default:
		if len(s) == 1 && s[0] >= '1' && s[0] <= '9' {
			if i := int(s[0] - '1'); i < len(m.options) {
				return m, m.result(i)
			}
		}
	}
	return m, nil
}

func (m Model) result(choice int) tea.Cmd {
	id, option := m.ID, m.options[choice]
	return func() tea.Msg { return ResultMsg{ID: id, Choice: choice, Option: option} }
}

// View renders the bordered prompt with numbered options.
func (m Model) View() string {
	if m.width <= 4 || m.height <= 4 {
		return ""
	}
	inner := m.width - 2 - 4 // border + horizontal padding

	rows := []string{
		m.titleStyle.Width(inner).Render(m.Title),
	}
	if m.Body != "" {
		rows = append(rows, m.bodyStyle.Width(inner).Render(m.Body))
	}
	rows = append(rows, m.bodyStyle.Width(inner).Render(""))
	for i, opt := range m.options {
		label := fmt.Sprintf(" %s ", opt)
		line := m.numStyle.Render(fmt.Sprintf(" %d ", i+1))
		if i == m.sel {
			line += m.selectedStyle.Render(label)
		} else {
			line += m.optionStyle.Render(label)
		}
		if pad := inner - lipgloss.Width(line); pad > 0 {
			line += m.bodyStyle.Render(strings.Repeat(" ", pad))
		}
		rows = append(rows, line)
	}
	return m.panelStyle.Width(m.width - 2).Render(strings.Join(rows, "\n"))
}
