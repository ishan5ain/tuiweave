package layout

import (
	"fmt"
	"testing"
)

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

func TestConstraintFormattingRemainsStable(t *testing.T) {
	tests := []struct {
		constraint Constraint
		want       string
	}{
		{Len(2), "Len(2)"},
		{Min(1), "Min(1)"},
		{Max(4), "Max(4)"},
		{Percent(10), "Percent(10)"},
		{Ratio(1, 4), "Ratio(1 / 4)"},
		{Fill(3), "Fill(3)"},
	}

	for _, tt := range tests {
		if got := fmt.Sprint(tt.constraint); got != tt.want {
			t.Errorf("formatted constraint = %q, want %q", got, tt.want)
		}
	}
}

func TestConstraintPrioritiesRemainStable(t *testing.T) {
	tests := []struct {
		name        string
		constraints []Constraint
		wantWidths  []int
	}{
		{
			name:        "minimum wins over percentage",
			constraints: []Constraint{Percent(100), Min(20)},
			wantWidths:  []int{30, 20},
		},
		{
			name:        "maximum caps a fill sibling",
			constraints: []Constraint{Fill(1), Max(20)},
			wantWidths:  []int{30, 20},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rects := Horizontal(tt.constraints...).Split(NewRect(0, 0, 50, 4))
			for i, want := range tt.wantWidths {
				if got := rects[i].Dx(); got != want {
					t.Errorf("segment %d width = %d, want %d", i, got, want)
				}
			}
		})
	}
}

func TestRatioAndWeightedFillRemainStable(t *testing.T) {
	rects := Horizontal(Ratio(1, 4), Fill(1), Fill(2)).Split(NewRect(0, 0, 100, 4))
	want := []int{25, 25, 50}
	for i, width := range want {
		if got := rects[i].Dx(); got != width {
			t.Errorf("segment %d width = %d, want %d", i, got, width)
		}
	}
}

func TestPaddingAndSpacingRemainStable(t *testing.T) {
	rects := Horizontal(Len(10), Fill(1)).
		WithPadding(2).
		WithSpacing(3).
		Split(NewRect(0, 0, 40, 10))

	want := []Rect{
		NewRect(2, 2, 10, 6),
		NewRect(15, 2, 23, 6),
	}
	for i := range want {
		if rects[i] != want[i] {
			t.Errorf("segment %d = %v, want %v", i, rects[i], want[i])
		}
	}
}

type fakeComponent struct{ w, h int }

func (f *fakeComponent) SetSize(w, h int) { f.w, f.h = w, h }

type fakeIntrinsic struct{}

func (fakeIntrinsic) SetSize(int, int)   {}
func (fakeIntrinsic) SizeMode() SizeMode { return SizeIntrinsic }

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

func TestSizeModeDefaultsToBounded(t *testing.T) {
	if got := SizeModeOf(&fakeComponent{}); got != SizeBounded {
		t.Fatalf("ordinary component mode = %d, want SizeBounded", got)
	}
	if got := SizeModeOf(fakeIntrinsic{}); got != SizeIntrinsic {
		t.Fatalf("intrinsic component mode = %d, want SizeIntrinsic", got)
	}
}
