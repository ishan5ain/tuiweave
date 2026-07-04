// Package diffview renders unified diffs with theme-role line coloring —
// additions Success, deletions Danger, hunk headers Info.
//
// Two forms: Sprint styles a diff for inline use (chat cells), and Model is
// a sized, scrollable component for standalone diff panes.
package diffview

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/viewport"
)

type styles struct {
	file    lipgloss.Style
	hunk    lipgloss.Style
	add     lipgloss.Style
	del     lipgloss.Style
	context lipgloss.Style
	meta    lipgloss.Style
}

func newStyles(theme gotui.Theme) styles {
	return styles{
		file:    lipgloss.NewStyle().Foreground(theme.Text).Bold(true),
		hunk:    lipgloss.NewStyle().Foreground(theme.Info),
		add:     lipgloss.NewStyle().Foreground(theme.Success),
		del:     lipgloss.NewStyle().Foreground(theme.Danger),
		context: lipgloss.NewStyle().Foreground(theme.TextMuted),
		meta:    lipgloss.NewStyle().Foreground(theme.TextFaint),
	}
}

func (s styles) styleLine(line string) lipgloss.Style {
	switch {
	case strings.HasPrefix(line, "+++"), strings.HasPrefix(line, "---"):
		return s.file
	case strings.HasPrefix(line, "@@"):
		return s.hunk
	case strings.HasPrefix(line, "+"):
		return s.add
	case strings.HasPrefix(line, "-"):
		return s.del
	case strings.HasPrefix(line, "diff "), strings.HasPrefix(line, "index "):
		return s.meta
	default:
		return s.context
	}
}

func render(st styles, diff string, width int) string {
	lines := strings.Split(strings.TrimRight(diff, "\n"), "\n")
	out := make([]string, len(lines))
	for i, line := range lines {
		if width > 0 {
			line = ansi.Truncate(line, width, "…")
		}
		out[i] = st.styleLine(line).Render(line)
	}
	return strings.Join(out, "\n")
}

// Sprint styles a unified diff at the given width for inline rendering.
func Sprint(theme gotui.Theme, diff string, width int) string {
	return render(newStyles(theme), diff, width)
}

// Model is a scrollable diff pane. Create one with New; it follows the
// standard component contract and scrolls like a viewport.
type Model struct {
	vp   viewport.Model
	st   styles
	diff string
}

// New returns an empty diff pane styled from the theme's roles.
func New(theme gotui.Theme) Model {
	return Model{vp: viewport.New(theme), st: newStyles(theme)}
}

// SetDiff replaces the displayed diff.
func (m *Model) SetDiff(diff string) {
	m.diff = diff
	m.vp.SetContent(render(m.st, diff, 0)) // viewport clips lines to width
}

// SetSize sets the box the pane renders in.
func (m *Model) SetSize(width, height int) { m.vp.SetSize(width, height) }

// Focus makes the pane respond to scroll keys.
func (m *Model) Focus() { m.vp.Focus() }

// Blur stops the pane from responding to keys.
func (m *Model) Blur() { m.vp.Blur() }

// Focused reports whether the pane handles keys.
func (m Model) Focused() bool { return m.vp.Focused() }

// Update delegates scrolling to the embedded viewport.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

// View renders the visible window of the diff.
func (m Model) View() string { return m.vp.View() }
