// Package mouse contains small, application-owned helpers for Bubble Tea
// mouse messages.
//
// Components still own the behavior they can determine from a message alone
// (for example, viewport scrolling). Applications own hit-testing and click
// routing because only the application knows where each component was laid
// out.
package mouse

import tea "charm.land/bubbletea/v2"

// WheelLines is the conventional number of lines moved by one vertical wheel
// event. Scrollable components should use this value unless they document a
// deliberate alternative.
const WheelLines = 3

// WheelDelta returns the standard vertical scroll delta for msg. A negative
// value means scroll up; a positive value means scroll down. Horizontal wheel
// events are left to the application because the generic tuiweave scrollable
// contract is vertical.
func WheelDelta(msg tea.Msg) (delta int, ok bool) {
	event, ok := msg.(tea.MouseWheelMsg)
	if !ok {
		return 0, false
	}
	switch event.Button {
	case tea.MouseWheelUp:
		return -WheelLines, true
	case tea.MouseWheelDown:
		return WheelLines, true
	default:
		return 0, false
	}
}

// Position returns the zero-based terminal coordinates carried by any
// Bubble Tea mouse message. It recognizes clicks, releases, wheel events, and
// motion events through tea.MouseMsg.
func Position(msg tea.Msg) (x, y int, ok bool) {
	event, ok := msg.(tea.MouseMsg)
	if !ok {
		return 0, 0, false
	}

	point := event.Mouse()
	return point.X, point.Y, true
}

// InBounds reports whether a mouse message falls inside the half-open box
// [x, x+width) × [y, y+height). Use this in application routing after layout
// has produced a component's terminal rectangle.
func InBounds(msg tea.Msg, x, y, width, height int) bool {
	if width <= 0 || height <= 0 {
		return false
	}

	px, py, ok := Position(msg)
	return ok && px >= x && px < x+width && py >= y && py < y+height
}
