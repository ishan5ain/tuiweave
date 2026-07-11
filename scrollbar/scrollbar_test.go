package scrollbar

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/snaptest"
	"github.com/ishan5ain/tuiweave/viewport"
)

func bar(height, total, visible, offset int) string {
	return ansi.Strip(Vertical(tuiweave.Dark(), height, total, visible, offset))
}

func TestThumbPositionAndSize(t *testing.T) {
	top := bar(10, 100, 10, 0)
	if !strings.HasPrefix(top, "┃") {
		t.Errorf("offset 0: thumb not at top:\n%s", top)
	}
	bottom := bar(10, 100, 10, 90)
	if !strings.HasSuffix(bottom, "┃") {
		t.Errorf("max offset: thumb not at bottom:\n%s", bottom)
	}
	if got := strings.Count(top, "┃"); got != 1 {
		t.Errorf("10%% visible of height 10: thumb size = %d, want 1", got)
	}
	if got := strings.Count(bar(10, 20, 10, 0), "┃"); got != 5 {
		t.Errorf("50%% visible: thumb size = %d, want 5", got)
	}
}

func TestAllFitsRendersTrackOnly(t *testing.T) {
	b := bar(5, 3, 5, 0)
	if strings.Contains(b, "┃") {
		t.Errorf("content that fits shows a thumb:\n%s", b)
	}
	if got := strings.Count(b, "│"); got != 5 {
		t.Errorf("track rows = %d, want 5", got)
	}
}

func TestOffsetClamped(t *testing.T) {
	if got := bar(10, 100, 10, 9999); !strings.HasSuffix(got, "┃") {
		t.Errorf("overscrolled offset not clamped to bottom:\n%s", got)
	}
}

func TestZeroHeight(t *testing.T) {
	if got := Vertical(tuiweave.Dark(), 0, 10, 5, 0); got != "" {
		t.Errorf("height 0 rendered %q", got)
	}
}

func TestForViewportGolden(t *testing.T) {
	vp := viewport.New(tuiweave.Dark())
	vp.SetSize(10, 6)
	vp.SetContent(strings.Repeat("line\n", 30))
	vp.ScrollTo(12) // mid-way

	b := For(tuiweave.Dark(), &vp)
	if got := len(strings.Split(b, "\n")); got != 6 {
		t.Fatalf("bar height = %d, want 6 (match viewport)", got)
	}
	snaptest.SnapCells(t, b, snaptest.WithRoles(tuiweave.Dark()))
}
