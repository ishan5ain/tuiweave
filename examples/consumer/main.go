// Command consumer is a small public-API consumer smoke app. It intentionally
// keeps orchestration in the application while composing layout, frame, list,
// statusbar, inspection, semantic-action routing, and theme contracts from
// tuiweave.
//
//	go run ./examples/consumer
//
// Use j/k or the arrow keys to navigate and q to quit.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/frame"
	"github.com/ishan5ain/tuiweave/inspect"
	"github.com/ishan5ain/tuiweave/layout"
	"github.com/ishan5ain/tuiweave/list"
	"github.com/ishan5ain/tuiweave/statusbar"
)

type model struct {
	theme         tuiweave.Theme
	width, height int
	body          layout.Rect
	content       layout.Rect
	status        layout.Rect
	items         list.Model
	bar           statusbar.Model
}

func newModel() model {
	theme := tuiweave.Dark()
	items := list.New(theme)
	items.SetItems("layout", "focus", "frame", "snaptest")
	items.Focus()

	return model{
		theme: theme,
		items: items,
		bar:   statusbar.New(theme),
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()
	case inspect.ActionMsg:
		cmd = m.dispatchAction(msg.ID)
	case tea.KeyPressMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		m.items, cmd = m.items.Update(msg)
	}

	m.syncStatus()
	return m, cmd
}

// dispatchAction validates the current qualified action before forwarding its
// local suffix. Even this single-component app keeps identity, visibility,
// enabled-state validation, and routing at the application boundary.
func (m *model) dispatchAction(id string) tea.Cmd {
	node := inspect.BindAt("items", inspect.FromRect(m.content), m.items)
	prefix := node.ID + "."
	if !strings.HasPrefix(id, prefix) {
		return nil
	}
	for _, action := range node.Actions {
		if action.ID != id || !action.Enabled {
			continue
		}
		next, cmd := m.items.Update(inspect.Invoke(strings.TrimPrefix(id, prefix)))
		m.items = next
		return cmd
	}
	return nil
}

func (m *model) layout() {
	layout.Vertical(layout.Fill(1), layout.Len(1)).
		Split(layout.NewRect(0, 0, m.width, m.height)).
		Assign(&m.body, &m.status)
	m.content = frame.PanelContentRect(m.body, frame.PanelOptions{Title: "Components"})
	m.items.SetSize(m.content.Dx(), m.content.Dy())
	m.bar.SetSize(m.status.Dx(), m.status.Dy())
}

func (m *model) syncStatus() {
	m.bar.SetLeft(
		statusbar.Segment{Text: "consumer", Kind: statusbar.KindAccent},
		statusbar.Segment{Text: m.items.SelectedItem(), Kind: statusbar.KindNormal},
	)
	m.bar.SetRight(statusbar.Segment{Text: "j/k navigate · q quit", Kind: statusbar.KindMuted})
}

func (m model) render() string {
	panel := frame.Panel(m.theme, m.items.View(), m.body.Dx(), frame.PanelOptions{Title: "Components"})
	return lipgloss.JoinVertical(lipgloss.Left, panel, m.bar.View())
}

// Inspect assembles application-owned identity and absolute bounds around the
// component's local semantic report.
func (m model) Inspect() inspect.Node {
	return inspect.Group("consumer", "application", inspect.Bounds{Width: m.width, Height: m.height},
		inspect.BindAt("items", inspect.FromRect(m.content), m.items),
	)
}

func (m model) View() tea.View { return tea.NewView(m.render()) }

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
