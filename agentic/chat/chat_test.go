package chat

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/agentic/markdown"
	"github.com/ishansain/gotui/inspect"
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

type countingRenderer struct{ calls int }

func (r *countingRenderer) Render(source string, _ int) (string, error) {
	r.calls++
	return source, nil
}

func TestAssistantSetSourceRefreshesSameLengthContent(t *testing.T) {
	r := &countingRenderer{}
	a := NewAssistant(gotui.Dark(), r)
	a.SetSource("one")
	a.Render(20)
	a.SetSource("two")
	a.Render(20)
	if r.calls != 2 {
		t.Fatalf("renderer calls = %d, want 2 after same-length replacement", r.calls)
	}
}

func TestCellIdentityLookupAndReplace(t *testing.T) {
	c := newTranscript(30, 4)
	a := NewAssistant(gotui.Dark(), markdown.NewRenderer(gotui.Dark()))
	a.SetID("assistant-1")
	c.Append(a)

	found, ok := c.Cell("assistant-1")
	if !ok || found != a {
		t.Fatalf("Cell lookup = (%v, %v), want assistant-1", found, ok)
	}
	replacement := NewText(gotui.Dark(), "replayed")
	replacement.SetID("assistant-1")
	if !c.Replace("assistant-1", replacement) {
		t.Fatal("Replace did not find assistant-1")
	}
	found, ok = c.Cell("assistant-1")
	if !ok || found != replacement {
		t.Fatal("Cell lookup did not return replacement")
	}
}

func TestInspectIncludesIdentifiedLifecycleChildren(t *testing.T) {
	c := newTranscript(30, 4)
	a := NewAssistant(gotui.Dark(), markdown.NewRenderer(gotui.Dark()))
	a.SetID("assistant-1")
	c.Append(a)
	node := c.Inspect()
	if len(node.Children) != 1 || node.Children[0].ID != "assistant-1" || node.Children[0].Status != string(StateStreaming) {
		t.Fatalf("inspection children = %+v", node.Children)
	}
	if _, err := inspect.Marshal(node); err != nil {
		t.Fatalf("marshal inspection: %v", err)
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
