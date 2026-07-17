package spinner

import (
	"testing"
	"time"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/layout"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func TestSpinnerAdvancesOnOwnTick(t *testing.T) {
	s := New(tuiweave.Dark())
	first := s.View()

	s, cmd := s.Update(TickMsg{Time: time.Now(), id: s.id, tag: s.tag})
	if cmd == nil {
		t.Fatal("Update on own TickMsg returned nil cmd, want next tick")
	}
	if s.View() == first {
		t.Error("frame did not advance on tick")
	}
}

func TestSpinnerIgnoresForeignAndStaleTicks(t *testing.T) {
	s := New(tuiweave.Dark())
	first := s.View()

	s2, cmd := s.Update(TickMsg{Time: time.Now(), id: s.id + 99, tag: s.tag})
	if cmd != nil || s2.View() != first {
		t.Error("spinner reacted to another spinner's tick")
	}
	s3, cmd := s.Update(TickMsg{Time: time.Now(), id: s.id, tag: s.tag + 1})
	if cmd != nil || s3.View() != first {
		t.Error("spinner reacted to a stale tick")
	}
}

func TestSpinnerGolden(t *testing.T) {
	s := New(tuiweave.Dark())
	snaptest.Snap(t, s.View())
	snaptest.SnapCells(t, s.View(), snaptest.WithRoles(tuiweave.Dark()))
}

func TestSetFrames(t *testing.T) {
	s := New(tuiweave.Dark())
	s.SetFrames("-", "\\", "|", "/")
	if got := s.View(); got == "" || len([]rune(got)) == 0 {
		t.Fatal("empty view after SetFrames")
	}
	s2, _ := s.Update(TickMsg{id: s.id, tag: s.tag})
	if s2.View() == s.View() {
		t.Error("frame did not advance after SetFrames")
	}
}

func TestSpinnerIsIntrinsic(t *testing.T) {
	if got := New(tuiweave.Dark()).SizeMode(); got != layout.SizeIntrinsic {
		t.Fatalf("spinner size mode = %d, want SizeIntrinsic", got)
	}
}
