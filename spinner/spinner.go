// Package spinner provides an animated activity indicator.
//
// The spinner is an intrinsic-size component: it renders a single glyph at
// its natural width and ignores SetSize (which exists to satisfy the
// component contract). Start the animation by returning Tick from Init or
// when work begins, and forward TickMsg to Update.
package spinner

import (
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/layout"
)

var lastID atomic.Int64

// TickMsg advances a spinner's animation. It is produced by Tick and routed
// back to the spinner that requested it.
type TickMsg struct {
	Time time.Time
	id   int64
	tag  int
}

// Model is a spinner component. Create one with New.
type Model struct {
	frames []string
	fps    time.Duration
	frame  int
	id     int64
	tag    int
	style  lipgloss.Style
}

// New returns a spinner styled with the theme's Accent role, animating the
// classic braille-dots frames at 12 FPS.
func New(theme gotui.Theme) Model {
	return Model{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		fps:    time.Second / 12,
		id:     lastID.Add(1),
		style:  lipgloss.NewStyle().Foreground(theme.Accent),
	}
}

// SetFrames replaces the animation frames.
func (m *Model) SetFrames(frames ...string) {
	if len(frames) > 0 {
		m.frames = frames
		m.frame = 0
	}
}

// SetSize implements the component contract. The spinner renders at its
// intrinsic size and ignores the box.
func (m *Model) SetSize(width, height int) {}

// SizeMode reports that the spinner renders at its natural glyph size.
func (m Model) SizeMode() layout.SizeMode { return layout.SizeIntrinsic }

// Tick starts (or continues) the animation. Return it from Init or when the
// spinner becomes visible.
func (m Model) Tick() tea.Cmd {
	return tea.Tick(m.fps, func(t time.Time) tea.Msg {
		return TickMsg{Time: t, id: m.id, tag: m.tag}
	})
}

// Update advances the animation on TickMsg and schedules the next tick.
// Messages for other spinners (or stale ticks) are ignored.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	tick, ok := msg.(TickMsg)
	if !ok || tick.id != m.id || tick.tag != m.tag {
		return m, nil
	}
	m.frame = (m.frame + 1) % len(m.frames)
	m.tag++
	return m, m.Tick()
}

// View renders the current frame.
func (m Model) View() string {
	return m.style.Render(m.frames[m.frame])
}
