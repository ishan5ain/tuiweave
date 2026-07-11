package viewport

import (
	"testing"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/inspect"
)

func TestSemanticScrollBottom(t *testing.T) {
	m := New(tuiweave.Dark())
	m.SetSize(10, 2)
	m.SetContent("one\ntwo\nthree")
	next, cmd := m.Update(inspect.Invoke(ActionBottom))
	if cmd != nil {
		t.Fatal("semantic scroll returned a command")
	}
	if got := next.YOffset(); got != 1 {
		t.Fatalf("offset = %d, want 1", got)
	}
}
