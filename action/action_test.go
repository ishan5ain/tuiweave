package action_test

import (
	"testing"

	"github.com/ishan5ain/tuiweave/action"
)

func TestItemHelpers(t *testing.T) {
	items := []action.Item{
		{ID: "open", Label: "Open"},
		{ID: "delete", Label: "Delete", Disabled: true},
	}

	if got := action.EnabledCount(items); got != 1 {
		t.Fatalf("EnabledCount() = %d, want 1", got)
	}

	item, ok := action.Find(items, "delete")
	if !ok || item.Label != "Delete" || item.Enabled() {
		t.Fatalf("Find() = %#v, %v; want disabled Delete", item, ok)
	}

	if _, ok := action.Find(items, "missing"); ok {
		t.Fatal("Find() found a missing action")
	}
}

func TestSelectionActionIDs(t *testing.T) {
	if got := action.SelectID("open"); got != "select.open" {
		t.Fatalf("SelectID() = %q, want select.open", got)
	}

	if got, ok := action.ParseSelectID("select.open"); !ok || got != "open" {
		t.Fatalf("ParseSelectID() = %q, %v; want open, true", got, ok)
	}
	for _, id := range []string{"activate", "select."} {
		if got, ok := action.ParseSelectID(id); ok || got != "" {
			t.Fatalf("ParseSelectID(%q) = %q, %v; want empty, false", id, got, ok)
		}
	}
}
