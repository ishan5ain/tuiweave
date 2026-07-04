package chat

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/agentic/markdown"
	"github.com/ishansain/gotui/snaptest"
)

func newTranscript(w, h int) Model {
	c := New(gotui.Dark())
	c.SetSize(w, h)
	return c
}

func TestTranscriptGolden(t *testing.T) {
	c := newTranscript(44, 14)
	r := markdown.NewRenderer(gotui.Dark())

	c.Append(NewText(gotui.Dark(), "session started"))
	c.Append(NewUser(gotui.Dark(), "explain the layout package"))
	a := NewAssistant(gotui.Dark(), r)
	a.Append("The **layout** package splits space with:\n\n- `Len` fixed\n- `Fill` grow\n")
	c.Append(a)
	c.Invalidate()

	snaptest.Snap(t, c.View())
}

func TestAutoFollowStreaming(t *testing.T) {
	c := newTranscript(30, 4)
	r := markdown.NewRenderer(gotui.Dark())
	a := NewAssistant(gotui.Dark(), r)
	c.Append(a)

	for i := range 20 {
		a.Append(fmt.Sprintf("line %d\n\n", i))
		c.Invalidate()
	}
	if !c.Following() {
		t.Fatal("transcript stopped following during streaming")
	}
	if !strings.Contains(ansi.Strip(c.View()), "line 19") {
		t.Errorf("view not at bottom while following:\n%s", ansi.Strip(c.View()))
	}
}

func TestScrollUpUnsticksScrollBottomResticks(t *testing.T) {
	c := newTranscript(30, 4)
	for i := range 20 {
		c.Append(NewText(gotui.Dark(), fmt.Sprintf("note %d", i)))
	}
	c.Focus()

	c, _ = c.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	if c.Following() {
		t.Fatal("scrolling up did not unstick auto-follow")
	}
	frozen := c.View()
	c.Append(NewText(gotui.Dark(), "newest"))
	if strings.Contains(ansi.Strip(c.View()), "newest") {
		t.Error("unstuck transcript jumped to bottom on new content")
	}
	if got := c.View(); got != frozen {
		t.Error("unstuck transcript moved when content was appended")
	}

	c, _ = c.Update(tea.KeyPressMsg{Code: 'G', Text: "G"})
	if !c.Following() {
		t.Error("scrolling to bottom did not re-stick auto-follow")
	}
	if !strings.Contains(ansi.Strip(c.View()), "newest") {
		t.Error("view not at bottom after re-stick")
	}
}

func TestAssistantRenderCache(t *testing.T) {
	r := markdown.NewRenderer(gotui.Dark())
	a := NewAssistant(gotui.Dark(), r)
	a.Append("hello **world**")

	first := a.Render(30)
	if second := a.Render(30); second != first {
		t.Error("cached render differs")
	}
	a.Append(" more")
	if after := a.Render(30); after == first {
		t.Error("render did not refresh after Append")
	}
	if wider := a.Render(40); wider == "" {
		t.Error("render at new width is empty")
	}
}

func TestCellFuncAdapter(t *testing.T) {
	c := newTranscript(20, 3)
	c.Append(CellFunc(func(w int) string {
		return strings.Repeat("~", w)
	}))
	if !strings.Contains(c.View(), strings.Repeat("~", 20)) {
		t.Error("CellFunc cell not rendered at transcript width")
	}
}
