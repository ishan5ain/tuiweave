// Command statusbar demonstrates the tuiweave statusbar component and the
// canonical app wiring pattern: window size → layout split → SetSize →
// composed string view.
//
// Run it with:
//
//	go run ./examples/statusbar
//
// Press t to toggle the theme, q to quit.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/layout"
	"github.com/ishan5ain/tuiweave/statusbar"
)

type model struct {
	dark          bool
	width, height int
	body          layout.Rect
	status        statusbar.Model
}

func newModel() model {
	m := model{dark: true}
	m.rebuildStatusbar()
	return m
}

func (m model) theme() tuiweave.Theme {
	if m.dark {
		return tuiweave.Dark()
	}
	return tuiweave.Light()
}

// rebuildStatusbar recreates the bar from the current theme and state.
// Styles are derived from the theme at construction, so a theme change
// means constructing a new component.
func (m *model) rebuildStatusbar() {
	sb := statusbar.New(m.theme())
	sb.SetSize(m.width, 1)
	sb.SetLeft(
		statusbar.Segment{Text: "tuiweave", Kind: statusbar.KindAccent},
		statusbar.Segment{Text: "examples/statusbar", Kind: statusbar.KindNormal},
	)
	themeName := "light"
	if m.dark {
		themeName = "dark"
	}
	sb.SetRight(
		statusbar.Segment{Text: "t: theme", Kind: statusbar.KindInfo},
		statusbar.Segment{Text: themeName, Kind: statusbar.KindMuted},
		statusbar.Segment{Text: fmt.Sprintf("%d×%d", m.width, m.height), Kind: statusbar.KindMuted},
	)
	m.status = sb
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		var status layout.Rect
		layout.Vertical(layout.Fill(1), layout.Len(1)).
			Split(layout.NewRect(0, 0, msg.Width, msg.Height)).
			Assign(&m.body, &status)
		m.rebuildStatusbar()
		m.status.SetSize(status.Dx(), status.Dy())
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "t":
			m.dark = !m.dark
			m.rebuildStatusbar()
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	theme := m.theme()
	body := lipgloss.NewStyle().
		Width(m.body.Dx()).
		Height(m.body.Dy()).
		Padding(1, 2).
		Background(theme.Surface).
		Foreground(theme.Text).
		Render("tuiweave statusbar demo\n\nPress t to toggle the theme, q to quit.")
	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, body, m.status.View()))
	v.AltScreen = true
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
