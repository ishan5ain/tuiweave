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
	"github.com/ishansain/gotui/focus"
	"github.com/ishansain/gotui/frame"
	"github.com/ishansain/gotui/layout"
	"github.com/ishansain/gotui/menu"
	"github.com/ishansain/gotui/splitpane"
	"github.com/ishansain/gotui/tabs"
	"github.com/ishansain/gotui/toolbar"
)

type model struct {
	theme         gotui.Theme
	width, height int
	nav           tabs.Model
	actions       menu.Model
	tools         toolbar.Model
	fm            focus.Manager
}

func newModel() model {
	theme := gotui.Dark()
	nav := tabs.New(theme)
	nav.SetTabs(
		tabs.Tab{ID: "overview", Label: "Overview"},
		tabs.Tab{ID: "activity", Label: "Activity"},
		tabs.Tab{ID: "settings", Label: "Settings"},
	)
	tools := toolbar.New(theme)
	tools.SetItems(
		toolbar.Item{ID: "refresh", Label: "Refresh", Description: "Reload data"},
		toolbar.Item{ID: "export", Label: "Export", Description: "Export data"},
		toolbar.Item{ID: "delete", Label: "Delete", Description: "Destructive action", Disabled: true},
	)
	actions := menu.New(theme)
	actions.SetItems(
		menu.Item{ID: "open", Label: "Open workspace", Description: "Open a workspace"},
		menu.Item{ID: "refresh", Label: "Refresh data", Description: "Reload current data"},
		menu.Item{ID: "delete", Label: "Delete workspace", Description: "Destructive action", Disabled: true},
		menu.Item{ID: "quit", Label: "Quit", Description: "Close the application"},
	)
	m := model{theme: theme, nav: nav, actions: actions, tools: tools, fm: focus.NewManager(3)}
	m.fm.Apply(&m.nav, &m.actions, &m.tools)
	return m
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()
	case toolbar.SelectedMsg:
		// This showcase has no backend action to perform; a real app handles
		// msg.ID here.
		return m, nil
	case menu.SelectedMsg:
		// This showcase has no backend action to perform; a real app handles
		// msg.ID here.
		return m, nil
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
		if msg.String() == "tab" {
			m.fm.Next()
			m.fm.Apply(&m.nav, &m.actions, &m.tools)
			return m, nil
		}
		if msg.String() == "shift+tab" {
			m.fm.Prev()
			m.fm.Apply(&m.nav, &m.actions, &m.tools)
			return m, nil
		}
		var cmds []tea.Cmd
		var cmd tea.Cmd
		m.nav, cmd = m.nav.Update(msg)
		cmds = append(cmds, cmd)
		m.actions, cmd = m.actions.Update(msg)
		cmds = append(cmds, cmd)
		m.tools, cmd = m.tools.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m *model) layout() {
	var header, body, footer layout.Rect
	layout.Vertical(
		layout.Len(2),
		layout.Fill(1),
		layout.Len(1),
	).Split(layout.NewRect(0, 0, m.width, m.height)).Assign(&header, &body, &footer)
	m.nav.SetSize(header.Dx(), 1)

	var jobs, services layout.Rect
	layout.Horizontal(layout.Fill(1), layout.Fill(1)).WithSpacing(1).Split(body).Assign(&jobs, &services)
	m.actions.SetSize(max(0, jobs.Dx()-4), max(0, jobs.Dy()-4))
	m.tools.SetSize(max(0, services.Dx()-4), 1)
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
	headerLine := headerStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			frame.Badge(m.theme, "OPS", frame.BadgeAccent),
			" ",
			lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render("service overview"),
		),
	)
	headerView := lipgloss.JoinVertical(lipgloss.Left, headerLine, m.nav.View())

	var jobs, services layout.Rect
	layout.Horizontal(layout.Fill(1), layout.Fill(1)).WithSpacing(1).Split(body).Assign(&jobs, &services)

	jobsView := frame.Panel(m.theme,
		m.actions.View(),
		jobs.Dx(),
		frame.PanelOptions{Title: "Actions", Focused: m.actions.Focused(), Padding: 1},
	)
	servicesView := frame.Panel(m.theme,
		"api        "+frame.Badge(m.theme, "healthy", frame.BadgeSuccess)+"\nworker     "+frame.Badge(m.theme, "busy", frame.BadgeInfo)+"\n"+m.tools.View(),
		services.Dx(),
		frame.PanelOptions{Title: "Services", Focused: m.tools.Focused(), Padding: 1},
	)

	footerView := lipgloss.NewStyle().
		Width(footer.Dx()).
		Foreground(m.theme.TextMuted).
		Render(frame.Divider(m.theme, footer.Dx()-1) + " q quit")

	bodyView := splitpane.Horizontal(m.theme, body.Dx(), splitpane.Options{Gap: 1},
		func(int) string { return jobsView },
		func(int) string { return servicesView },
	)
	return lipgloss.JoinVertical(lipgloss.Left, headerView, bodyView, footerView)
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
