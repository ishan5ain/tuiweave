// Package action defines stable, application-owned actions shared by
// selectable components such as menu, toolbar, and palette.
package action

import "strings"

// SelectPrefix prefixes semantic actions that select an item by stable ID.
const SelectPrefix = "select."

// Item is one selectable application action. ID should remain stable across
// updates so semantic clients and application handlers can identify it.
// Disabled items remain visible but cannot be selected or activated.
type Item struct {
	ID          string
	Label       string
	Description string
	Disabled    bool
}

// SelectID returns the semantic action ID for selecting an item by stable ID.
func SelectID(id string) string { return SelectPrefix + id }

// ParseSelectID extracts a stable item ID from a semantic selection action.
// It returns false for unrelated actions and for an empty item ID.
func ParseSelectID(id string) (string, bool) {
	if !strings.HasPrefix(id, SelectPrefix) {
		return "", false
	}
	itemID := strings.TrimPrefix(id, SelectPrefix)
	return itemID, itemID != ""
}

// Enabled reports whether the action can be selected or activated.
func (i Item) Enabled() bool { return !i.Disabled }

// Find returns the action with id, if present.
func Find(items []Item, id string) (Item, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return Item{}, false
}

// EnabledCount reports how many actions can be selected or activated.
func EnabledCount(items []Item) int {
	count := 0
	for _, item := range items {
		if item.Enabled() {
			count++
		}
	}
	return count
}
