# dialog

`dialog` renders a width-bounded confirmation panel and emits a typed result.

```go
d := dialog.New(theme)
d.Title = "Delete workspace?"
d.Body = "This cannot be undone."
d.SetSize(width, height)
d, cmd = d.Update(msg)

// Handle dialog.ResultMsg in the application.
```

The dialog renders at natural content height and reports `layout.SizeWidthBounded`. The application owns visibility, overlay composition, modal focus, and routing keys exclusively to the active layer.

See the [operations example](../examples/ops) and [modal compatibility notes](../AGENT-CATALOG.md#api-compatibility-notes).

## API highlights

- `Title`, `Body`, `ID`, and button labels configure the prompt.
- `ResultMsg` reports `ID` and whether the user confirmed.
- `SetSize`, `Update`, and `View` implement the width-bounded modal contract.
- `ActionConfirm`, `ActionCancel`, `ActionNext`, and `ActionPrevious` support semantic control.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/dialog)
- [Operations example](../examples/ops)
- [Modal compatibility notes](../AGENT-CATALOG.md#api-compatibility-notes)
