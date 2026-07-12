# toolbar

`toolbar` presents a horizontal, focusable strip of application-owned actions.

```go
t := toolbar.New(theme)
t.SetItems(items...)
t.SetSize(width, 1)
t.Focus()
t, cmd = t.Update(msg)

// Handle toolbar.SelectedMsg in the application.
```

Disabled actions remain visible and are skipped. The selected item stays visible at narrow widths. The toolbar emits activation intent only; the application owns command side effects.

See the [frame example](../examples/frame) and [toolbar recipe](../AGENTS.md#toolbar).

## API highlights

- `SetItems` accepts shared `action.Item` definitions.
- `SelectedID` exposes the current stable action ID.
- `SelectedMsg` reports activation to the application.
- `Focus`, `Blur`, `Update`, and `View` manage the one-row interaction loop.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/toolbar)
- [Frame example](../examples/frame)
- [Toolbar recipe](../AGENTS.md#toolbar)
