package markdown

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/snaptest"
)

const sample = `# Title

Some **bold** and *italic* text with ` + "`inline code`" + ` and a [link](https://example.com).

- first item
- second item

> a quote

` + "```go\nfunc main() {\n\tfmt.Println(\"hi\") // greet\n}\n```" + `
`

func TestRenderGolden(t *testing.T) {
	r := NewRenderer(gotui.Dark())
	out, err := r.Render(sample, 40)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	snaptest.Snap(t, out)
	snaptest.SnapCells(t, out, snaptest.WithRoles(gotui.Dark()))
}

func TestRenderWrapsToWidth(t *testing.T) {
	r := NewRenderer(gotui.Dark())
	out, err := r.Render(strings.Repeat("word ", 30), 24)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	for i, line := range strings.Split(out, "\n") {
		if got := ansi.StringWidth(line); got > 24 {
			t.Errorf("line %d width = %d, want <= 24: %q", i, got, line)
		}
	}
}

func TestRenderDeterministic(t *testing.T) {
	r := NewRenderer(gotui.Dark())
	a, _ := r.Render(sample, 40)
	b, _ := r.Render(sample, 40)
	if a != b {
		t.Error("two renders of the same source differ")
	}
}

func TestRendererCachePerWidth(t *testing.T) {
	r := NewRenderer(gotui.Dark()).(*glamourRenderer)
	if _, err := r.Render("hi", 40); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Render("hi", 40); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Render("hi", 20); err != nil {
		t.Fatal(err)
	}
	if got := len(r.renderers); got != 2 {
		t.Errorf("cached %d renderers, want 2 (one per width)", got)
	}
}

func TestSprintFallsBackOnError(t *testing.T) {
	r := NewRenderer(gotui.Dark())
	if got := Sprint(r, "plain", 0); got != "plain" { // width 0 errors
		t.Errorf("Sprint fallback = %q, want raw source", got)
	}
}

func TestRenderInvalidWidth(t *testing.T) {
	r := NewRenderer(gotui.Dark())
	if _, err := r.Render("x", 0); err == nil {
		t.Error("Render(width=0) returned nil error")
	}
}
