package toolcall

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/agentic/chat"
	"github.com/ishansain/gotui/inspect"
	"github.com/ishansain/gotui/snaptest"
)

func newTestBlock() *Block {
	b := New(gotui.Dark(), "Bash", "go test ./...")
	b.SetStatus(StatusSuccess)
	b.AppendOutput("ok  \tgotui/layout\t0.3s\nok  \tgotui/snaptest\t0.2s")
	return b
}

func TestCollapsedGolden(t *testing.T) {
	b := newTestBlock()
	view := b.Render(50)
	if !strings.Contains(view, "(2 output lines)") {
		t.Errorf("collapsed block missing line-count hint: %q", view)
	}
	snaptest.SnapCells(t, view, snaptest.WithRoles(gotui.Dark()))
}

func TestExpandedGolden(t *testing.T) {
	b := newTestBlock()
	b.Expanded = true
	snaptest.Snap(t, b.Render(50))
	snaptest.SnapCells(t, b.Render(50), snaptest.WithRoles(gotui.Dark()))
}

func TestOutputCapAndHiddenCount(t *testing.T) {
	b := New(gotui.Dark(), "Read", "main.go")
	b.MaxOutputLines = 3
	b.Expanded = true
	b.AppendOutput("l1\nl2\nl3\nl4\nl5")
	view := b.Render(40)
	if !strings.Contains(view, "+2 more lines") {
		t.Errorf("capped output missing hidden count: %q", view)
	}
	if strings.Contains(view, "l4") {
		t.Errorf("capped output shows hidden line: %q", view)
	}
}

func TestStatusIcons(t *testing.T) {
	b := New(gotui.Dark(), "Bash", "")
	for status, icon := range icons {
		b.SetStatus(status)
		if !strings.Contains(b.Render(30), icon) {
			t.Errorf("status %d: icon %q missing", status, icon)
		}
	}
}

func TestNoOutputNoHint(t *testing.T) {
	b := New(gotui.Dark(), "Bash", "ls")
	if view := b.Render(30); strings.Contains(view, "output lines") {
		t.Errorf("block without output shows hint: %q", view)
	}
}

func TestIdentityLifecycleCancelAndRetry(t *testing.T) {
	b := New(gotui.Dark(), "Bash", "go test ./...")
	b.SetID("tool-1")
	b.SetStatus(StatusRunning)
	b.AppendOutput("partial output")
	b.Cancel()
	if b.CellID() != "tool-1" || b.Lifecycle() != chat.StateCancelled {
		t.Fatalf("cancelled block identity/state = %q/%q", b.CellID(), b.Lifecycle())
	}
	b.Retry()
	if b.Attempt() != 1 || b.Status() != StatusPending || b.Lifecycle() != chat.StatePending {
		t.Fatalf("retried block = attempt %d, status %d, state %q", b.Attempt(), b.Status(), b.Lifecycle())
	}
	if strings.Contains(b.Render(40), "output lines") {
		t.Fatal("retry retained prior output")
	}
	if !b.ApplyAction(ActionCancel) || b.Status() != StatusCancelled {
		t.Fatal("cancel action did not cancel the retried block")
	}
	b.SetStatus(StatusError)
	if !b.ApplyAction(ActionRetry) || b.Status() != StatusPending {
		t.Fatal("retry action did not return an errored block to pending")
	}

	node := inspect.Bind("tools", b)
	if node.ID != "tools" || b.CellKind() != "toolcall" {
		t.Fatalf("bound tool node = %+v", node)
	}
}

func TestTranscriptInspectionIncludesToolCallMetadata(t *testing.T) {
	c := chat.New(gotui.Dark())
	c.SetSize(40, 5)
	b := New(gotui.Dark(), "Bash", "go test ./...")
	b.SetID("tool-1")
	c.Append(b)

	node := c.Inspect()
	if len(node.Children) != 1 {
		t.Fatalf("transcript children = %+v", node.Children)
	}
	child := node.Children[0]
	if child.ID != "tool-1" || child.Kind != "toolcall" || child.Status != string(chat.StatePending) {
		t.Fatalf("tool child = %+v", child)
	}
	if len(child.Actions) != 2 {
		t.Fatalf("tool actions = %+v", child.Actions)
	}
}

func TestRenderStaysWithinWidth(t *testing.T) {
	b := New(gotui.Dark(), "界工具", "执行 界界界")
	b.Expanded = true
	b.AppendOutput("界 output\nsecond line")
	for _, width := range []int{1, 2, 3, 5, 8, 12, 20} {
		for lineNo, line := range strings.Split(b.Render(width), "\n") {
			if got := ansi.StringWidth(line); got > width {
				t.Fatalf("width %d line %d rendered as %d: %q", width, lineNo+1, got, line)
			}
		}
	}
}
