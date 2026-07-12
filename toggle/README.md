# toggle

`toggle` is a focusable boolean setting whose value remains application-owned.

```go
t := toggle.New(theme)
t.ID = "auto-refresh"
t.SetLabel("Auto-refresh")
t.SetChecked(true)
t.SetSize(width, 1)
t.Focus()
```

Space, Enter, `x`, and semantic actions change the value and emit `toggle.ChangedMsg`. The application owns persistence and side effects. Disabled toggles remain visible but reject focus and changes.

See the [frame example](../examples/frame) and [toggle recipe](../AGENTS.md#toggle).

## API highlights

- `SetLabel`, `SetChecked`, and `Checked` manage visible setting state.
- `ChangedMsg` reports a changed value to the application.
- `Focus`, `Blur`, `Update`, and `View` implement the local control loop.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/toggle)
- [Frame example](../examples/frame)
- [Toggle recipe](../AGENTS.md#toggle)
