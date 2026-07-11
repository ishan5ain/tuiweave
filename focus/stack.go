package focus

// Group is the fresh set of component addresses belonging to one focus
// layer. A Stack never stores Groups: callers pass current addresses to Apply
// after every MVU update.
type Group []Focusable

// Stack is a copy-safe focus stack. Layer zero is the application's normal
// focus manager; each Push adds one nested modal or transient focus layer.
// Only the top layer receives focus when Apply is called.
//
// The stack stores managers, not component pointers. Mutating methods copy the
// internal slice before changing it, so a value copied by an MVU update cannot
// mutate the previous model's focus state through shared slice storage.
type Stack struct {
	managers []Manager
}

// NewStack returns a focus stack with one root layer containing n components.
// The root layer starts at index zero and is active until the first Push.
func NewStack(n int) Stack {
	return Stack{managers: []Manager{NewManager(n)}}
}

// Depth reports the number of nested layers above the root. Zero means that
// the application's normal focus group is active.
func (s Stack) Depth() int {
	if len(s.managers) == 0 {
		return 0
	}
	return len(s.managers) - 1
}

// Active reports whether a nested focus layer currently owns focus.
func (s Stack) Active() bool { return s.Depth() > 0 }

// Index returns the focused component index in the top layer. The zero value
// returns zero, matching a newly created manager.
func (s Stack) Index() int {
	if len(s.managers) == 0 {
		return 0
	}
	return s.managers[len(s.managers)-1].Index()
}

// Set moves focus in the top layer, wrapping in either direction.
func (s *Stack) Set(index int) {
	s.ensureRoot()
	s.clone()
	s.managers[len(s.managers)-1].Set(index)
}

// Next moves focus to the next component in the top layer, wrapping.
func (s *Stack) Next() {
	s.ensureRoot()
	s.clone()
	s.managers[len(s.managers)-1].Next()
}

// Prev moves focus to the previous component in the top layer, wrapping.
func (s *Stack) Prev() {
	s.ensureRoot()
	s.clone()
	s.managers[len(s.managers)-1].Prev()
}

// Push adds a nested focus layer with n components. Its first component is
// focused; the previous layer's index is retained until Pop.
func (s *Stack) Push(n int) {
	s.ensureRoot()
	s.clone()
	s.managers = append(s.managers, NewManager(n))
}

// Pop removes the top nested layer and restores the layer below it. It
// returns false when called on the root layer.
func (s *Stack) Pop() bool {
	if len(s.managers) <= 1 {
		return false
	}
	s.clone()
	s.managers = s.managers[:len(s.managers)-1]
	return true
}

// Apply focuses the top layer and blurs every other supplied group. Groups
// must be passed from the root layer outward, with one group per layer. Pass
// fresh component addresses each time; Groups are not retained.
//
// For a root, palette, and confirmation dialog, for example:
//
//	stack.Apply(
//		focus.Group{&m.tabs, &m.table},
//		focus.Group{&m.palette},
//		focus.Group{&m.confirm},
//	)
func (s Stack) Apply(groups ...Group) {
	active := s.Depth()
	for i, group := range groups {
		if i == active && i < len(s.managers) {
			s.managers[i].Apply(group...)
			continue
		}
		Blur(group...)
	}
}

func (s *Stack) ensureRoot() {
	if len(s.managers) == 0 {
		s.managers = []Manager{NewManager(0)}
	}
}

func (s *Stack) clone() {
	s.managers = append([]Manager(nil), s.managers...)
}
