// Package toolcall renders a tool invocation as a transcript block: status
// icon, tool name, argument summary, and collapsible output.
//
// Block is a chat cell, not an MVU component: create it with New, keep the
// pointer, and mutate it as the tool progresses (SetID, SetStatus, AppendOutput,
// Retry, Cancel, Expanded). The chat transcript re-renders cells every frame,
// so mutations show up immediately.
package toolcall

import (
	"fmt"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/agentic/chat"
)

// Status is the lifecycle state of a tool call.
type Status int

const (
	StatusPending Status = iota
	StatusRunning
	StatusSuccess
	StatusError
	StatusCancelled
)

// Block is one tool call in a transcript. Mutate it via the exported fields
// and methods; it renders on demand.
type Block struct {
	id string

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

	attempt int

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
			StatusPending:   lipgloss.NewStyle().Foreground(theme.TextFaint),
			StatusRunning:   lipgloss.NewStyle().Foreground(theme.Warning),
			StatusSuccess:   lipgloss.NewStyle().Foreground(theme.Success),
			StatusError:     lipgloss.NewStyle().Foreground(theme.Danger),
			StatusCancelled: lipgloss.NewStyle().Foreground(theme.TextFaint),
		},
		nameStyle:   lipgloss.NewStyle().Foreground(theme.Text).Bold(true),
		summarySt:   lipgloss.NewStyle().Foreground(theme.TextMuted),
		outputStyle: lipgloss.NewStyle().Foreground(theme.TextMuted).Background(theme.SurfaceSunken),
		hintStyle:   lipgloss.NewStyle().Foreground(theme.TextFaint),
	}
}

// SetID assigns the application-owned stable identity for this tool call.
func (b *Block) SetID(id string) { b.id = id }

// CellID implements chat.CellIdentity.
func (b *Block) CellID() string { return b.id }

// CellKind implements chat.CellKind.
func (b *Block) CellKind() string { return "toolcall" }

// Lifecycle implements chat.CellLifecycle.
func (b *Block) Lifecycle() chat.State {
	switch b.status {
	case StatusPending:
		return chat.StatePending
	case StatusRunning:
		return chat.StateRunning
	case StatusSuccess:
		return chat.StateComplete
	case StatusError:
		return chat.StateFailed
	case StatusCancelled:
		return chat.StateCancelled
	default:
		return chat.StateFailed
	}
}

// Attempt returns the zero-based retry count for this tool call.
func (b *Block) Attempt() int { return b.attempt }

// Retry clears prior output, increments the attempt, and returns the block to
// pending state while preserving its stable identity.
func (b *Block) Retry() {
	b.attempt++
	b.status = StatusPending
	b.output = nil
	b.Expanded = false
}

// Cancel marks the tool call as cancelled without discarding its output.
func (b *Block) Cancel() { b.status = StatusCancelled }

// SetStatus updates the lifecycle state.
func (b *Block) SetStatus(s Status) { b.status = s }

// Status returns the lifecycle state.
func (b *Block) Status() Status { return b.status }

// AppendOutput adds output text (split on newlines) to the block.
func (b *Block) AppendOutput(s string) {
	b.output = append(b.output, strings.Split(strings.TrimRight(s, "\n"), "\n")...)
}

var icons = map[Status]string{
	StatusPending:   "○",
	StatusRunning:   "◐",
	StatusSuccess:   "✓",
	StatusError:     "✗",
	StatusCancelled: "⊘",
}

// Render draws the block at the given width (chat cell contract).
func (b *Block) Render(width int) string {
	if width <= 0 {
		return ""
	}
	header := b.renderHeader(width)

	if len(b.output) == 0 {
		return header
	}
	if !b.Expanded {
		hint := fmt.Sprintf("(%d output lines)", len(b.output))
		remaining := width - ansi.StringWidth(header)
		if remaining <= 1 {
			return header
		}
		hint = ansi.Truncate(hint, remaining-1, "…")
		return header + " " + b.hintStyle.Render(hint)
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
		rows = append(rows, b.renderOutputLine(line, width))
	}
	if hidden > 0 {
		hint := "  " + b.hintStyle.Render(fmt.Sprintf("… +%d more lines", hidden))
		rows = append(rows, ansi.Truncate(hint, width, ""))
	}
	return strings.Join(rows, "\n")
}

func (b *Block) renderHeader(width int) string {
	icon := icons[b.status]
	iconPrefix := b.iconStyles[b.status].Render(icon) + " "
	used := ansi.StringWidth(icon) + 1
	name := ansi.Truncate(b.Name, max(0, width-used), "…")
	header := iconPrefix + b.nameStyle.Render(name)
	used += ansi.StringWidth(name)
	if b.Summary != "" && used+1 < width {
		summary := ansi.Truncate(b.Summary, width-used-1, "…")
		header += " " + b.summarySt.Render(summary)
	}
	return ansi.Truncate(header, width, "")
}

func (b *Block) renderOutputLine(line string, width int) string {
	if width < 4 {
		return b.outputStyle.Render(ansi.Truncate(line, width, "…"))
	}
	line = ansi.Truncate(line, width-4, "…")
	pad := strings.Repeat(" ", max(0, width-4-ansi.StringWidth(line)))
	return "  " + b.outputStyle.Render(" "+line+pad+" ")
}
