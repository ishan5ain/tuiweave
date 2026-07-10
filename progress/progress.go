// Package progress renders a compact, themed progress indicator for a task or
// operation. It is passive: applications own the value and update it as work
// advances.
package progress

import (
	"fmt"
	"math"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
)

// Status selects the semantic role used for the filled portion of the bar.
type Status int

const (
	// StatusNormal uses the theme accent role.
	StatusNormal Status = iota
	// StatusSuccess indicates completed or healthy work.
	StatusSuccess
	// StatusWarning indicates work that needs attention.
	StatusWarning
	// StatusDanger indicates failed or blocked work.
	StatusDanger
	// StatusInfo indicates informational or background work.
	StatusInfo
)

// Model is a one-line passive progress component. Create one with New.
type Model struct {
	width, height int
	label         string
	percent       float64
	status        Status
	showPercent   bool

	labelStyle   lipgloss.Style
	trackStyle   lipgloss.Style
	fillStyles   map[Status]lipgloss.Style
	percentStyle lipgloss.Style
}

// New returns a progress indicator styled from the theme's semantic roles.
func New(theme gotui.Theme) Model {
	base := lipgloss.NewStyle().Background(theme.SurfaceRaised)
	return Model{
		showPercent: true,
		labelStyle:  base.Foreground(theme.TextMuted),
		trackStyle:  lipgloss.NewStyle().Foreground(theme.TextFaint).Background(theme.SurfaceSunken),
		fillStyles: map[Status]lipgloss.Style{
			StatusNormal:  lipgloss.NewStyle().Foreground(theme.TextInverted).Background(theme.Accent),
			StatusSuccess: lipgloss.NewStyle().Foreground(theme.TextInverted).Background(theme.Success),
			StatusWarning: lipgloss.NewStyle().Foreground(theme.TextInverted).Background(theme.Warning),
			StatusDanger:  lipgloss.NewStyle().Foreground(theme.TextInverted).Background(theme.Danger),
			StatusInfo:    lipgloss.NewStyle().Foreground(theme.TextInverted).Background(theme.Info),
		},
		percentStyle: base.Foreground(theme.TextMuted),
	}
}

// SetSize sets the box the indicator renders in. It renders one row whenever
// both dimensions are positive.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
}

// SetLabel sets the optional task label shown before the bar.
func (m *Model) SetLabel(label string) { m.label = label }

// Label returns the configured task label.
func (m Model) Label() string { return m.label }

// SetPercent sets the completion fraction. Values are clamped to [0, 1]; NaN
// is treated as zero so an unfinished calculation cannot corrupt rendering.
func (m *Model) SetPercent(percent float64) {
	if math.IsNaN(percent) || percent < 0 {
		percent = 0
	}
	if percent > 1 {
		percent = 1
	}
	m.percent = percent
}

// Percent returns the clamped completion fraction.
func (m Model) Percent() float64 { return m.percent }

// SetStatus changes the semantic status of the filled portion.
func (m *Model) SetStatus(status Status) {
	if _, ok := m.fillStyles[status]; !ok {
		status = StatusNormal
	}
	m.status = status
}

// Status returns the current semantic status.
func (m Model) Status() Status { return m.status }

// SetShowPercent controls whether the numeric percentage is shown.
func (m *Model) SetShowPercent(show bool) { m.showPercent = show }

// ShowPercent reports whether the numeric percentage is rendered.
func (m Model) ShowPercent() bool { return m.showPercent }

// Update implements the component contract. Progress handles no messages.
func (m Model) Update(_ tea.Msg) (Model, tea.Cmd) { return m, nil }

// View renders one exact-width row. The label and percentage give up space
// to the bar at narrow widths; the bar itself remains visible whenever the
// box has a positive width.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	suffix := ""
	if m.showPercent {
		suffix = fmt.Sprintf(" %3.0f%%", m.percent*100)
	}
	suffixWidth := lipgloss.Width(suffix)
	if suffixWidth+1 > m.width {
		suffix = ""
		suffixWidth = 0
	}

	label := m.label
	labelWidth := 0
	if label != "" {
		labelWidth = lipgloss.Width(label) + 1
		available := m.width - suffixWidth - 1
		if labelWidth > available {
			if available >= 2 {
				label = ansi.Truncate(label, available-1, "…")
				labelWidth = lipgloss.Width(label) + 1
			} else {
				label = ""
				labelWidth = 0
			}
		}
		if label != "" {
			label += " "
		} else {
			labelWidth = 0
		}
	}

	barWidth := m.width - labelWidth - suffixWidth
	if barWidth < 1 {
		label = ""
		labelWidth = 0
		barWidth = m.width - suffixWidth
		if barWidth < 1 {
			suffix = ""
			suffixWidth = 0
			barWidth = m.width
		}
	}

	filled := int(float64(barWidth) * m.percent)
	track := barWidth - filled
	fillStyle, ok := m.fillStyles[m.status]
	if !ok {
		fillStyle = m.fillStyles[StatusNormal]
	}
	return m.labelStyle.Render(label) +
		fillStyle.Render(strings.Repeat("█", filled)) +
		m.trackStyle.Render(strings.Repeat("░", track)) +
		m.percentStyle.Render(suffix)
}
