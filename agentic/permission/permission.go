// Package permission provides the modal prompt agentic tools show before a
// sensitive action: a question with vertically stacked options and optional
// structured provenance for the operation being approved.
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
	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/layout"
)

// ResultMsg is emitted (as a command) when the user picks an option.
type ResultMsg struct {
	// ID identifies which prompt answered.
	ID string
	// Choice is the picked option's index; Option is its label.
	Choice int
	Option string
}

// Provenance describes the operation behind a permission request. All fields
// are optional so existing title/body prompts remain valid.
type Provenance struct {
	// Tool identifies the tool or integration requesting approval.
	Tool string
	// Operation is the intended operation, such as "execute" or "write".
	Operation string
	// Target identifies the primary resource, repository, or service.
	Target string
	// Scope describes the affected paths, resources, or permission boundary.
	Scope string
	// Detail contains the exact command, patch, request, or other operation data.
	Detail string
	// Impact explains the expected user-visible or system effect.
	Impact string
	// Reversibility should be "reversible", "irreversible", or empty when
	// unknown.
	Reversibility string
	// Policy explains why approval is required or which policy applies.
	Policy string
}

// Model is a permission prompt component. Create one with New.
type Model struct {
	width, height int
	sel           int
	options       []string
	provenance    Provenance

	// ID tags the ResultMsg this prompt emits.
	ID string
	// Title is the question, e.g. `Run "go test ./..."?`.
	Title string
	// Body adds detail below the title.
	Body string

	panelStyle      lipgloss.Style
	titleStyle      lipgloss.Style
	bodyStyle       lipgloss.Style
	optionStyle     lipgloss.Style
	selectedStyle   lipgloss.Style
	numStyle        lipgloss.Style
	provenanceStyle lipgloss.Style
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
		provenanceStyle: lipgloss.NewStyle().
			Foreground(theme.TextMuted).
			Background(theme.SurfaceRaised),
	}
}

// SetProvenance replaces the structured operation details shown in the prompt.
func (m *Model) SetProvenance(p Provenance) { m.provenance = p }

// Provenance returns the structured operation details for this prompt.
func (m Model) Provenance() Provenance { return m.provenance }

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

// SizeMode reports that prompt width is bounded while content height is
// determined by the title, body, provenance, and options.
func (m Model) SizeMode() layout.SizeMode { return layout.SizeWidthBounded }

// Update handles navigation and answering: up/down/j/k move, number keys
// answer directly, enter answers the selection, esc picks the last option
// (the safe default).
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if next, cmd, handled := m.applyAction(msg); handled {
		return next, cmd
	}
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
	inner := m.width - 2 - 4 // border + horizontal padding
	if inner <= 0 || m.height <= 4 {
		return ""
	}

	rows := []string{
		bounded(m.titleStyle, m.Title, inner),
	}
	if m.Body != "" {
		rows = append(rows, bounded(m.bodyStyle, m.Body, inner))
	}
	rows = append(rows, m.provenanceLines(inner)...)
	rows = append(rows, bounded(m.bodyStyle, "", inner))
	for i, opt := range m.options {
		label := fmt.Sprintf(" %s ", opt)
		line := m.numStyle.Render(fmt.Sprintf(" %d ", i+1))
		if i == m.sel {
			line += m.selectedStyle.Render(label)
		} else {
			line += m.optionStyle.Render(label)
		}
		if lipgloss.Width(line) > inner {
			line = ansi.Truncate(line, inner, "")
		}
		if pad := inner - lipgloss.Width(line); pad > 0 {
			line += m.bodyStyle.Render(strings.Repeat(" ", pad))
		}
		rows = append(rows, line)
	}
	return m.panelStyle.Width(m.width).Render(strings.Join(rows, "\n"))
}

func (m Model) provenanceLines(width int) []string {
	p := m.provenance
	fields := []struct {
		label string
		value string
	}{
		{"Tool", p.Tool},
		{"Operation", p.Operation},
		{"Target", p.Target},
		{"Scope", p.Scope},
		{"Detail", p.Detail},
		{"Impact", p.Impact},
		{"Reversibility", p.Reversibility},
		{"Policy", p.Policy},
	}
	rows := make([]string, 0, len(fields))
	for _, field := range fields {
		if field.value == "" {
			continue
		}
		rows = append(rows, bounded(m.provenanceStyle, field.label+": "+field.value, width))
	}
	return rows
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
