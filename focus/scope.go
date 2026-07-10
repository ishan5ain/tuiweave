package focus

// Scope is a value-type focus scope for modal or conditionally visible
// components. It stores only a child Manager, an active bit, and the parent's
// previous index; callers still pass fresh component addresses to Apply.
type Scope struct {
	manager  Manager
	active   bool
	parentIx int
}

// NewScope returns an inactive scope with a child tab order of n components.
// The first child is focused when Enter activates the scope.
func NewScope(n int) Scope { return Scope{manager: NewManager(n)} }

// Enter activates the scope and remembers the parent's current focus index.
// Re-entering an active scope does not overwrite the saved parent position.
func (s *Scope) Enter(parent Manager) {
	if s.active {
		return
	}
	s.parentIx = parent.Index()
	s.active = true
	s.manager.Set(0)
}

// Exit deactivates the scope and restores the parent's saved focus index.
// ApplyBackground must be called afterward to push the restored state into
// fresh component addresses.
func (s *Scope) Exit(parent *Manager) {
	if !s.active {
		return
	}
	parent.Set(s.parentIx)
	s.active = false
}

// Active reports whether the scoped focus group is currently active.
func (s Scope) Active() bool { return s.active }

// Set moves the child scope's focus index, wrapping in either direction.
func (s *Scope) Set(index int) { s.manager.Set(index) }

// Next moves to the next component in the active child scope.
func (s *Scope) Next() { s.manager.Next() }

// Prev moves to the previous component in the active child scope.
func (s *Scope) Prev() { s.manager.Prev() }

// Index returns the child scope's focus index.
func (s Scope) Index() int { return s.manager.Index() }

// ApplyBackground applies the parent manager while the scope is inactive. If
// the scope is active, every background component is blurred instead.
func (s Scope) ApplyBackground(parent Manager, items ...Focusable) {
	if s.active {
		Blur(items...)
		return
	}
	parent.Apply(items...)
}

// Apply focuses the scoped components while active. When inactive, it blurs
// them so a hidden modal cannot retain keyboard focus.
func (s Scope) Apply(items ...Focusable) {
	if s.active {
		s.manager.Apply(items...)
		return
	}
	Blur(items...)
}

// Blur clears focus from a group of components.
func Blur(items ...Focusable) {
	for _, item := range items {
		item.Blur()
	}
}
