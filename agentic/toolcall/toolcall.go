// Package toolcall renders a tool invocation as a transcript block: status
// icon, tool name, argument summary, and collapsible output.
//
// Block is a chat cell, not an MVU component: create it with New, keep the
// pointer, and mutate it as the tool progresses (SetStatus, AppendOutput,
// Expanded). The chat transcript re-renders cells every frame, so mutations
// show up immediately.
package toolcall

import (
	"fmt"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
)

// Status is the lifecycle state of a tool call.
type Status int

const (
	StatusPending Status = iota
	StatusRunning
	StatusSuccess
	StatusError
)

// Block is one tool call in a transcript. Mutate it via the exported fields
// and methods; it renders on demand.
type Block struct {
	// Name is the tool name, e.g. "Bash".
	Name string
	// Summary is a one-line description of the call, e.g. the command.
	Summary string
	// Expanded shows the output block when true; collapsed shows a line-count
	// hint instead.
	Expanded bool
	// MaxOutputLines caps the expanded output; the rest collapses into a
	// "+N more" line. Default 8.
	MaxOutputLines int

	status Status
	output []string

	iconStyles  map[Status]lipgloss.Style
	nameStyle   lipgloss.Style
	summarySt   lipgloss.Style
	outputStyle lipgloss.Style
	hintStyle   lipgloss.Style
}

// New returns a pending tool-call block styled from the theme's roles.
func New(theme gotui.Theme, name, summary string) *Block {
	return &Block{
		Name:           name,
		Summary:        summary,
		MaxOutputLines: 8,
		iconStyles: map[Status]lipgloss.Style{
			StatusPending: lipgloss.NewStyle().Foreground(theme.TextFaint),
			StatusRunning: lipgloss.NewStyle().Foreground(theme.Warning),
			StatusSuccess: lipgloss.NewStyle().Foreground(theme.Success),
			StatusError:   lipgloss.NewStyle().Foreground(theme.Danger),
		},
		nameStyle:   lipgloss.NewStyle().Foreground(theme.Text).Bold(true),
		summarySt:   lipgloss.NewStyle().Foreground(theme.TextMuted),
		outputStyle: lipgloss.NewStyle().Foreground(theme.TextMuted).Background(theme.SurfaceSunken),
		hintStyle:   lipgloss.NewStyle().Foreground(theme.TextFaint),
	}
}

// SetStatus updates the lifecycle state.
func (b *Block) SetStatus(s Status) { b.status = s }

// Status returns the lifecycle state.
func (b *Block) Status() Status { return b.status }

// AppendOutput adds output text (split on newlines) to the block.
func (b *Block) AppendOutput(s string) {
	b.output = append(b.output, strings.Split(strings.TrimRight(s, "\n"), "\n")...)
}

var icons = map[Status]string{
	StatusPending: "○",
	StatusRunning: "◐",
	StatusSuccess: "✓",
	StatusError:   "✗",
}

// Render draws the block at the given width (chat cell contract).
func (b *Block) Render(width int) string {
	if width <= 0 {
		return ""
	}
	header := b.iconStyles[b.status].Render(icons[b.status]) + " " +
		b.nameStyle.Render(b.Name)
	if b.Summary != "" {
		header += " " + b.summarySt.Render(
			ansi.Truncate(b.Summary, max(0, width-ansi.StringWidth(b.Name)-2), "…"))
	}

	if len(b.output) == 0 {
		return header
	}
	if !b.Expanded {
		return header + " " + b.hintStyle.Render(fmt.Sprintf("(%d output lines)", len(b.output)))
	}

	shown := b.output
	hidden := 0
	if b.MaxOutputLines > 0 && len(shown) > b.MaxOutputLines {
		hidden = len(shown) - b.MaxOutputLines
		shown = shown[:b.MaxOutputLines]
	}
	rows := make([]string, 0, len(shown)+2)
	rows = append(rows, header)
	for _, line := range shown {
		line = ansi.Truncate(line, width-4, "…")
		pad := strings.Repeat(" ", max(0, width-4-ansi.StringWidth(line)))
		rows = append(rows, "  "+b.outputStyle.Render(" "+line+pad+" "))
	}
	if hidden > 0 {
		rows = append(rows, "  "+b.hintStyle.Render(fmt.Sprintf("… +%d more lines", hidden)))
	}
	return strings.Join(rows, "\n")
}
