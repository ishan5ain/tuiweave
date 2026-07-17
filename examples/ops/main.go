// Command ops is a non-agentic operations-console reference app. It composes
// the Phase 4 vocabulary into a small service dashboard with tabs, actions,
// a table, status controls, and an overlay command palette.
//
//	go run ./examples/ops
//
// tab/shift+tab cycles focus, ctrl+p opens commands, selecting restart opens a
// nested confirmation, and q quits.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/action"
	"github.com/ishan5ain/tuiweave/button"
	"github.com/ishan5ain/tuiweave/dialog"
	"github.com/ishan5ain/tuiweave/focus"
	"github.com/ishan5ain/tuiweave/frame"
	"github.com/ishan5ain/tuiweave/inspect"
	"github.com/ishan5ain/tuiweave/layout"
	"github.com/ishan5ain/tuiweave/line"
	"github.com/ishan5ain/tuiweave/menu"
	"github.com/ishan5ain/tuiweave/mouse"
	"github.com/ishan5ain/tuiweave/overlay"
	"github.com/ishan5ain/tuiweave/palette"
	"github.com/ishan5ain/tuiweave/progress"
	"github.com/ishan5ain/tuiweave/splitpane"
	"github.com/ishan5ain/tuiweave/stack"
	"github.com/ishan5ain/tuiweave/statusbar"
	"github.com/ishan5ain/tuiweave/table"
	"github.com/ishan5ain/tuiweave/tabs"
	"github.com/ishan5ain/tuiweave/toggle"
)

type model struct {
	theme         tuiweave.Theme
	width, height int

	tabs        tabs.Model
	actions     menu.Model
	rows        table.Model
	load        progress.Model
	autoRefresh toggle.Model
	openLogs    button.Model
	commands    palette.Model
	confirm     dialog.Model
	status      statusbar.Model

	fm          focus.Stack
	showPalette bool
	showConfirm bool
	notice      string

	// The app retains the screen-space rectangles it owns during layout. They
	// are used both for routing/inspection and to keep semantic bounds tied to
	// the same rectangles that size the components.
	tabsArea, actionsArea                             layout.Rect
	rowsArea, loadArea, autoRefreshArea, openLogsArea layout.Rect
	commandsArea, commandsPanelArea, confirmArea      layout.Rect
}

func newModel() model {
	theme := tuiweave.Dark()
	m := model{
		theme:       theme,
		tabs:        tabs.New(theme),
		actions:     menu.New(theme),
		rows:        table.New(theme),
		load:        progress.New(theme),
		autoRefresh: toggle.New(theme),
		openLogs:    button.New(theme),
		commands:    palette.New(theme),
		confirm:     dialog.New(theme),
		status:      statusbar.New(theme),
		fm:          focus.NewStack(5),
		notice:      "ready",
	}
	m.tabs.SetTabs(
		tabs.Tab{ID: "services", Label: "Services"},
		tabs.Tab{ID: "jobs", Label: "Jobs"},
		tabs.Tab{ID: "audit", Label: "Audit"},
	)
	operations := []action.Item{
		{ID: "restart", Label: "Restart service", Description: "Restart the selected service"},
		{ID: "drain", Label: "Drain traffic", Description: "Remove a service from rotation"},
		{ID: "ack", Label: "Acknowledge alert", Description: "Mark the current alert handled"},
		{ID: "delete", Label: "Delete service", Description: "Destructive action", Disabled: true},
	}
	m.actions.SetItems(operations...)
	m.load.SetLabel("Deploy")
	m.load.SetPercent(0.72)
	m.load.SetStatus(progress.StatusInfo)
	m.autoRefresh.ID = "auto-refresh"
	m.autoRefresh.SetLabel("Auto-refresh")
	m.autoRefresh.SetChecked(true)
	m.openLogs.ID = "open-logs"
	m.openLogs.SetLabel("Open logs")
	m.confirm.ID = "restart"
	m.confirm.Title = "Restart service?"
	m.confirm.Body = "Restarting the selected service may interrupt traffic."
	m.confirm.ConfirmLabel = "Restart"
	m.confirm.CancelLabel = "Cancel"
	paletteItems := append([]action.Item(nil), operations[:3]...)
	paletteItems = append(paletteItems, action.Item{
		ID: "refresh", Label: "Refresh data", Description: "Reload service state",
	})
	paletteItems = append(paletteItems, operations[3])
	m.commands.SetItems(paletteItems...)
	m.syncRows()
	m.applyFocus()
	m.syncStatus()
	return m
}

func (m *model) applyFocus() {
	m.fm.Apply(
		focus.Group{&m.tabs, &m.actions, &m.rows, &m.autoRefresh, &m.openLogs},
		focus.Group{&m.commands},
		focus.Group{}, // dialog routes its own button selection while visible
	)
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
	focusName := "confirm"
	switch m.fm.Depth() {
	case 0:
		focusName = [...]string{"tabs", "actions", "table", "refresh", "open logs"}[m.fm.Index()]
	case 1:
		focusName = "palette"
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
	case inspect.ActionMsg:
		if cmd, handled := m.dispatchAction(msg.ID); handled {
			cmds = append(cmds, cmd)
		}
	case palette.SelectedMsg:
		if msg.ID == "restart" {
			m.showConfirm = true
			m.fm.Push(0)
			m.applyFocus()
			m.notice = "confirm restart"
			break
		}
		m.showPalette = false
		m.fm.Pop()
		m.applyFocus()
		m.notice = "ran " + msg.Label
	case dialog.ResultMsg:
		if !m.showConfirm {
			break
		}
		m.showConfirm = false
		m.fm.Pop()
		m.applyFocus()
		if msg.OK {
			m.showPalette = false
			m.fm.Pop()
			m.applyFocus()
			m.notice = "restarted service"
		} else {
			m.notice = "restart cancelled"
		}
	case menu.SelectedMsg:
		m.notice = "ran " + msg.Label
	case toggle.ChangedMsg:
		m.notice = fmt.Sprintf("%s %v", msg.ID, msg.Checked)
	case button.PressedMsg:
		m.notice = "opened logs"
	case tea.MouseClickMsg:
		// Nested layers own all input while visible. In particular, a click on
		// the dimmed application must not move root focus or activate controls.
		if m.showConfirm || m.showPalette || msg.Button != tea.MouseLeft {
			break
		}
		cmds = append(cmds, m.routeRootClick(msg))
	case tea.KeyPressMsg:
		if m.showConfirm {
			m.confirm, cmd = m.confirm.Update(msg)
			cmds = append(cmds, cmd)
			break
		}
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
			m.fm.Push(1)
			m.applyFocus()
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

// routeRootClick uses the same screen-space rectangles retained by layout and
// Inspect. Components render local strings, so the application translates a
// global click into local selection, focus, and activation decisions.
func (m *model) routeRootClick(msg tea.MouseClickMsg) tea.Cmd {
	switch {
	case mouse.InBounds(msg, m.tabsArea.Min.X, m.tabsArea.Min.Y, m.tabsArea.Dx(), m.tabsArea.Dy()):
		m.focusRoot(0)
	case mouse.InBounds(msg, m.actionsArea.Min.X, m.actionsArea.Min.Y, m.actionsArea.Dx(), m.actionsArea.Dy()):
		m.actions.Select(m.actions.YOffset() + msg.Y - m.actionsArea.Min.Y)
		m.focusRoot(1)
	case mouse.InBounds(msg, m.rowsArea.Min.X, m.rowsArea.Min.Y, m.rowsArea.Dx(), m.rowsArea.Dy()):
		// The table header and rule occupy its first two rows. Clicking either
		// focuses the table without changing its selected data row.
		if row := msg.Y - m.rowsArea.Min.Y - 2; row >= 0 {
			if index := m.rows.YOffset() + row; index < m.rows.TotalLines() {
				m.rows.Select(index)
			}
		}
		m.focusRoot(2)
	case mouse.InBounds(msg, m.autoRefreshArea.Min.X, m.autoRefreshArea.Min.Y, m.autoRefreshArea.Dx(), m.autoRefreshArea.Dy()):
		m.focusRoot(3)
		next, cmd := m.autoRefresh.Update(inspect.Invoke(toggle.ActionToggle))
		m.autoRefresh = next
		return cmd
	case mouse.InBounds(msg, m.openLogsArea.Min.X, m.openLogsArea.Min.Y, m.openLogsArea.Dx(), m.openLogsArea.Dy()):
		m.focusRoot(4)
		next, cmd := m.openLogs.Update(inspect.Invoke(button.ActionActivate))
		m.openLogs = next
		return cmd
	}
	return nil
}

func (m *model) focusRoot(index int) {
	m.fm.Set(index)
	m.applyFocus()
}

func (m *model) layout() {
	var header, body, footer layout.Rect
	layout.Vertical(
		layout.Len(2),
		layout.Fill(1),
		layout.Len(1),
	).Split(layout.NewRect(0, 0, m.width, m.height)).Assign(&header, &body, &footer)
	m.tabsArea = layout.NewRect(header.Min.X, header.Min.Y+1, header.Dx(), 1)
	m.tabs.SetSize(header.Dx(), 1)

	var actions, operations layout.Rect
	layout.Horizontal(layout.Fill(1), layout.Fill(1)).WithSpacing(1).Split(body).Assign(&actions, &operations)
	actionContent := frame.PanelContentRect(actions, frame.PanelOptions{Padding: 1})
	m.actionsArea = actionContent
	m.actions.SetSize(actionContent.Dx(), actionContent.Dy())
	operationContent := frame.PanelContentRect(operations, frame.PanelOptions{Padding: 1})
	layout.Vertical(
		layout.Fill(1),
		layout.Len(1),
		layout.Len(1),
		layout.Len(1),
	).Split(operationContent).Assign(&m.rowsArea, &m.loadArea, &m.autoRefreshArea, &m.openLogsArea)
	m.rows.SetSize(m.rowsArea.Dx(), m.rowsArea.Dy())
	m.load.SetSize(m.loadArea.Dx(), m.loadArea.Dy())
	m.autoRefresh.SetSize(m.autoRefreshArea.Dx(), m.autoRefreshArea.Dy())
	m.openLogs.SetSize(m.openLogsArea.Dx(), m.openLogsArea.Dy())

	m.status.SetSize(footer.Dx(), footer.Dy())
	paletteWidth := min(56, max(16, m.width-4))
	paletteHeight := min(8, max(5, m.height-6))
	m.commands.SetSize(max(1, paletteWidth-4), max(1, paletteHeight-4))
	m.confirm.SetSize(min(44, max(1, m.width-4)), 10)

	paletteOptions := frame.PanelOptions{
		Title:   "Command palette",
		Focused: true,
		Padding: 1,
	}
	m.commandsPanelArea = centeredArea(
		layout.NewRect(0, 0, m.width, m.height),
		frame.Panel(m.theme, m.commands.View(), paletteWidth, paletteOptions),
	)
	m.commandsArea = frame.PanelContentRect(m.commandsPanelArea, paletteOptions)
	m.confirmArea = centeredArea(
		layout.NewRect(0, 0, m.width, m.height),
		m.confirm.View(),
	)
}

// centeredArea returns the screen-space rectangle used by overlay.Center for
// a rendered overlay. Components inside a decorative panel use a separate
// content rectangle derived from frame.PanelContentRect.
func centeredArea(base layout.Rect, view string) layout.Rect {
	width, height := lipgloss.Width(view), lipgloss.Height(view)
	return layout.NewRect(
		base.Min.X+max(0, (base.Dx()-width)/2),
		base.Min.Y+max(0, (base.Dy()-height)/2),
		width,
		height,
	)
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
	operationOptions := frame.PanelOptions{
		Title:   "Operations",
		Focused: m.rows.Focused() || m.autoRefresh.Focused() || m.openLogs.Focused(),
		Padding: 1,
	}
	innerWidth := frame.PanelContentRect(operations, operationOptions).Dx()
	operationsContent := stack.Vertical(m.theme, innerWidth, stack.Options{Gap: 0},
		func(int) string { return m.rows.View() },
		func(int) string { return m.load.View() },
		func(int) string { return m.autoRefresh.View() },
		func(int) string { return m.openLogs.View() },
	)
	operationsView := frame.Panel(m.theme, operationsContent, operations.Dx(), operationOptions)
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
	if m.showPalette {
		paletteWidth := m.commandsPanelArea.Dx()
		prompt := frame.Panel(m.theme, m.commands.View(), paletteWidth, frame.PanelOptions{
			Title:   "Command palette",
			Focused: true,
			Padding: 1,
		})
		base = overlay.Center(base, prompt)
	}
	if m.showConfirm {
		base = overlay.Center(base, m.confirm.View())
	}
	return base
}

// dispatchAction validates a tree-level semantic action against the current
// visible tree, then forwards only its component-local suffix. It deliberately
// returns the component command without executing it; Bubble Tea delivers the
// resulting message through the app's normal Update path.
func (m *model) dispatchAction(id string) (tea.Cmd, bool) {
	node, localID, ok := findEnabledAction(m.Inspect(), id)
	if !ok {
		return nil, false
	}

	switch node.ID {
	case "commands":
		next, cmd := m.commands.Update(inspect.Invoke(localID))
		m.commands = next
		return cmd, true
	case "confirm-restart":
		next, cmd := m.confirm.Update(inspect.Invoke(localID))
		m.confirm = next
		return cmd, true
	default:
		// The example intentionally exposes only the palette/dialog workflow as
		// a routed semantic surface. Other nodes remain inspectable, while
		// their application-specific side effects stay keyboard/app-owned.
		return nil, false
	}
}

// findEnabledAction also acts as the visibility check: hidden overlays are
// absent from Inspect, so their qualified IDs cannot be dispatched.
func findEnabledAction(node inspect.Node, id string) (inspect.Node, string, bool) {
	for _, action := range node.Actions {
		if action.ID != id || !action.Enabled {
			continue
		}
		prefix := node.ID + "."
		if node.ID == "" || !strings.HasPrefix(id, prefix) {
			return inspect.Node{}, "", false
		}
		return node, strings.TrimPrefix(id, prefix), true
	}
	for _, child := range node.Children {
		if target, localID, ok := findEnabledAction(child, id); ok {
			return target, localID, true
		}
	}
	return inspect.Node{}, "", false
}

func (m model) Inspect() inspect.Node {
	children := []inspect.Node{
		inspect.BindAt("tabs", inspect.FromRect(m.tabsArea), m.tabs),
		inspect.BindAt("actions", inspect.FromRect(m.actionsArea), m.actions),
		inspect.BindAt("operations", inspect.FromRect(m.rowsArea), m.rows),
		inspect.BindAt("load", inspect.FromRect(m.loadArea), m.load),
		inspect.BindAt("auto-refresh", inspect.FromRect(m.autoRefreshArea), m.autoRefresh),
		inspect.BindAt("open-logs", inspect.FromRect(m.openLogsArea), m.openLogs),
	}
	if m.showPalette {
		children = append(children, inspect.BindAt("commands", inspect.FromRect(m.commandsArea), m.commands))
	}
	if m.showConfirm {
		children = append(children, inspect.BindAt("confirm-restart", inspect.FromRect(m.confirmArea), m.confirm))
	}
	return inspect.Group("ops", "application", inspect.Bounds{Width: m.width, Height: m.height}, children...)
}

func (m model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
