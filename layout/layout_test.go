package layout

import "testing"

func TestVerticalSplitAssign(t *testing.T) {
	var header, body, status Rect
	Vertical(Len(3), Fill(1), Len(1)).
		Split(NewRect(0, 0, 80, 24)).
		Assign(&header, &body, &status)

	if got := header.Dy(); got != 3 {
		t.Errorf("header height = %d, want 3", got)
	}
	if got := body.Dy(); got != 20 {
		t.Errorf("body height = %d, want 20", got)
	}
	if got := status.Dy(); got != 1 {
		t.Errorf("status height = %d, want 1", got)
	}
	for _, r := range []Rect{header, body, status} {
		if got := r.Dx(); got != 80 {
			t.Errorf("segment width = %d, want 80", got)
		}
	}
}

func TestHorizontalPercentFill(t *testing.T) {
	rects := Horizontal(Percent(25), Fill(1)).Split(NewRect(0, 0, 100, 10))
	if len(rects) != 2 {
		t.Fatalf("got %d rects, want 2", len(rects))
	}
	if got := rects[0].Dx(); got != 25 {
		t.Errorf("sidebar width = %d, want 25", got)
	}
	if got := rects[1].Dx(); got != 75 {
		t.Errorf("main width = %d, want 75", got)
	}
}

type fakeComponent struct{ w, h int }

func (f *fakeComponent) SetSize(w, h int) { f.w, f.h = w, h }

func TestApplySizesComponents(t *testing.T) {
	var top, bottom fakeComponent
	rects := Vertical(Len(5), Fill(1)).Apply(NewRect(0, 0, 40, 20), &top, &bottom)

	if top.w != 40 || top.h != 5 {
		t.Errorf("top sized %dx%d, want 40x5", top.w, top.h)
	}
	if bottom.w != 40 || bottom.h != 15 {
		t.Errorf("bottom sized %dx%d, want 40x15", bottom.w, bottom.h)
	}
	if len(rects) != 2 {
		t.Errorf("got %d rects, want 2", len(rects))
	}
}

func TestApplyExtraComponentsUntouched(t *testing.T) {
	var a, b fakeComponent
	Vertical(Fill(1)).Apply(NewRect(0, 0, 10, 10), &a, &b)
	if b.w != 0 || b.h != 0 {
		t.Errorf("extra component was sized %dx%d, want untouched", b.w, b.h)
	}
}
