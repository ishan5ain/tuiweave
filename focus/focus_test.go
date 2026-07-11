package focus

import "testing"

type fake struct{ focused bool }

func (f *fake) Focus() { f.focused = true }
func (f *fake) Blur()  { f.focused = false }

func exactlyOneFocused(t *testing.T, items []*fake, want int) {
	t.Helper()
	for i, f := range items {
		if f.focused != (i == want) {
			t.Fatalf("item %d focused=%v, want focus only on %d", i, f.focused, want)
		}
	}
}

func TestManagerCyclesAndApplies(t *testing.T) {
	a, b, c := &fake{}, &fake{}, &fake{}
	items := []*fake{a, b, c}
	fm := NewManager(3)

	fm.Apply(a, b, c)
	exactlyOneFocused(t, items, 0)

	fm.Next()
	fm.Apply(a, b, c)
	exactlyOneFocused(t, items, 1)

	fm.Next()
	fm.Next() // wraps
	fm.Apply(a, b, c)
	exactlyOneFocused(t, items, 0)

	fm.Prev() // wraps backward
	fm.Apply(a, b, c)
	exactlyOneFocused(t, items, 2)
	if fm.Index() != 2 {
		t.Errorf("Index() = %d, want 2", fm.Index())
	}
}

func TestManagerSurvivesValueCopies(t *testing.T) {
	// Manager must stay correct when embedded in an MVU model that is
	// copied on every Update: state lives in the index, not in pointers.
	fm := NewManager(2)
	fm.Next()
	fmCopy := fm

	a, b := &fake{}, &fake{}
	fmCopy.Apply(a, b)
	exactlyOneFocused(t, []*fake{a, b}, 1)
}

func TestManagerSetOutOfRangeWraps(t *testing.T) {
	fm := NewManager(2)
	fm.Set(5)
	if fm.Index() != 1 {
		t.Errorf("Set(5): Index() = %d, want 1", fm.Index())
	}
	fm.Set(-1)
	if fm.Index() != 1 {
		t.Errorf("Set(-1): Index() = %d, want 1", fm.Index())
	}
}

func TestManagerEmpty(t *testing.T) {
	fm := NewManager(0)
	fm.Next() // must not panic
	fm.Prev()
	fm.Set(3)
	a, b := &fake{focused: true}, &fake{focused: true}
	fm.Apply(a, b)
	if a.focused || b.focused {
		t.Fatalf("empty manager retained focus: a=%v b=%v", a.focused, b.focused)
	}
}

func TestZeroValueStackBlursRootGroup(t *testing.T) {
	a, b := &fake{focused: true}, &fake{focused: true}
	var stack Stack
	stack.Apply(Group{a, b})
	if a.focused || b.focused {
		t.Fatalf("zero stack retained focus: a=%v b=%v", a.focused, b.focused)
	}
}

func TestScopeRestoresParentAndIsolatesBackground(t *testing.T) {
	background := []*fake{{}, {}, {}}
	modal := []*fake{{}, {}}
	parent := NewManager(len(background))
	parent.Set(1)
	parent.Apply(toFocusables(background)...)
	scope := NewScope(len(modal))

	scope.Enter(parent)
	scope.ApplyBackground(parent, toFocusables(background)...)
	scope.Apply(toFocusables(modal)...)
	if !scope.Active() || scope.Index() != 0 {
		t.Fatalf("active scope=%v index=%d, want active index 0", scope.Active(), scope.Index())
	}
	for i, item := range background {
		if item.focused {
			t.Fatalf("background item %d retained focus while scope active", i)
		}
	}
	exactlyOneFocused(t, modal, 0)

	scope.Next()
	scope.Apply(toFocusables(modal)...)
	exactlyOneFocused(t, modal, 1)
	scope.Exit(&parent)
	scope.ApplyBackground(parent, toFocusables(background)...)
	scope.Apply(toFocusables(modal)...)
	if scope.Active() || parent.Index() != 1 {
		t.Fatalf("after exit active=%v parent index=%d, want inactive index 1", scope.Active(), parent.Index())
	}
	exactlyOneFocused(t, background, 1)
	for i, item := range modal {
		if item.focused {
			t.Fatalf("modal item %d retained focus after scope exit", i)
		}
	}
}

func TestScopeSurvivesValueCopies(t *testing.T) {
	scope := NewScope(2)
	scope.Enter(NewManager(1))
	scope.Next()
	copy := scope
	a, b := &fake{}, &fake{}
	copy.Apply(a, b)
	exactlyOneFocused(t, []*fake{a, b}, 1)
}

func TestStackNestsAndRestoresFocusLayers(t *testing.T) {
	background := []*fake{{}, {}, {}}
	palette := []*fake{{}}
	confirm := []*fake{{}, {}}
	stack := NewStack(len(background))
	stack.Set(2)
	groups := func() []Group {
		return []Group{
			Group{background[0], background[1], background[2]},
			Group{palette[0]},
			Group{confirm[0], confirm[1]},
		}
	}

	stack.Apply(groups()...)
	exactlyOneFocused(t, background, 2)

	stack.Push(1)
	stack.Apply(groups()...)
	if stack.Depth() != 1 || !stack.Active() {
		t.Fatalf("after palette push: depth=%d active=%v", stack.Depth(), stack.Active())
	}
	exactlyOneFocused(t, palette, 0)
	for i, item := range background {
		if item.focused {
			t.Fatalf("background item %d retained focus under palette", i)
		}
	}

	stack.Push(2)
	stack.Apply(groups()...)
	if stack.Depth() != 2 || stack.Index() != 0 {
		t.Fatalf("after confirmation push: depth=%d index=%d", stack.Depth(), stack.Index())
	}
	exactlyOneFocused(t, confirm, 0)
	stack.Next()
	stack.Apply(groups()...)
	exactlyOneFocused(t, confirm, 1)

	if !stack.Pop() {
		t.Fatal("Pop() returned false for nested confirmation")
	}
	stack.Apply(groups()...)
	if stack.Depth() != 1 || stack.Index() != 0 {
		t.Fatalf("after confirmation pop: depth=%d index=%d", stack.Depth(), stack.Index())
	}
	exactlyOneFocused(t, palette, 0)
	for i, item := range confirm {
		if item.focused {
			t.Fatalf("confirmation item %d retained focus after pop", i)
		}
	}

	if !stack.Pop() {
		t.Fatal("Pop() returned false for palette")
	}
	stack.Apply(groups()...)
	if stack.Depth() != 0 || stack.Active() || stack.Index() != 2 {
		t.Fatalf("after palette pop: depth=%d active=%v index=%d", stack.Depth(), stack.Active(), stack.Index())
	}
	exactlyOneFocused(t, background, 2)
	if stack.Pop() {
		t.Fatal("Pop() returned true at root")
	}
}

func TestStackMutationsSurviveValueCopies(t *testing.T) {
	original := NewStack(2)
	original.Set(1)

	copy := original
	copy.Push(1)
	copy.Next()
	if original.Depth() != 0 || original.Index() != 1 {
		t.Fatalf("original changed through copied stack: depth=%d index=%d", original.Depth(), original.Index())
	}
	if copy.Depth() != 1 || copy.Index() != 0 {
		t.Fatalf("copy state: depth=%d index=%d, want depth 1 index 0", copy.Depth(), copy.Index())
	}
}

func toFocusables(items []*fake) []Focusable {
	result := make([]Focusable, len(items))
	for i, item := range items {
		result[i] = item
	}
	return result
}
