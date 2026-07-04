// Command table is a mock git-status app: a table of changed files beside a
// scrollable diff of the selected row, with scrollbars on both panes. It
// exercises the two components no other example does — table and
// diffview.Model — plus gotui/scrollbar.
//
//	go run ./examples/table
//
// tab toggles focus, j/k navigate, ctrl+c quits.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/agentic/diffview"
	"github.com/ishansain/gotui/focus"
	"github.com/ishansain/gotui/layout"
	"github.com/ishansain/gotui/scrollbar"
	"github.com/ishansain/gotui/statusbar"
	"github.com/ishansain/gotui/table"
)

type change struct {
	file, status string
	adds, dels   int
}

var changes = func() []change {
	out := make([]change, 0, 24)
	for i := 1; i <= 24; i++ {
		status := "modified"
		if i%5 == 0 {
			status = "added"
		}
		out = append(out, change{
			file:   fmt.Sprintf("internal/pkg%02d/handler.go", i),
			status: status,
			adds:   i * 3,
			dels:   i,
		})
	}
	return out
}()

func diffFor(c change) string {
	var b strings.Builder
	fmt.Fprintf(&b, "--- a/%s\n+++ b/%s\n@@ -1,%d +1,%d @@\n", c.file, c.file, c.dels+2, c.adds+2)
	b.WriteString(" func Handle(w http.ResponseWriter, r *http.Request) {\n")
	for i := range c.dels {
		fmt.Fprintf(&b, "-\tlog.Printf(\"old %d\")\n", i+1)
	}
	for i := range c.adds {
		fmt.Fprintf(&b, "+\tslog.Info(\"new\", \"n\", %d)\n", i+1)
	}
	b.WriteString(" }")
	return b.String()
}

type model struct {
	theme         gotui.Theme
	width, height int

	files  table.Model
	diff   diffview.Model
	status statusbar.Model
	fm     focus.Manager
}

func newModel() model {
	theme := gotui.Dark()
	m := model{
		theme:  theme,
		files:  table.New(theme),
		diff:   diffview.New(theme),
		status: statusbar.New(theme),
		fm:     focus.NewManager(2),
	}
	m.files.SetColumns(
		table.Column{Title: "File"}, // flex
		table.Column{Title: "Status", Width: 8},
		table.Column{Title: "+/-", Width: 9},
	)
	rows := make([][]string, len(changes))
	for i, c := range changes {
		rows[i] = []string{c.file, c.status, fmt.Sprintf("+%d -%d", c.adds, c.dels)}
	}
	m.files.SetRows(rows...)
	m.applyFocus()
	m.syncDiff()
	return m
}

func (m *model) applyFocus() { m.fm.Apply(&m.files, &m.diff) }

func (m *model) syncDiff() {
	if sel := m.files.Selected(); sel >= 0 {
		m.diff.SetDiff(diffFor(changes[sel]))
	}
	m.syncStatus()
}

func (m *model) syncStatus() {
	pane := [...]string{"files", "diff"}[m.fm.Index()]
	m.status.SetLeft(
		statusbar.Segment{Text: "git-status", Kind: statusbar.KindAccent},
		statusbar.Segment{Text: pane, Kind: statusbar.KindNormal},
	)
	m.status.SetRight(
		statusbar.Segment{Text: fmt.Sprintf("%d files", len(changes)), Kind: statusbar.KindMuted},
	)
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		var main, status layout.Rect
		layout.Vertical(layout.Fill(1), layout.Len(1)).
			Split(layout.NewRect(0, 0, msg.Width, msg.Height)).
			Assign(&main, &status)

		var filesPane, filesBar, diffPane, diffBar layout.Rect
		layout.Horizontal(
			layout.Percent(45),
			layout.Len(1),
			layout.Fill(1),
			layout.Len(1),
		).WithSpacing(1).
			Split(main).
			Assign(&filesPane, &filesBar, &diffPane, &diffBar)

		m.files.SetSize(filesPane.Dx(), filesPane.Dy())
		m.diff.SetSize(diffPane.Dx(), diffPane.Dy())
		m.status.SetSize(status.Dx(), status.Dy())

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.fm.Next()
			m.applyFocus()
			m.syncStatus()
		default:
			before := m.files.Selected()
			m.files, cmd = m.files.Update(msg)
			cmds = append(cmds, cmd)
			m.diff, cmd = m.diff.Update(msg)
			cmds = append(cmds, cmd)
			if m.files.Selected() != before {
				m.syncDiff()
			}
		}

	case tea.MouseWheelMsg:
		m.diff, cmd = m.diff.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// render composes the full frame; separate from View so tests can golden it.
func (m model) render() string {
	// The table's scrollbar spans its row area, which starts two lines down
	// (header + rule); pad the bar to align.
	filesBar := " \n \n" + scrollbar.For(m.theme, m.files)
	diffBar := scrollbar.For(m.theme, m.diff)

	main := lipgloss.JoinHorizontal(lipgloss.Top,
		m.files.View(), filesBar, " ", m.diff.View(), diffBar,
	)
	return lipgloss.JoinVertical(lipgloss.Left, main, m.status.View())
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
