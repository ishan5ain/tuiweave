// Command browser is a mock file-browser reference app. It combines a
// filterable file list with a scrollable preview, tabs, framed panes, and
// semantic inspection without reading the host file system.
//
//	go run ./examples/browser
//
// tab/shift+tab cycles focus, / focuses the filter, and q quits.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/focus"
	"github.com/ishan5ain/tuiweave/frame"
	"github.com/ishan5ain/tuiweave/inspect"
	"github.com/ishan5ain/tuiweave/layout"
	"github.com/ishan5ain/tuiweave/line"
	"github.com/ishan5ain/tuiweave/list"
	"github.com/ishan5ain/tuiweave/scrollbar"
	"github.com/ishan5ain/tuiweave/stack"
	"github.com/ishan5ain/tuiweave/statusbar"
	"github.com/ishan5ain/tuiweave/tabs"
	"github.com/ishan5ain/tuiweave/textinput"
	"github.com/ishan5ain/tuiweave/viewport"
)

type fileEntry struct {
	path    string
	content string
}

func mainPreview() string {
	return strings.Join([]string{
		"package main",
		"",
		"import (",
		"\t\"fmt\"",
		"\t\"os\"",
		")",
		"",
		"func main() {",
		"\tconfig := loadConfig()",
		"\tif err := run(config); err != nil {",
		"\t\tfmt.Fprintln(os.Stderr, err)",
		"\t\tos.Exit(1)",
		"\t}",
		"}",
		"",
		"func loadConfig() Config {",
		"\treturn Config{Mode: \"development\"}",
		"}",
		"",
		"func run(config Config) error {",
		"\tfmt.Println(config.Mode)",
		"\treturn nil",
		"}",
		"",
		"type Config struct {",
		"\tMode string",
		"}",
	}, "\n")
}

func entriesFor(tabID string) []fileEntry {
	switch tabID {
	case "recent":
		return []fileEntry{
			{path: "README.md", content: "# tuiweave\n\nA small, agent-friendly TUI toolkit.\n\nStart with the layout and component contracts."},
			{path: "AGENTS.md", content: "# Agent Conventions\n\nUse theme roles, SetSize, MVU updates, and snapshot tests.\n\nCompose from the existing vocabulary first."},
			{path: "examples/ops/main.go", content: "// Command ops is a reference operations console.\n\nIt composes tabs, actions, tables, controls, and a command palette."},
			{path: "action/action.go", content: "package action\n\n// Item is one selectable application action.\ntype Item struct {\n\tID string\n\tLabel string\n}"},
		}
	case "pinned":
		return []fileEntry{
			{path: "DESIGN.md", content: "# Design\n\nThe core remains domain-neutral.\n\nComposition stays explicit so applications own routing and state."},
			{path: "ROADMAP.md", content: "# Roadmap\n\nHarden the public component vocabulary through real application use."},
			{path: "go.mod", content: "module github.com/ishan5ain/tuiweave\n\ngo 1.23\n\nrequire charm.land/bubbletea/v2"},
		}
	default:
		return []fileEntry{
			{path: "cmd/tuiweave/main.go", content: mainPreview()},
			{path: "layout/layout.go", content: "package layout\n\n// Rect is a terminal-space rectangle.\ntype Rect struct {\n\tX, Y, W, H int\n}"},
			{path: "list/list.go", content: "package list\n\n// Model is a filterable, focusable list.\ntype Model struct {\n\titems []string\n}"},
			{path: "examples/browser/main.go", content: "// Command browser is a mock file-browser reference app.\n\n// It exercises filtering, preview scrolling, focus, and inspection."},
			{path: "examples/frame/main.go", content: "// Command frame demonstrates composition primitives.\n\n// Panels, tabs, menus, toolbars, and controls share explicit layout."},
			{path: "agentic/chat/chat.go", content: "package chat\n\n// Model renders a streaming transcript.\n// Cells remain application-owned pointers."},
			{path: "agentic/toolcall/toolcall.go", content: "package toolcall\n\n// Block renders tool status and output.\n// The application owns execution."},
		}
	}
}

type model struct {
	theme         tuiweave.Theme
	width, height int

	tabs    tabs.Model
	filter  textinput.Model
	files   list.Model
	preview viewport.Model
	status  statusbar.Model

	entries []fileEntry
	fm      focus.Manager
}

func newModel() model {
	theme := tuiweave.Dark()
	m := model{
		theme:   theme,
		tabs:    tabs.New(theme),
		filter:  textinput.New(theme),
		files:   list.New(theme),
		preview: viewport.New(theme),
		status:  statusbar.New(theme),
		fm:      focus.NewManager(4),
	}
	m.tabs.SetTabs(
		tabs.Tab{ID: "workspace", Label: "Workspace"},
		tabs.Tab{ID: "recent", Label: "Recent"},
		tabs.Tab{ID: "pinned", Label: "Pinned"},
	)
	m.filter.Placeholder = "filter files"
	m.loadEntries()
	m.applyFocus()
	m.syncStatus()
	return m
}

func (m *model) applyFocus() {
	m.fm.Apply(&m.tabs, &m.filter, &m.files, &m.preview)
}

func (m *model) loadEntries() {
	m.entries = entriesFor(m.tabs.SelectedID())
	paths := make([]string, len(m.entries))
	for i, entry := range m.entries {
		paths[i] = entry.path
	}
	m.files.SetItems(paths...)
	m.files.SetFilter(m.filter.Value())
	m.syncPreview()
}

func (m *model) syncPreview() {
	selected := m.files.Selected()
	if selected < 0 || selected >= len(m.entries) {
		m.preview.SetContent("No file selected\n\nChoose a file from the list.")
	} else {
		m.preview.SetContent(m.entries[selected].content)
	}
	m.preview.GotoTop()
}

func (m *model) syncStatus() {
	focusName := [...]string{"tabs", "filter", "files", "preview"}[m.fm.Index()]
	selected := "no file"
	if index := m.files.Selected(); index >= 0 && index < len(m.entries) {
		selected = m.entries[index].path
	}
	m.status.SetLeft(
		statusbar.Segment{Text: "browser", Kind: statusbar.KindAccent},
		statusbar.Segment{Text: focusName, Kind: statusbar.KindNormal},
	)
	m.status.SetRight(
		statusbar.Segment{Text: selected, Kind: statusbar.KindMuted},
		statusbar.Segment{Text: fmt.Sprintf("%d files", m.files.FilteredLen()), Kind: statusbar.KindInfo},
	)
}

func (m model) Init() tea.Cmd { return nil }

func (m *model) delegate(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.tabs, cmd = m.tabs.Update(msg)
	cmds = append(cmds, cmd)
	m.filter, cmd = m.filter.Update(msg)
	cmds = append(cmds, cmd)
	m.files, cmd = m.files.Update(msg)
	cmds = append(cmds, cmd)
	m.preview, cmd = m.preview.Update(msg)
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "/":
			m.fm.Set(1)
			m.applyFocus()
		default:
			if msg.String() == "tab" {
				m.fm.Next()
				m.applyFocus()
			} else if msg.String() == "shift+tab" {
				m.fm.Prev()
				m.applyFocus()
			} else {
				beforeTab := m.tabs.SelectedID()
				beforeQuery := m.filter.Value()
				beforeFile := m.files.Selected()
				cmds = append(cmds, m.delegate(msg))
				if beforeTab != m.tabs.SelectedID() {
					m.loadEntries()
				} else {
					if beforeQuery != m.filter.Value() {
						m.files.SetFilter(m.filter.Value())
					}
					if beforeFile != m.files.Selected() || beforeQuery != m.filter.Value() {
						m.syncPreview()
					}
				}
			}
		}

	default:
		beforeFile := m.files.Selected()
		cmds = append(cmds, m.delegate(msg))
		if beforeFile != m.files.Selected() {
			m.syncPreview()
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

	var left, right layout.Rect
	layout.Horizontal(
		layout.Percent(36),
		layout.Fill(1),
	).WithSpacing(1).Split(body).Assign(&left, &right)

	leftOptions := frame.PanelOptions{Title: "Files", Focused: m.filter.Focused() || m.files.Focused(), Padding: 1}
	leftContent := frame.PanelContentRect(left, leftOptions)
	var filter, files layout.Rect
	layout.Vertical(layout.Len(1), layout.Fill(1)).Split(leftContent).Assign(&filter, &files)
	m.filter.SetSize(filter.Dx(), filter.Dy())
	m.files.SetSize(files.Dx(), files.Dy())

	rightOptions := frame.PanelOptions{Title: "Preview", Focused: m.preview.Focused(), Padding: 1}
	var previewPanel, previewBar layout.Rect
	layout.Horizontal(layout.Fill(1), layout.Len(1)).Split(right).Assign(&previewPanel, &previewBar)
	rightContent := frame.PanelContentRect(previewPanel, rightOptions)
	m.preview.SetSize(rightContent.Dx(), rightContent.Dy())
	m.status.SetSize(footer.Dx(), footer.Dy())
}

func (m model) render() string {
	if m.width < 24 || m.height < 10 {
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
			frame.Badge(m.theme, "BROWSER", frame.BadgeAccent),
			" ",
			lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render("mock workspace"),
		),
		"/ filter",
		line.JoinOptions{Gap: 1})
	headerView := stack.Vertical(m.theme, header.Dx(), stack.Options{},
		func(int) string { return headerStyle.Render(headerLine) },
		func(int) string { return m.tabs.View() },
	)

	var left, right layout.Rect
	layout.Horizontal(layout.Percent(36), layout.Fill(1)).WithSpacing(1).
		Split(body).Assign(&left, &right)
	leftOptions := frame.PanelOptions{Title: "Files", Focused: m.filter.Focused() || m.files.Focused(), Padding: 1}
	rightOptions := frame.PanelOptions{Title: "Preview", Focused: m.preview.Focused(), Padding: 1}
	leftContent := frame.PanelContentRect(left, leftOptions)
	leftInner := stack.Vertical(m.theme, leftContent.Dx(), stack.Options{},
		func(int) string { return m.filter.View() },
		func(int) string { return m.files.View() },
	)
	leftView := frame.Panel(m.theme, leftInner, left.Dx(), leftOptions)
	var previewPanel, previewBar layout.Rect
	layout.Horizontal(layout.Fill(1), layout.Len(1)).Split(right).Assign(&previewPanel, &previewBar)
	rightView := frame.Panel(m.theme, m.preview.View(), previewPanel.Dx(), rightOptions)
	previewScroll := strings.Join([]string{
		" ",
		" ",
		scrollbar.For(m.theme, m.preview),
		" ",
		" ",
	}, "\n")
	rightView = lipgloss.JoinHorizontal(lipgloss.Top, rightView, previewScroll)
	bodyView := lipgloss.JoinHorizontal(lipgloss.Top, leftView, " ", rightView)

	footerView := lipgloss.NewStyle().
		Width(footer.Dx()).
		Foreground(m.theme.TextMuted).
		Render(line.Join(footer.Dx(), frame.Divider(m.theme, footer.Dx()), "tab focus · / filter · q quit", line.JoinOptions{Gap: 1, NoEllipsis: true}))

	return stack.Vertical(m.theme, m.width, stack.Options{},
		func(int) string { return headerView },
		func(int) string { return bodyView },
		func(int) string { return footerView },
	)
}

func (m model) Inspect() inspect.Node {
	return inspect.Group("browser", "application", inspect.Bounds{Width: m.width, Height: m.height},
		inspect.Bind("tabs", m.tabs),
		inspect.Bind("filter", m.filter),
		inspect.Bind("files", m.files),
		inspect.Bind("preview", m.preview),
	)
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
