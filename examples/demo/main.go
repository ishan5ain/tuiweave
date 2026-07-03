// Command demo is the Phase 2 exit-criterion app: a multi-pane TUI composed
// purely from gotui components, laid out with gotui/layout, with focus
// cycling and a modal dialog composited by gotui/overlay.
//
//	go run ./examples/demo
//
// tab cycles focus, enter (in the input) adds an item, ctrl+d opens the quit
// dialog, ctrl+c quits immediately.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/dialog"
	"github.com/ishansain/gotui/focus"
	"github.com/ishansain/gotui/help"
	"github.com/ishansain/gotui/layout"
	"github.com/ishansain/gotui/list"
	"github.com/ishansain/gotui/overlay"
	"github.com/ishansain/gotui/spinner"
	"github.com/ishansain/gotui/statusbar"
	"github.com/ishansain/gotui/textinput"
	"github.com/ishansain/gotui/viewport"
)

type model struct {
	theme         gotui.Theme
	width, height int

	list    list.Model
	view    viewport.Model
	input   textinput.Model
	helpbar help.Model
	status  statusbar.Model
	spin    spinner.Model
	quitDlg dialog.Model

	fm         focus.Manager
	showDialog bool
}

func newModel() model {
	theme := gotui.Dark()
	m := model{
		theme:   theme,
		list:    list.New(theme),
		view:    viewport.New(theme),
		input:   textinput.New(theme),
		helpbar: help.New(theme),
		status:  statusbar.New(theme),
		spin:    spinner.New(theme),
		quitDlg: dialog.New(theme),
	}
	m.list.SetItems("welcome", "layout", "theming", "testing")
	m.input.Placeholder = "add an item, then press enter"
	m.helpbar.SetBindings(
		help.Binding{Key: "tab", Desc: "focus"},
		help.Binding{Key: "enter", Desc: "add item"},
		help.Binding{Key: "j/k", Desc: "navigate"},
		help.Binding{Key: "ctrl+d", Desc: "quit dialog"},
		help.Binding{Key: "ctrl+c", Desc: "quit"},
	)
	m.quitDlg.ID = "quit"
	m.quitDlg.Title = "Quit demo?"
	m.quitDlg.Body = "This closes the gotui Phase 2 demo."
	m.fm = focus.NewManager(3)
	m.applyFocus()
	m.syncViewport()
	return m
}

// applyFocus pushes the manager's index into the components. Called with the
// components' current addresses after every focus change — the manager holds
// no pointers, so it survives the model being copied between Updates.
func (m *model) applyFocus() {
	m.fm.Apply(&m.list, &m.view, &m.input)
}

func (m *model) syncViewport() {
	sel := m.list.SelectedItem()
	var b strings.Builder
	fmt.Fprintf(&b, "── %s ──\n\n", sel)
	for i := 1; i <= 40; i++ {
		fmt.Fprintf(&b, "%s: content line %d\n", sel, i)
	}
	m.view.SetContent(b.String())
}

func (m *model) syncStatus() {
	focusName := [...]string{"list", "viewport", "input"}[m.fm.Index()]
	m.status.SetLeft(
		statusbar.Segment{Text: "demo", Kind: statusbar.KindAccent},
		statusbar.Segment{Text: focusName, Kind: statusbar.KindNormal},
	)
	m.status.SetRight(
		statusbar.Segment{Text: m.spin.View() + " live", Kind: statusbar.KindInfo},
		statusbar.Segment{Text: fmt.Sprintf("%d items", m.list.Len()), Kind: statusbar.KindMuted},
		statusbar.Segment{Text: fmt.Sprintf("%3.0f%%", m.view.ScrollPercent()*100), Kind: statusbar.KindMuted},
	)
}

func (m model) Init() tea.Cmd { return m.spin.Tick() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()

	case spinner.TickMsg:
		m.spin, cmd = m.spin.Update(msg)
		cmds = append(cmds, cmd)

	case dialog.ResultMsg:
		m.showDialog = false
		if msg.OK {
			return m, tea.Quit
		}

	case tea.KeyPressMsg:
		if m.showDialog {
			m.quitDlg, cmd = m.quitDlg.Update(msg)
			m.syncStatus()
			return m, cmd
		}
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+d":
			m.showDialog = true
			m.syncStatus()
			return m, nil
		case "tab":
			m.fm.Next()
			m.applyFocus()
		case "shift+tab":
			m.fm.Prev()
			m.applyFocus()
		case "enter":
			if m.fm.Index() == 2 && strings.TrimSpace(m.input.Value()) != "" {
				m.list.SetItems(append(m.list.Items(), strings.TrimSpace(m.input.Value()))...)
				m.list.Select(m.list.Len() - 1)
				m.input.Reset()
				m.syncViewport()
			}
		default:
			before := m.list.Selected()
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
			m.view, cmd = m.view.Update(msg)
			cmds = append(cmds, cmd)
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
			if m.list.Selected() != before {
				m.syncViewport()
			}
		}

	case tea.MouseWheelMsg:
		m.view, cmd = m.view.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.syncStatus()
	return m, tea.Batch(cmds...)
}

func (m *model) layout() {
	var main, input, helpRow, status layout.Rect
	layout.Vertical(
		layout.Fill(1), // panes
		layout.Len(1),  // input
		layout.Len(1),  // help
		layout.Len(1),  // statusbar
	).Split(layout.NewRect(0, 0, m.width, m.height)).
		Assign(&main, &input, &helpRow, &status)

	layout.Horizontal(
		layout.Percent(30),
		layout.Fill(1),
	).WithSpacing(1).
		Apply(main, &m.list, &m.view)

	m.input.SetSize(input.Dx(), input.Dy())
	m.helpbar.SetSize(helpRow.Dx(), helpRow.Dy())
	m.status.SetSize(status.Dx(), status.Dy())
	m.quitDlg.SetSize(min(40, m.width-4), 10)
}

// render composes the full frame; separate from View so tests can golden it.
func (m model) render() string {
	panes := lipgloss.JoinHorizontal(lipgloss.Top, m.list.View(), " ", m.view.View())
	base := lipgloss.JoinVertical(lipgloss.Left,
		panes,
		m.input.View(),
		m.helpbar.View(),
		m.status.View(),
	)
	if m.showDialog {
		base = overlay.Center(base, m.quitDlg.View())
	}
	return base
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
