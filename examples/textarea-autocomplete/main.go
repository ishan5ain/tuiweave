// Textarea autocomplete demonstrates an app-owned completion engine paired
// with textarea.ReplaceRange. The textarea keeps logical cursor and editing
// state; the application owns token detection and replacement policy.
//
//	go run ./examples/textarea-autocomplete
//
// Type a command fragment such as "run git", use up/down to choose a
// completion, and press enter to replace the token before the cursor.
package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/autocomplete"
	"github.com/ishan5ain/tuiweave/frame"
	"github.com/ishan5ain/tuiweave/layout"
	"github.com/ishan5ain/tuiweave/statusbar"
	"github.com/ishan5ain/tuiweave/textarea"
)

type model struct {
	theme         tuiweave.Theme
	width, height int
	input         textarea.Model
	suggestions   autocomplete.Model
	status        statusbar.Model
}

func newModel() model {
	theme := tuiweave.Dark()
	input := textarea.New(theme)
	input.Prompt = "> "
	input.Placeholder = "type a command fragment"
	input.Focus()

	suggestions := autocomplete.New(theme)
	suggestions.SetItems(
		autocomplete.Item{ID: "git-status", Value: "git status", Label: "git status", Description: "show changes"},
		autocomplete.Item{ID: "git-log", Value: "git log --oneline", Label: "git log --oneline", Description: "show commits"},
		autocomplete.Item{ID: "go-test", Value: "go test ./...", Label: "go test ./...", Description: "run tests"},
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
		m.applyCompletion(msg)
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

		if m.completionActive() && isCompletionKey(msg) {
			var suggestionCmd tea.Cmd
			m.suggestions, suggestionCmd = m.suggestions.Update(msg)
			cmds = append(cmds, suggestionCmd)
			break
		}

		beforeValue := m.input.Value()
		beforeCursor := m.input.CursorPosition()
		var inputCmd tea.Cmd
		m.input, inputCmd = m.input.Update(msg)
		cmds = append(cmds, inputCmd)
		if m.input.Value() != beforeValue || m.input.CursorPosition() != beforeCursor {
			m.syncQuery()
		}
		if m.input.Value() != beforeValue {
			m.layout()
		}
	}
	m.syncStatus()
	return m, tea.Batch(cmds...)
}

func isCompletionKey(msg tea.KeyPressMsg) bool {
	switch msg.String() {
	case "up", "down", "k", "j", "pgup", "pgdown", "home", "g", "end", "G", "enter":
		return true
	default:
		return false
	}
}

func (m model) completionActive() bool {
	return m.suggestions.Query() != "" && m.suggestions.FilteredLen() > 0
}

func (m model) tokenRange() (textarea.Position, textarea.Position) {
	end := m.input.CursorPosition()
	lines := strings.Split(m.input.Value(), "\n")
	if end.Row < 0 || end.Row >= len(lines) {
		return end, end
	}
	line := []rune(lines[end.Row])
	column := min(end.Column, len(line))
	start := column
	for start > 0 && !unicode.IsSpace(line[start-1]) {
		start--
	}
	return textarea.Position{Row: end.Row, Column: start}, end
}

func (m *model) applyCompletion(msg autocomplete.SelectedMsg) {
	start, end := m.tokenRange()
	m.input.ReplaceRange(start, end, msg.Value)
	m.syncQuery()
	m.layout()
}

func (m *model) syncQuery() {
	start, end := m.tokenRange()
	value := m.input.Value()
	lines := strings.Split(value, "\n")
	if start.Row < 0 || start.Row >= len(lines) || end.Row != start.Row {
		m.suggestions.SetQuery("")
		return
	}
	runes := []rune(lines[end.Row])
	startCol := min(start.Column, len(runes))
	endCol := min(end.Column, len(runes))
	m.suggestions.SetQuery(string(runes[startCol:endCol]))
}

func (m *model) layout() {
	var main, footer layout.Rect
	layout.Vertical(layout.Fill(1), layout.Len(1)).Split(
		layout.NewRect(0, 0, m.width, m.height),
	).Assign(&main, &footer)
	content := frame.PanelContentRect(main, frame.PanelOptions{Title: "Textarea completion", Focused: m.input.Focused(), Padding: 1})
	inputHeight := min(3, max(1, m.input.ContentHeight()))
	var input, suggestions layout.Rect
	layout.Vertical(layout.Len(inputHeight), layout.Fill(1)).Split(content).Assign(&input, &suggestions)
	m.input.SetSize(input.Dx(), input.Dy())
	m.suggestions.SetSize(suggestions.Dx(), suggestions.Dy())
	m.status.SetSize(footer.Dx(), footer.Dy())
}

func (m *model) syncStatus() {
	selected := m.suggestions.SelectedID()
	if selected == "" {
		selected = "none"
	}
	m.status.SetLeft(
		statusbar.Segment{Text: "textarea", Kind: statusbar.KindAccent},
		statusbar.Segment{Text: selected, Kind: statusbar.KindNormal},
	)
	m.status.SetRight(
		statusbar.Segment{Text: fmt.Sprintf("%d matches", m.suggestions.FilteredLen()), Kind: statusbar.KindMuted},
		statusbar.Segment{Text: "↑↓ choose  enter accept", Kind: statusbar.KindInfo},
	)
}

func (m model) render() string {
	if m.width < 20 || m.height < 7 {
		return "resize terminal"
	}
	content := lipgloss.JoinVertical(lipgloss.Left, m.input.View(), m.suggestions.View())
	panel := frame.Panel(m.theme, content, m.width, frame.PanelOptions{
		Title:   "Textarea completion",
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
