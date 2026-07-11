// Package focus provides tab-order managers and copy-safe focus scopes for
// interactive components.
//
// It exists to close a known failure mode of hand-wired MVU apps: focus
// state drifting across components (two things focused, or none).
//
// Manager is a value type holding only the focus index — safe to embed in an
// MVU model that is copied on every Update. It never stores component
// pointers: pass fresh addresses to Apply after every change, and exactly
// one component ends up focused.
//
//	// in the model:  fm focus.Manager  (focus.NewManager(3))
//	case "tab":
//	    m.fm.Next()
//	    m.fm.Apply(&m.list, &m.view, &m.input)
//
// For nested modals, use Stack. It stores one Manager per layer and callers
// pass fresh component groups to Apply after every model update.
package focus

// Focusable is implemented by interactive tuiweave components.
type Focusable interface {
	Focus()
	Blur()
}

// Manager owns a focus index over a fixed number of components, in tab
// order. The zero value manages zero components; create one with NewManager.
type Manager struct {
	n, idx int
}

// NewManager returns a manager over n components with the first focused.
// Call Apply to push that state into the components.
func NewManager(n int) Manager {
	return Manager{n: n}
}

// Set moves focus to index i, wrapping in either direction.
func (m *Manager) Set(i int) {
	if m.n == 0 {
		return
	}
	m.idx = ((i % m.n) + m.n) % m.n
}

// Next moves focus to the next component, wrapping.
func (m *Manager) Next() { m.Set(m.idx + 1) }

// Prev moves focus to the previous component, wrapping.
func (m *Manager) Prev() { m.Set(m.idx - 1) }

// Index returns the index of the focused component.
func (m Manager) Index() int { return m.idx }

// Apply focuses the component at the current index and blurs all others.
// Pass the components in the same order as their tab order, with their
// current addresses (i.e. call this inside Update, after mutating focus).
func (m Manager) Apply(items ...Focusable) {
	for i, item := range items {
		if m.n > 0 && i == m.idx {
			item.Focus()
		} else {
			item.Blur()
		}
	}
}
