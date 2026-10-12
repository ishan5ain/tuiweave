# palette

`palette` combines query editing with filtered discovery and activation of application-owned actions.

```go
p := palette.New(theme)
p.SetItems(items...)
p.SetSize(width, height)
p.Focus()
p, cmd = p.Update(msg)

// Handle palette.SelectedMsg in the application.
```

Filtering is case-insensitive over action ID, label, and description. Disabled results remain visible and are skipped. The selected index refers to the original action set; visibility and side effects remain application-owned.

See the [palette example](../examples/palette) and [palette recipe](../AGENTS.md#palette).

## API highlights

- `SetItems` supplies shared `action.Item` definitions.
- `SetQuery` controls case-insensitive filtering.
- `SelectedID` and `SelectedMsg` expose the current/activated action.
- `Focus`, `Blur`, `SetSize`, `Update`, and `View` make up the bounded loop.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/palette)
- [Palette example](../examples/palette)
- [Palette recipe](../AGENTS.md#palette)

## Terminal paste

Forward `tea.PasteMsg` to a visible palette alongside key messages. Its focused
query input accepts normalized paste and refreshes filtering and selection;
paste never activates a command. Hidden palettes must remain blurred or be
excluded from application routing, and the top modal owns all input.
