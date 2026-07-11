package list

import (
	"testing"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/inspect"
)

func TestSemanticActionNextMovesSelection(t *testing.T) {
	m := New(tuiweave.Dark())
	m.SetSize(20, 3)
	m.SetItems("one", "two", "three")
	m.Focus()

	next, cmd := m.Update(inspect.Invoke(ActionNext))
	if cmd != nil {
		t.Fatal("semantic navigation returned a command")
	}
	if got := next.SelectedItem(); got != "two" {
		t.Fatalf("selected item = %q, want two", got)
	}
}

func TestSemanticActionsExposeQualifiedIDsWhenBound(t *testing.T) {
	m := New(tuiweave.Dark())
	m.SetItems("one", "two")
	node := inspect.Bind("files", m)
	for _, action := range node.Actions {
		if len(action.ID) < len("files.") || action.ID[:len("files.")] != "files." {
			t.Fatalf("action ID %q is not qualified", action.ID)
		}
	}
}
