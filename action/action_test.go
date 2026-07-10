package action_test

import (
	"testing"

	"github.com/ishansain/gotui/action"
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
