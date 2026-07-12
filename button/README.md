# button

`button` is a focusable target for one application-owned action.

```go
b := button.New(theme)
b.ID = "open-workspace"
b.SetLabel("Open workspace")
b.SetSize(width, 1)
b.Focus()
b, cmd = b.Update(msg)
```

Enter, space, and semantic activation emit `button.PressedMsg`. Disabled buttons remain visible but reject focus and activation. Use `menu` or `toolbar` when several actions belong together.

See the [frame example](../examples/frame) and [button recipe](../AGENTS.md#button).

## API highlights

- `SetLabel` and the exported `ID` identify the action target.
- `PressedMsg` reports Enter/space activation.
- `Focus`, `Blur`, `Update`, and `View` implement the local control loop.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/button)
- [Frame example](../examples/frame)
- [Button recipe](../AGENTS.md#button)
