// Package viewport provides a scrollable window over pre-rendered content.
//
// The viewport draws no chrome of its own — wrap it in a bordered panel or
// dialog if you need a frame. Content may contain ANSI styling; lines are
// clipped to the viewport width ANSI-aware.
package viewport

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
)

// Model is a viewport component. Create one with New.
type Model struct {
	width, height int
	lines         []string
	yoff          int
	focused       bool
}

// New returns an empty viewport. The theme parameter is part of the
// component contract (future chrome like scrollbars will use it).
func New(_ gotui.Theme) Model {
	return Model{}
}

// SetSize sets the box the viewport renders in.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
	m.clamp()
}

// SetContent replaces the viewport content, preserving the scroll position
// when possible.
func (m *Model) SetContent(s string) {
	m.lines = strings.Split(s, "\n")
	m.clamp()
}

// Focus makes the viewport respond to scroll keys.
func (m *Model) Focus() { m.focused = true }

// Blur stops the viewport from responding to keys.
func (m *Model) Blur() { m.focused = false }

// Focused reports whether the viewport handles keys.
func (m Model) Focused() bool { return m.focused }

// YOffset returns the index of the first visible line.
func (m Model) YOffset() int { return m.yoff }

// AtBottom reports whether the last content line is visible.
func (m Model) AtBottom() bool { return m.yoff >= m.maxYOffset() }

// ScrollPercent returns the scroll position in [0, 1].
func (m Model) ScrollPercent() float64 {
	if m.maxYOffset() == 0 {
		return 1
	}
	return float64(m.yoff) / float64(m.maxYOffset())
}

// ScrollTo scrolls so line y is the first visible line, clamped to bounds.
func (m *Model) ScrollTo(y int) {
	m.yoff = y
	m.clamp()
}

// ScrollBy scrolls by delta lines (negative is up), clamped to bounds.
func (m *Model) ScrollBy(delta int) { m.ScrollTo(m.yoff + delta) }

// GotoTop scrolls to the first line.
func (m *Model) GotoTop() { m.ScrollTo(0) }

// GotoBottom scrolls so the last line is visible.
func (m *Model) GotoBottom() { m.ScrollTo(m.maxYOffset()) }

func (m Model) maxYOffset() int {
	return max(0, len(m.lines)-m.height)
}

func (m *Model) clamp() {
	m.yoff = max(0, min(m.yoff, m.maxYOffset()))
}

// Update handles scroll keys (when focused) and mouse wheel events.
//
//	up/k  down/j     line
//	pgup/b  pgdown/f line page
//	u / d            half page
//	g / home         top
//	G / end          bottom
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseWheelMsg:
		switch msg.Button {
		case tea.MouseWheelUp:
			m.ScrollBy(-3)
		case tea.MouseWheelDown:
			m.ScrollBy(3)
		}
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		switch msg.String() {
		case "up", "k":
			m.ScrollBy(-1)
		case "down", "j":
			m.ScrollBy(1)
		case "pgup", "b":
			m.ScrollBy(-m.height)
		case "pgdown", "f":
			m.ScrollBy(m.height)
		case "u":
			m.ScrollBy(-m.height / 2)
		case "d":
			m.ScrollBy(m.height / 2)
		case "g", "home":
			m.GotoTop()
		case "G", "end":
			m.GotoBottom()
		}
	}
	return m, nil
}

// View renders the visible window, each line clipped and padded to the
// viewport width so the box is always fully painted.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	rows := make([]string, m.height)
	for i := range rows {
		y := m.yoff + i
		if y >= len(m.lines) {
			rows[i] = strings.Repeat(" ", m.width)
			continue
		}
		line := ansi.Truncate(m.lines[y], m.width, "")
		if pad := m.width - ansi.StringWidth(line); pad > 0 {
			line += strings.Repeat(" ", pad)
		}
		rows[i] = line
	}
	return strings.Join(rows, "\n")
}
