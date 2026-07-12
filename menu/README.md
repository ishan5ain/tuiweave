# menu

`menu` presents a vertical, focusable set of application-owned actions.

```go
m := menu.New(theme)
m.SetItems(items...)
m.SetSize(width, height)
m.Focus()
m, cmd = m.Update(msg)

// Handle menu.SelectedMsg in the application.
```

Items use shared `action.Item` definitions. Disabled actions remain visible and are skipped. Enter emits a typed command, so always reassign the model and collect the returned `tea.Cmd`.

See the [frame example](../examples/frame) and [menu recipe](../AGENTS.md#menu).

## API highlights

- `SetItems` accepts shared `action.Item` definitions.
- `SelectedID` exposes the current stable action ID.
- `SelectedMsg` reports Enter activation to the application.
- `Focus`, `Blur`, `Update`, and `View` manage the local interaction loop.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/menu)
- [Frame example](../examples/frame)
- [Menu recipe](../AGENTS.md#menu)
