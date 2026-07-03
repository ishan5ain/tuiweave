// Package overlay composites one rendered view on top of another — the
// mechanism behind modals, popovers, and toasts.
//
// Compositing happens in terminal cell space (via ultraviolet buffers, an
// internal detail), so the overlay cleanly replaces the cells beneath it,
// styling included — no string-splicing artifacts.
package overlay

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// Place draws over on top of base with over's top-left corner at column x,
// row y (cell coordinates, 0-based), and returns the composed view. The
// result has base's dimensions; parts of over outside base are clipped.
func Place(base, over string, x, y int) string {
	bs := uv.NewStyledString(base)
	bounds := bs.Bounds()
	buf := uv.NewScreenBuffer(bounds.Dx(), bounds.Dy())
	bs.Draw(buf, buf.Bounds())

	os := uv.NewStyledString(over)
	ob := os.Bounds()
	os.Draw(buf, uv.Rect(x, y, ob.Dx(), ob.Dy()))

	return buf.Render()
}

// Center draws over centered on base and returns the composed view.
func Center(base, over string) string {
	bs := uv.NewStyledString(base)
	ob := uv.NewStyledString(over).Bounds()
	bb := bs.Bounds()
	x := max(0, (bb.Dx()-ob.Dx())/2)
	y := max(0, (bb.Dy()-ob.Dy())/2)
	return Place(base, over, x, y)
}
