// Command frame demonstrates gotui's domain-neutral framing helpers in a
// small dashboard-like composition.
//
//	go run ./examples/frame
//
// The example intentionally keeps orchestration in the app: layout divides the
// window, while frame returns styled strings that the app composes.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/frame"
	"github.com/ishansain/gotui/layout"
)

type model struct {
	theme         gotui.Theme
	width, height int
}

func newModel() model {
	return model{theme: gotui.Dark()}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) render() string {
	if m.width < 10 || m.height < 5 {
		return "resize terminal"
	}

	var header, body, footer layout.Rect
	layout.Vertical(
		layout.Len(2),
		layout.Fill(1),
		layout.Len(1),
	).Split(layout.NewRect(0, 0, m.width, m.height)).Assign(&header, &body, &footer)

	headerStyle := lipgloss.NewStyle().
		Width(header.Dx()).
		Foreground(m.theme.Text).
		Background(m.theme.Surface)
	headerView := headerStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			frame.Badge(m.theme, "OPS", frame.BadgeAccent),
			" ",
			lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render("service overview"),
		),
	)

	var jobs, services layout.Rect
	layout.Horizontal(layout.Fill(1), layout.Fill(1)).WithSpacing(1).Split(body).Assign(&jobs, &services)

	jobsView := frame.Panel(m.theme,
		"queue      3\ncompleted  18\nfailed      0",
		jobs.Dx(),
		frame.PanelOptions{Title: "Jobs", Focused: true, Padding: 1},
	)
	servicesView := frame.Panel(m.theme,
		"api        "+frame.Badge(m.theme, "healthy", frame.BadgeSuccess)+"\nworker     "+frame.Badge(m.theme, "busy", frame.BadgeInfo),
		services.Dx(),
		frame.PanelOptions{Title: "Services", Padding: 1},
	)

	footerView := lipgloss.NewStyle().
		Width(footer.Dx()).
		Foreground(m.theme.TextMuted).
		Render(frame.Divider(m.theme, footer.Dx()-1) + " q quit")

	return lipgloss.JoinVertical(lipgloss.Left, headerView, lipgloss.JoinHorizontal(lipgloss.Top, jobsView, " ", servicesView), footerView)
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
