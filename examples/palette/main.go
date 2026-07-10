// Command palette demonstrates the reusable command-palette foundation with
// stable action IDs, filtering, semantic activation, and a framed app shell.
//
//	go run ./examples/palette
//
// Type to filter, use up/down to navigate, and press enter to activate.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/frame"
	"github.com/ishansain/gotui/layout"
	"github.com/ishansain/gotui/palette"
	"github.com/ishansain/gotui/statusbar"
)

type model struct {
	theme         gotui.Theme
	width, height int
	commands      palette.Model
	status        statusbar.Model
}

func newModel() model {
	theme := gotui.Dark()
	commands := palette.New(theme)
	commands.SetItems(
		palette.Item{ID: "open", Label: "Open workspace", Description: "Choose a workspace"},
		palette.Item{ID: "format", Label: "Format document", Description: "Run the formatter"},
		palette.Item{ID: "search", Label: "Search files", Description: "Find a file by name"},
		palette.Item{ID: "delete", Label: "Delete workspace", Description: "Destructive action", Disabled: true},
		palette.Item{ID: "settings", Label: "Open settings", Description: "Edit preferences"},
	)
	commands.Focus()
	return model{
		theme:    theme,
		commands: commands,
		status:   statusbar.New(theme),
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()
	case palette.SelectedMsg:
		m.status.SetLeft(
			statusbar.Segment{Text: "selected", Kind: statusbar.KindAccent},
			statusbar.Segment{Text: msg.Label, Kind: statusbar.KindNormal},
		)
		return m, nil
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
		var cmd tea.Cmd
		m.commands, cmd = m.commands.Update(msg)
		cmds = append(cmds, cmd)
	}
	m.syncStatus()
	return m, tea.Batch(cmds...)
}

func (m *model) layout() {
	var main, footer layout.Rect
	layout.Vertical(
		layout.Fill(1),
		layout.Len(1),
	).Split(layout.NewRect(0, 0, m.width, m.height)).Assign(&main, &footer)
	m.commands.SetSize(max(0, main.Dx()-4), max(0, main.Dy()-4))
	m.status.SetSize(footer.Dx(), footer.Dy())
}

func (m *model) syncStatus() {
	selected := m.commands.SelectedID()
	if selected == "" {
		selected = "none"
	}
	m.status.SetLeft(
		statusbar.Segment{Text: "palette", Kind: statusbar.KindAccent},
		statusbar.Segment{Text: selected, Kind: statusbar.KindNormal},
	)
	m.status.SetRight(
		statusbar.Segment{Text: fmt.Sprintf("%d/%d", m.commands.FilteredLen(), len(m.commands.Items())), Kind: statusbar.KindMuted},
		statusbar.Segment{Text: "enter run", Kind: statusbar.KindInfo},
	)
}

func (m model) render() string {
	if m.width < 10 || m.height < 6 {
		return "resize terminal"
	}
	var main, footer layout.Rect
	layout.Vertical(layout.Fill(1), layout.Len(1)).Split(layout.NewRect(0, 0, m.width, m.height)).Assign(&main, &footer)
	content := frame.Panel(m.theme, m.commands.View(), m.width, frame.PanelOptions{
		Title:   "Commands",
		Focused: m.commands.Focused(),
		Padding: 1,
	})
	return lipgloss.JoinVertical(lipgloss.Left, content, m.status.View())
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
