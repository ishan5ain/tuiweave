package mouse

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestWheelDelta(t *testing.T) {
	tests := []struct {
		name  string
		msg   tea.Msg
		want  int
		found bool
	}{
		{name: "up", msg: tea.MouseWheelMsg{Button: tea.MouseWheelUp}, want: -WheelLines, found: true},
		{name: "down", msg: tea.MouseWheelMsg{Button: tea.MouseWheelDown}, want: WheelLines, found: true},
		{name: "horizontal left", msg: tea.MouseWheelMsg{Button: tea.MouseWheelLeft}},
		{name: "horizontal right", msg: tea.MouseWheelMsg{Button: tea.MouseWheelRight}},
		{name: "other message", msg: tea.KeyPressMsg{Code: 'j'}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, found := WheelDelta(test.msg)
			if got != test.want || found != test.found {
				t.Fatalf("WheelDelta() = (%d, %v), want (%d, %v)", got, found, test.want, test.found)
			}
		})
	}
}

func TestPosition(t *testing.T) {
	msg := tea.MouseClickMsg{X: 12, Y: 7, Button: tea.MouseLeft}
	x, y, ok := Position(msg)
	if x != 12 || y != 7 || !ok {
		t.Fatalf("Position() = (%d, %d, %v), want (12, 7, true)", x, y, ok)
	}

	if _, _, ok := Position(tea.KeyPressMsg{Code: 'j'}); ok {
		t.Fatal("Position() recognized a non-mouse message")
	}
}

func TestInBoundsUsesHalfOpenCoordinates(t *testing.T) {
	inside := tea.MouseClickMsg{X: 10, Y: 4, Button: tea.MouseLeft}
	if !InBounds(inside, 10, 4, 8, 3) {
		t.Fatal("point at the top-left corner should be inside")
	}

	for name, msg := range map[string]tea.Msg{
		"right edge":  tea.MouseClickMsg{X: 18, Y: 5, Button: tea.MouseLeft},
		"bottom edge": tea.MouseClickMsg{X: 12, Y: 7, Button: tea.MouseLeft},
		"outside":     tea.MouseClickMsg{X: 4, Y: 2, Button: tea.MouseLeft},
		"not mouse":   tea.KeyPressMsg{Code: 'j'},
	} {
		if InBounds(msg, 10, 4, 8, 3) {
			t.Errorf("%s should be outside", name)
		}
	}

	if InBounds(inside, 10, 4, 0, 3) {
		t.Error("zero-width boxes should not contain mouse events")
	}
}
