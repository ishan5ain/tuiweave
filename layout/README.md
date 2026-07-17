# layout

`layout` partitions terminal space into rectangles and sizes components from those rectangles.

```go
layout.Vertical(
	layout.Len(3),
	layout.Fill(1),
	layout.Len(1),
).Apply(layout.NewRect(0, 0, msg.Width, msg.Height),
	&m.header, &m.body, &m.status)
```

Run layout on every `tea.WindowSizeMsg`. Bounded components render inside the box supplied through `SetSize`; use `SizeModeOf` when composing intrinsic or width-bounded exceptions. Applications should use `layout.Rect` and never import Ultraviolet directly.

See the [demo](../examples/demo) and [layout conventions](../AGENTS.md#wiring-an-app-the-only-layout-pattern).

## API highlights

- `Vertical` and `Horizontal` create split layouts.
- `Len`, `Min`, `Max`, `Percent`, `Ratio`, and `Fill` create constraints.
- `Constraint` is a tuiweave-owned sealed interface; the solver adapter remains
  internal to this package.
- `Apply` assigns rectangles and sizes `layout.Sizable` components.
- `SizeModeOf` reports bounded, width-bounded, or intrinsic sizing.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/layout)
- [Canonical wiring](../AGENTS.md#wiring-an-app-the-only-layout-pattern)
- [Compatibility and layout migration policy](../COMPATIBILITY.md#layout-compatibility)
