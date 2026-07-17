package layout_test

import (
	"image"
	"testing"

	"github.com/ishan5ain/tuiweave/layout"
)

func TestRectUsesStandardLibraryIdentity(t *testing.T) {
	var standard image.Rectangle = layout.NewRect(2, 3, 11, 7)
	var public layout.Rect = image.Rect(2, 3, 13, 10)

	if standard != public {
		t.Fatalf("layout rect = %v, want %v", public, standard)
	}
	if public.Dx() != 11 || public.Dy() != 7 {
		t.Fatalf("layout rect size = %dx%d, want 11x7", public.Dx(), public.Dy())
	}

	reversed := layout.NewRect(5, 6, -2, -3)
	wantReversed := image.Rectangle{Min: image.Pt(5, 6), Max: image.Pt(3, 3)}
	if reversed != wantReversed {
		t.Fatalf("reversed layout rect = %v, want %v", reversed, wantReversed)
	}
}

func TestDocumentedConstraintConstructorsCompose(t *testing.T) {
	constraints := []layout.Constraint{
		layout.Len(2),
		layout.Min(1),
		layout.Max(4),
		layout.Percent(10),
		layout.Ratio(1, 4),
		layout.Fill(1),
	}

	parts := layout.Horizontal(constraints...).Split(layout.NewRect(0, 0, 40, 3))
	if len(parts) != len(constraints) {
		t.Fatalf("split returned %d parts, want %d", len(parts), len(constraints))
	}
}
