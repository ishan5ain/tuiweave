// Package layout partitions terminal space into rectangles using flexbox-like
// constraints, and sizes tuiweave components from the result.
//
// It is a thin facade over ultraviolet's constraint solver: consumers depend
// on this package, never on ultraviolet directly. The typical pattern, run on
// every tea.WindowSizeMsg:
//
//	var header, body, status layout.Rect
//	layout.Vertical(
//		layout.Len(3),   // fixed header
//		layout.Fill(1),  // body takes the rest
//		layout.Len(1),   // one-line statusbar
//	).Split(layout.Rect{Max: image.Pt(msg.Width, msg.Height)}).
//		Assign(&header, &body, &status)
//
//	m.list.SetSize(body.Dx(), body.Dy())
//
// or, when components implement Sizable, in one step:
//
//	layout.Vertical(layout.Len(3), layout.Fill(1), layout.Len(1)).
//		Apply(layout.NewRect(0, 0, msg.Width, msg.Height), &m.header, &m.list, &m.status)
package layout

import (
	"fmt"
	"image"

	uvlayout "github.com/charmbracelet/ultraviolet/layout"
)

// Rect is a rectangular screen region. It aliases image.Rectangle, so the
// usual Dx/Dy/Intersect methods apply.
type Rect = image.Rectangle

// NewRect returns the rectangle at origin x, y with the given size.
func NewRect(x, y, w, h int) Rect {
	return Rect{Min: image.Pt(x, y), Max: image.Pt(x+w, y+h)}
}

// Constraint describes how one segment of a split should be sized.
// Constraints are produced by Len, Min, Max, Percent, Ratio, and Fill. The
// interface is sealed so layout can preserve its representation independently
// of the internal solver.
type Constraint interface {
	isConstraint()
}

type constraintKind uint8

const (
	constraintLen constraintKind = iota
	constraintMin
	constraintMax
	constraintPercent
	constraintRatio
	constraintFill
)

type constraint struct {
	kind  constraintKind
	value int
	den   int
}

func (constraint) isConstraint() {}

func (c constraint) String() string {
	switch c.kind {
	case constraintLen:
		return fmt.Sprintf("Len(%d)", c.value)
	case constraintMin:
		return fmt.Sprintf("Min(%d)", c.value)
	case constraintMax:
		return fmt.Sprintf("Max(%d)", c.value)
	case constraintPercent:
		return fmt.Sprintf("Percent(%d)", c.value)
	case constraintRatio:
		return fmt.Sprintf("Ratio(%d / %d)", c.value, c.den)
	case constraintFill:
		return fmt.Sprintf("Fill(%d)", c.value)
	default:
		return "Constraint(?)"
	}
}

// Len fixes a segment to exactly n cells.
func Len(n int) Constraint { return constraint{kind: constraintLen, value: n} }

// Min gives a segment at least n cells.
func Min(n int) Constraint { return constraint{kind: constraintMin, value: n} }

// Max caps a segment at n cells.
func Max(n int) Constraint { return constraint{kind: constraintMax, value: n} }

// Percent sizes a segment as a percentage (0–100) of the total area.
func Percent(p int) Constraint { return constraint{kind: constraintPercent, value: p} }

// Ratio sizes a segment as num/den of the total area.
func Ratio(num, den int) Constraint {
	return constraint{kind: constraintRatio, value: num, den: den}
}

// Fill distributes leftover space among Fill segments proportionally to
// weight, like flex-grow.
func Fill(weight int) Constraint { return constraint{kind: constraintFill, value: weight} }

func toUltravioletConstraints(constraints []Constraint) []uvlayout.Constraint {
	converted := make([]uvlayout.Constraint, len(constraints))
	for i, c := range constraints {
		converted[i] = toUltravioletConstraint(c)
	}
	return converted
}

func toUltravioletConstraint(c Constraint) uvlayout.Constraint {
	if c == nil {
		return nil
	}

	owned, ok := c.(constraint)
	if !ok {
		panic("layout: unsupported constraint implementation")
	}

	switch owned.kind {
	case constraintLen:
		return uvlayout.Len(owned.value)
	case constraintMin:
		return uvlayout.Min(owned.value)
	case constraintMax:
		return uvlayout.Max(owned.value)
	case constraintPercent:
		return uvlayout.Percent(owned.value)
	case constraintRatio:
		return uvlayout.Ratio{Num: owned.value, Den: owned.den}
	case constraintFill:
		return uvlayout.Fill(owned.value)
	default:
		panic("layout: unknown constraint kind")
	}
}

// Layout splits an area into segments along one direction.
type Layout struct {
	inner uvlayout.Layout
}

// Vertical returns a Layout that stacks segments top to bottom.
func Vertical(constraints ...Constraint) Layout {
	return Layout{inner: uvlayout.Vertical(toUltravioletConstraints(constraints)...)}
}

// Horizontal returns a Layout that arranges segments left to right.
func Horizontal(constraints ...Constraint) Layout {
	return Layout{inner: uvlayout.Horizontal(toUltravioletConstraints(constraints)...)}
}

// WithSpacing sets the gap, in cells, between adjacent segments.
func (l Layout) WithSpacing(n int) Layout {
	l.inner.Spacing = n
	return l
}

// WithPadding insets the area by top, right, bottom, left before splitting,
// following CSS shorthand: Pad(all), Pad(vertical, horizontal), or
// Pad(top, right, bottom, left).
func (l Layout) WithPadding(sides ...int) Layout {
	l.inner.Padding = uvlayout.Pad(sides...)
	return l
}

// Split partitions area into one rectangle per constraint.
func (l Layout) Split(area Rect) Splitted {
	return Splitted(l.inner.Split(area))
}

// Splitted holds the rectangles produced by Split, in constraint order.
type Splitted []Rect

// Assign copies the rectangles into the given pointers, in order. Extra
// pointers are left untouched; extra rectangles are dropped.
func (s Splitted) Assign(areas ...*Rect) {
	uvlayout.Splitted(s).Assign(areas...)
}

// Sizable is implemented by every tuiweave component: SetSize supplies the box
// for bounded components and the available constraints for documented
// intrinsic or width-bounded components.
type Sizable interface {
	SetSize(width, height int)
}

// SizeMode describes how a component interprets the box supplied by SetSize.
// Components are bounded by default; only documented exceptions implement
// SizeModeAware.
type SizeMode uint8

const (
	// SizeBounded means the component renders exactly within its assigned box.
	SizeBounded SizeMode = iota
	// SizeWidthBounded means width is constrained but height is natural or
	// otherwise component-defined.
	SizeWidthBounded
	// SizeIntrinsic means SetSize is accepted for interface compatibility but
	// the component renders at its natural size.
	SizeIntrinsic
)

// SizeModeAware is an optional contract for components that are not fully
// bounded. Applications can use SizeModeOf when composing mixed components.
type SizeModeAware interface {
	SizeMode() SizeMode
}

// SizeModeOf returns a component's declared size mode, defaulting to
// SizeBounded for ordinary components.
func SizeModeOf(component any) SizeMode {
	if aware, ok := component.(SizeModeAware); ok {
		return aware.SizeMode()
	}
	return SizeBounded
}

// Apply splits area and sizes each component from its rectangle, in order.
// It returns the rectangles for callers that also need positions (e.g. to
// join views or place overlays). Extra components are left untouched.
func (l Layout) Apply(area Rect, components ...Sizable) Splitted {
	rects := l.Split(area)
	for i, c := range components {
		if i >= len(rects) {
			break
		}
		c.SetSize(rects[i].Dx(), rects[i].Dy())
	}
	return rects
}
