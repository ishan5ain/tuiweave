package inspect

import (
	"strings"
	"testing"

	"github.com/ishan5ain/tuiweave/layout"
)

type fakeComponent struct{}

func (fakeComponent) Inspect() Node {
	return Node{Kind: "fake", Bounds: Bounds{Width: 4, Height: 1}}
}

func TestBindAndMarshal(t *testing.T) {
	node := Group("app", "application", FromRect(layout.NewRect(0, 0, 20, 4)),
		Bind("input", fakeComponent{}),
	)
	got, err := Marshal(node)
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	for _, want := range []string{`"id": "app"`, `"id": "input"`, `"kind": "fake"`, `"width": 20`, `"height": 4`} {
		if !strings.Contains(text, want) {
			t.Errorf("marshal missing %q in:\n%s", want, text)
		}
	}
}

func TestBindAtOverridesComponentBounds(t *testing.T) {
	node := BindAt("fake", Bounds{X: 3, Y: 2, Width: 8, Height: 1}, fakeComponent{})
	if node.ID != "fake" || node.Bounds.X != 3 || node.Bounds.Y != 2 || node.Bounds.Width != 8 {
		t.Fatalf("BindAt = %+v", node)
	}
}
