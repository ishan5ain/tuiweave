# viewport

`viewport` scrolls pre-rendered, optionally styled content inside an exact-size box.

```go
vp := viewport.New(theme)
vp.SetContent(content)
vp.SetSize(width, height)
vp, cmd = vp.Update(msg)
```

Keyboard navigation is focus-gated, but mouse-wheel scrolling works while blurred. Always forward wheel messages; use application-owned bounds when multiple scrollable panes are present. Add `scrollbar.For` in a separate layout column.

See the [browser example](../examples/browser) and [mouse conventions](../AGENTS.md#mouse-input).

## API highlights

- `SetContent` replaces the pre-rendered source.
- `SetSize` establishes the visible window.
- `YOffset`, `GotoTop`, and `GotoBottom` expose local scroll state.
- `Update` handles keyboard navigation and wheel messages.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/viewport)
- [Browser example](../examples/browser)
- [Mouse input](../AGENTS.md#mouse-input)
