package dialog

import (
	"testing"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/inspect"
)

func TestSemanticCancelProducesResultCommand(t *testing.T) {
	m := New(tuiweave.Dark())
	m.ID = "quit"
	if node := m.Inspect(); !node.Focused || node.Status != "awaiting_input" {
		t.Fatalf("visible dialog inspection = %+v", node)
	}
	next, cmd := m.Update(inspect.Invoke(ActionCancel))
	if next.Title != m.Title {
		t.Fatal("semantic cancel unexpectedly changed dialog content")
	}
	if cmd == nil {
		t.Fatal("semantic cancel returned no command")
	}
	msg, ok := cmd().(ResultMsg)
	if !ok || msg.ID != "quit" || msg.OK {
		t.Fatalf("result = %#v, want quit/cancel", msg)
	}
}
