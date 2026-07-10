package dialog

import (
	"testing"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/inspect"
)

func TestSemanticCancelProducesResultCommand(t *testing.T) {
	m := New(gotui.Dark())
	m.ID = "quit"
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
