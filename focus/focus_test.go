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
	fm.Apply()
}
