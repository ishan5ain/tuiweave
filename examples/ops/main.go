// Command ops is a non-agentic operations-console reference app. It composes
// the Phase 4 vocabulary into a small service dashboard with tabs, actions,
// a table, status controls, and an overlay command palette.
//
//	go run ./examples/ops
//
// tab/shift+tab cycles focus, ctrl+p opens commands, and q quits.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/button"
	"github.com/ishansain/gotui/focus"
	"github.com/ishansain/gotui/frame"
	"github.com/ishansain/gotui/inspect"
	"github.com/ishansain/gotui/layout"
	"github.com/ishansain/gotui/line"
	"github.com/ishansain/gotui/menu"
	"github.com/ishansain/gotui/overlay"
	"github.com/ishansain/gotui/palette"
	"github.com/ishansain/gotui/progress"
	"github.com/ishansain/gotui/splitpane"
	"github.com/ishansain/gotui/stack"
	"github.com/ishansain/gotui/statusbar"
	"github.com/ishansain/gotui/table"
	"github.com/ishansain/gotui/tabs"
	"github.com/ishansain/gotui/toggle"
)

type model struct {
	theme         gotui.Theme
	width, height int

	tabs        tabs.Model
	actions     menu.Model
	rows        table.Model
	load        progress.Model
	autoRefresh toggle.Model
	openLogs    button.Model
	commands    palette.Model
	status      statusbar.Model

	fm          focus.Manager
	showPalette bool
	notice      string
}

func newModel() model {
	theme := gotui.Dark()
	m := model{
		theme:       theme,
		tabs:        tabs.New(theme),
		actions:     menu.New(theme),
		rows:        table.New(theme),
		load:        progress.New(theme),
		autoRefresh: toggle.New(theme),
		openLogs:    button.New(theme),
		commands:    palette.New(theme),
		status:      statusbar.New(theme),
		fm:          focus.NewManager(5),
		notice:      "ready",
	}
	m.tabs.SetTabs(
		tabs.Tab{ID: "services", Label: "Services"},
		tabs.Tab{ID: "jobs", Label: "Jobs"},
		tabs.Tab{ID: "audit", Label: "Audit"},
	)
	m.actions.SetItems(
		menu.Item{ID: "restart", Label: "Restart service", Description: "Restart the selected service"},
		menu.Item{ID: "drain", Label: "Drain traffic", Description: "Remove a service from rotation"},
		menu.Item{ID: "ack", Label: "Acknowledge alert", Description: "Mark the current alert handled"},
		menu.Item{ID: "delete", Label: "Delete service", Description: "Destructive action", Disabled: true},
	)
	m.load.SetLabel("Deploy")
	m.load.SetPercent(0.72)
	m.load.SetStatus(progress.StatusInfo)
	m.autoRefresh.ID = "auto-refresh"
	m.autoRefresh.SetLabel("Auto-refresh")
	m.autoRefresh.SetChecked(true)
	m.openLogs.ID = "open-logs"
	m.openLogs.SetLabel("Open logs")
	m.commands.SetItems(
		palette.Item{ID: "restart", Label: "Restart service", Description: "Restart the selected service"},
		palette.Item{ID: "drain", Label: "Drain traffic", Description: "Remove a service from rotation"},
		palette.Item{ID: "ack", Label: "Acknowledge alert", Description: "Mark the current alert handled"},
		palette.Item{ID: "refresh", Label: "Refresh data", Description: "Reload service state"},
		palette.Item{ID: "delete", Label: "Delete service", Description: "Destructive action", Disabled: true},
	)
	m.syncRows()
	m.applyFocus()
	m.syncStatus()
	return m
}

func (m *model) applyFocus() {
	m.fm.Apply(&m.tabs, &m.actions, &m.rows, &m.autoRefresh, &m.openLogs)
}

func (m *model) syncRows() {
	switch m.tabs.SelectedID() {
	case "jobs":
		m.rows.SetColumns(
			table.Column{Title: "Job"},
			table.Column{Title: "State", Width: 10},
			table.Column{Title: "Owner", Width: 12},
		)
		m.rows.SetRows(
			[]string{"nightly-backup", "running", "platform"},
			[]string{"index-rebuild", "queued", "search"},
			[]string{"billing-rollup", "healthy", "finance"},
		)
	case "audit":
		m.rows.SetColumns(
			table.Column{Title: "Event"},
			table.Column{Title: "Actor", Width: 12},
			table.Column{Title: "Age", Width: 8},
		)
		m.rows.SetRows(
			[]string{"deploy completed", "release-bot", "2m"},
			[]string{"config changed", "isha", "18m"},
			[]string{"alert cleared", "on-call", "1h"},
		)
	default:
		m.rows.SetColumns(
			table.Column{Title: "Service"},
			table.Column{Title: "State", Width: 10},
			table.Column{Title: "Region", Width: 12},
		)
		m.rows.SetRows(
			[]string{"api", "healthy", "us-west-2"},
			[]string{"worker", "busy", "us-west-2"},
			[]string{"scheduler", "degraded", "us-east-1"},
		)
	}
}

func (m *model) syncStatus() {
	focusName := "palette"
	if !m.showPalette {
		focusName = [...]string{"tabs", "actions", "table", "refresh", "open logs"}[m.fm.Index()]
	}
	m.status.SetLeft(
		statusbar.Segment{Text: "ops", Kind: statusbar.KindAccent},
		statusbar.Segment{Text: focusName, Kind: statusbar.KindNormal},
	)
	m.status.SetRight(
		statusbar.Segment{Text: m.tabs.SelectedID(), Kind: statusbar.KindMuted},
		statusbar.Segment{Text: m.notice, Kind: statusbar.KindInfo},
	)
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()
	case palette.SelectedMsg:
		m.showPalette = false
		m.commands.Blur()
		m.notice = "ran " + msg.Label
	case menu.SelectedMsg:
		m.notice = "ran " + msg.Label
	case toggle.ChangedMsg:
		m.notice = fmt.Sprintf("%s %v", msg.ID, msg.Checked)
	case button.PressedMsg:
		m.notice = "opened logs"
	case tea.KeyPressMsg:
		if m.showPalette {
			m.commands, cmd = m.commands.Update(msg)
			cmds = append(cmds, cmd)
			break
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "ctrl+p":
			m.showPalette = true
			m.commands.SetQuery("")
			m.commands.Focus()
		default:
			beforeTab := m.tabs.SelectedID()
			switch msg.String() {
			case "tab":
				m.fm.Next()
				m.applyFocus()
			case "shift+tab":
				m.fm.Prev()
				m.applyFocus()
			default:
				m.tabs, cmd = m.tabs.Update(msg)
				cmds = append(cmds, cmd)
				m.actions, cmd = m.actions.Update(msg)
				cmds = append(cmds, cmd)
				m.rows, cmd = m.rows.Update(msg)
				cmds = append(cmds, cmd)
				m.autoRefresh, cmd = m.autoRefresh.Update(msg)
				cmds = append(cmds, cmd)
				m.openLogs, cmd = m.openLogs.Update(msg)
				cmds = append(cmds, cmd)
			}
			if beforeTab != m.tabs.SelectedID() {
				m.syncRows()
			}
		}
	}
	m.syncStatus()
	return m, tea.Batch(cmds...)
}

func (m *model) layout() {
	var header, body, footer layout.Rect
	layout.Vertical(
		layout.Len(2),
		layout.Fill(1),
		layout.Len(1),
	).Split(layout.NewRect(0, 0, m.width, m.height)).Assign(&header, &body, &footer)
	m.tabs.SetSize(header.Dx(), 1)

	var actions, operations layout.Rect
	layout.Horizontal(layout.Fill(1), layout.Fill(1)).WithSpacing(1).Split(body).Assign(&actions, &operations)
	m.actions.SetSize(max(0, actions.Dx()-4), max(0, actions.Dy()-4))
	innerWidth := max(0, operations.Dx()-4)
	m.rows.SetSize(innerWidth, max(3, operations.Dy()-7))
	m.load.SetSize(innerWidth, 1)
	m.autoRefresh.SetSize(innerWidth, 1)
	m.openLogs.SetSize(innerWidth, 1)

	m.status.SetSize(footer.Dx(), footer.Dy())
	paletteWidth := min(56, max(16, m.width-4))
	paletteHeight := min(8, max(5, m.height-6))
	m.commands.SetSize(max(1, paletteWidth-4), max(1, paletteHeight-4))
}

func (m model) render() string {
	if m.width < 20 || m.height < 10 {
		return "resize terminal"
	}

	var header, body, footer layout.Rect
	layout.Vertical(layout.Len(2), layout.Fill(1), layout.Len(1)).
		Split(layout.NewRect(0, 0, m.width, m.height)).Assign(&header, &body, &footer)

	headerStyle := lipgloss.NewStyle().
		Width(header.Dx()).
		Foreground(m.theme.Text).
		Background(m.theme.Surface)
	headerLine := line.Join(header.Dx(),
		lipgloss.JoinHorizontal(lipgloss.Top,
			frame.Badge(m.theme, "OPS", frame.BadgeAccent),
			" ",
			lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render("operations console"),
		),
		"ctrl+p commands",
		line.JoinOptions{Gap: 1})
	headerView := stack.Vertical(m.theme, header.Dx(), stack.Options{},
		func(int) string { return headerStyle.Render(headerLine) },
		func(int) string { return m.tabs.View() },
	)

	var actions, operations layout.Rect
	layout.Horizontal(layout.Fill(1), layout.Fill(1)).WithSpacing(1).Split(body).Assign(&actions, &operations)
	actionsView := frame.Panel(m.theme, m.actions.View(), actions.Dx(), frame.PanelOptions{
		Title:   "Actions",
		Focused: m.actions.Focused(),
		Padding: 1,
	})
	innerWidth := max(0, operations.Dx()-4)
	operationsContent := stack.Vertical(m.theme, innerWidth, stack.Options{Gap: 0},
		func(int) string { return m.rows.View() },
		func(int) string { return m.load.View() },
		func(int) string { return m.autoRefresh.View() },
		func(int) string { return m.openLogs.View() },
	)
	operationsView := frame.Panel(m.theme, operationsContent, operations.Dx(), frame.PanelOptions{
		Title:   "Operations",
		Focused: m.rows.Focused() || m.autoRefresh.Focused() || m.openLogs.Focused(),
		Padding: 1,
	})
	bodyView := splitpane.Horizontal(m.theme, body.Dx(), splitpane.Options{Gap: 1},
		func(int) string { return actionsView },
		func(int) string { return operationsView },
	)
	footerView := lipgloss.NewStyle().
		Width(footer.Dx()).
		Foreground(m.theme.TextMuted).
		Render(line.Join(footer.Dx(), frame.Divider(m.theme, footer.Dx()), "tab focus · ctrl+p commands · q quit", line.JoinOptions{Gap: 1, NoEllipsis: true}))

	base := stack.Vertical(m.theme, m.width, stack.Options{},
		func(int) string { return headerView },
		func(int) string { return bodyView },
		func(int) string { return footerView },
	)
	if !m.showPalette {
		return base
	}
	paletteWidth := min(56, max(16, m.width-4))
	prompt := frame.Panel(m.theme, m.commands.View(), paletteWidth, frame.PanelOptions{
		Title:   "Command palette",
		Focused: true,
		Padding: 1,
	})
	return overlay.Center(base, prompt)
}

func (m model) Inspect() inspect.Node {
	children := []inspect.Node{
		inspect.Bind("tabs", m.tabs),
		inspect.Bind("actions", m.actions),
		inspect.Bind("operations", m.rows),
		inspect.Bind("load", m.load),
		inspect.Bind("auto-refresh", m.autoRefresh),
		inspect.Bind("open-logs", m.openLogs),
	}
	if m.showPalette {
		children = append(children, inspect.Bind("commands", m.commands))
	}
	return inspect.Group("ops", "application", inspect.Bounds{Width: m.width, Height: m.height}, children...)
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
