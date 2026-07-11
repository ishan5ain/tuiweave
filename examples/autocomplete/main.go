// Autocomplete demonstrates an app-owned textinput paired with the reusable
// suggestion window. The app owns query synchronization and insertion.
//
//	go run ./examples/autocomplete
//
// Type a command prefix, use up/down to navigate, and press enter to accept.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/autocomplete"
	"github.com/ishansain/gotui/frame"
	"github.com/ishansain/gotui/layout"
	"github.com/ishansain/gotui/statusbar"
	"github.com/ishansain/gotui/textinput"
)

type model struct {
	theme         gotui.Theme
	width, height int
	input         textinput.Model
	suggestions   autocomplete.Model
	status        statusbar.Model
}

func newModel() model {
	theme := gotui.Dark()
	input := textinput.New(theme)
	input.Prompt = "> "
	input.Placeholder = "type a command prefix"
	input.Focus()

	suggestions := autocomplete.New(theme)
	suggestions.SetItems(
		autocomplete.Item{ID: "git-checkout", Value: "git checkout ", Label: "git checkout", Description: "switch branch"},
		autocomplete.Item{ID: "git-status", Value: "git status", Label: "git status", Description: "show changes"},
		autocomplete.Item{ID: "go-test", Value: "go test ./...", Label: "go test", Description: "run tests"},
		autocomplete.Item{ID: "go-run", Value: "go run ./cmd", Label: "go run", Description: "run a command"},
		autocomplete.Item{ID: "deploy", Value: "deploy preview", Label: "deploy preview", Description: "preview a deployment", Disabled: true},
	)
	suggestions.Focus()

	m := model{
		theme:       theme,
		input:       input,
		suggestions: suggestions,
		status:      statusbar.New(theme),
	}
	m.syncQuery()
	m.syncStatus()
	return m
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()
	case autocomplete.SelectedMsg:
		m.input.SetValue(msg.Value)
		m.syncQuery()
		m.status.SetLeft(
			statusbar.Segment{Text: "accepted", Kind: statusbar.KindAccent},
			statusbar.Segment{Text: msg.Label, Kind: statusbar.KindNormal},
		)
		return m, nil
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

		before := m.input.Value()
		var inputCmd tea.Cmd
		m.input, inputCmd = m.input.Update(msg)
		cmds = append(cmds, inputCmd)
		if m.input.Value() != before {
			m.syncQuery()
		}

		var suggestionCmd tea.Cmd
		m.suggestions, suggestionCmd = m.suggestions.Update(msg)
		cmds = append(cmds, suggestionCmd)
	}
	m.syncStatus()
	return m, tea.Batch(cmds...)
}

func (m *model) layout() {
	var main, footer layout.Rect
	layout.Vertical(layout.Fill(1), layout.Len(1)).Split(
		layout.NewRect(0, 0, m.width, m.height),
	).Assign(&main, &footer)
	content := frame.PanelContentRect(main, frame.PanelOptions{Title: "Command completion", Focused: m.input.Focused(), Padding: 1})
	var input, suggestions layout.Rect
	layout.Vertical(layout.Len(1), layout.Fill(1)).Split(content).Assign(&input, &suggestions)
	m.input.SetSize(input.Dx(), input.Dy())
	m.suggestions.SetSize(suggestions.Dx(), suggestions.Dy())
	m.status.SetSize(footer.Dx(), footer.Dy())
}

func (m *model) syncQuery() { m.suggestions.SetQuery(m.input.Value()) }

func (m *model) syncStatus() {
	selected := m.suggestions.SelectedID()
	if selected == "" {
		selected = "none"
	}
	m.status.SetLeft(
		statusbar.Segment{Text: "completion", Kind: statusbar.KindAccent},
		statusbar.Segment{Text: selected, Kind: statusbar.KindNormal},
	)
	m.status.SetRight(
		statusbar.Segment{Text: fmt.Sprintf("%d matches", m.suggestions.FilteredLen()), Kind: statusbar.KindMuted},
		statusbar.Segment{Text: "↑↓ choose  enter accept", Kind: statusbar.KindInfo},
	)
}

func (m model) render() string {
	if m.width < 16 || m.height < 6 {
		return "resize terminal"
	}
	content := lipgloss.JoinVertical(lipgloss.Left, m.input.View(), m.suggestions.View())
	panel := frame.Panel(m.theme, content, m.width, frame.PanelOptions{
		Title:   "Command completion",
		Focused: m.input.Focused(),
		Padding: 1,
	})
	return lipgloss.JoinVertical(lipgloss.Left, panel, m.status.View())
}

func (m model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
