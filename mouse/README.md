# mouse

`mouse` normalizes Bubble Tea mouse input and helps applications route events using their layout rectangles.

```go
if mouse.InBounds(msg, area.Min.X, area.Min.Y, area.Dx(), area.Dy()) {
	m.preview, cmd = m.preview.Update(msg)
}
```

Use `WheelDelta` for the shared three-line wheel convention and `Position` for app-owned click handling. Components render local strings and must not guess global coordinates. Forward wheel events to scrollable components even when they are blurred.

See the [browser example](../examples/browser) and [mouse conventions](../AGENTS.md#mouse-input).

## API highlights

- `WheelDelta` normalizes vertical wheel messages.
- `Position` extracts app-owned pointer coordinates.
- `InBounds` tests a pointer against a layout rectangle.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/mouse)
- [Mouse input](../AGENTS.md#mouse-input)
