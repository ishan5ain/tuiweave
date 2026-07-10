// Package action defines stable, application-owned actions shared by
// selectable components such as menu, toolbar, and palette.
package action

// Item is one selectable application action. ID should remain stable across
// updates so semantic clients and application handlers can identify it.
// Disabled items remain visible but cannot be selected or activated.
type Item struct {
	ID          string
	Label       string
	Description string
	Disabled    bool
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
