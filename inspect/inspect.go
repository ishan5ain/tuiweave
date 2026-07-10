// Package inspect defines an optional, data-only semantic representation of a
// gotui interface. It does not own component routing, layout, or rendering.
// Applications assemble a tree from component reports and may expose it to
// tests, debugging tools, or agents.
package inspect

import (
	"bytes"
	"encoding/json"

	"github.com/ishansain/gotui/layout"
)

// Bounds is a component's terminal rectangle in cell coordinates.
type Bounds struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// FromRect converts a gotui layout rectangle to inspection bounds.
func FromRect(r layout.Rect) Bounds {
	return Bounds{X: r.Min.X, Y: r.Min.Y, Width: r.Dx(), Height: r.Dy()}
}

// Selection describes the selected item or row in a component.
type Selection struct {
	Index int    `json:"index"`
	Count int    `json:"count"`
	Label string `json:"label,omitempty"`
}

// Scroll describes a component's current scroll window.
type Scroll struct {
	Total   int `json:"total"`
	Visible int `json:"visible"`
	Offset  int `json:"offset"`
}

// Node is one semantic UI element. Attributes are deliberately string-valued:
// component packages can expose small, stable facts without making inspect a
// second domain model or leaking backend-specific types into the core.
type Node struct {
	ID         string            `json:"id"`
	Kind       string            `json:"kind"`
	Bounds     Bounds            `json:"bounds"`
	Label      string            `json:"label,omitempty"`
	Focused    bool              `json:"focused"`
	Status     string            `json:"status,omitempty"`
	Selected   *Selection        `json:"selected,omitempty"`
	Scroll     *Scroll           `json:"scroll,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Children   []Node            `json:"children,omitempty"`
}

// Inspectable is implemented by components that can report semantic state.
// The returned node has no application-assigned identity; use Bind when
// assembling it into an application's inspection tree.
type Inspectable interface {
	Inspect() Node
}

// Bind gives a component report a stable application-owned ID.
func Bind(id string, component Inspectable) Node {
	node := component.Inspect()
	node.ID = id
	return node
}

// BindAt gives a component report an application-owned ID and screen bounds.
// Use this when the app already has a layout rectangle for the component.
func BindAt(id string, bounds Bounds, component Inspectable) Node {
	node := Bind(id, component)
	node.Bounds = bounds
	return node
}

// Group creates an application-owned container node with ordered children.
func Group(id, kind string, bounds Bounds, children ...Node) Node {
	return Node{ID: id, Kind: kind, Bounds: bounds, Children: children}
}

// Marshal returns deterministic, indented JSON for a semantic node tree.
// encoding/json sorts map keys, so Attributes remain stable in snapshots.
func Marshal(node Node) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(node); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}
