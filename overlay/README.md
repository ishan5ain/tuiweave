# overlay

`overlay` composites a modal or popover over an already-rendered base view in terminal cell space.

```go
view := overlay.Center(baseView, dialogView)
// Or place at an application-owned origin:
view = overlay.Place(baseView, popoverView, x, y)
```

The application owns visibility, layout, focus, and message routing. Composition preserves ANSI styles, wide glyphs, and combining graphemes; content outside the base bounds is clipped.

See the [operations example](../examples/ops) and [modal routing conventions](../AGENT-CATALOG.md#canonical-wiring).

## API highlights

- `Place` composites an overlay at an app-owned origin.
- `Center` centers an overlay over the base view.
- Both helpers return strings and leave visibility and routing to the app.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/overlay)
- [Modal routing](../AGENT-CATALOG.md#canonical-wiring)
