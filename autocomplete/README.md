# autocomplete

`autocomplete` is a bounded suggestion window placed beside an application-owned input.

```go
ac := autocomplete.New(theme)
ac.SetItems(autocomplete.Item{
	ID: "git-status", Value: "git status", Label: "git status",
})
ac.SetQuery(input.Value())
ac.SetSize(width, height)
```

The application synchronizes the query and applies `autocomplete.SelectedMsg.Value` to its input. Matching uses value/label prefixes with description substring fallback. Disabled suggestions remain visible and are skipped.

See the [autocomplete example](../examples/autocomplete) and [recipe](../AGENTS.md#autocomplete).

## API highlights

- `SetItems` supplies `Item` candidates with stable IDs and inserted values.
- `SetQuery` synchronizes the app-owned input text.
- `SelectedMsg` reports ID, original index, label, and inserted value.
- `Focus`, `Blur`, `SetSize`, `Update`, and `View` manage the suggestion window.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/autocomplete)
- [Autocomplete example](../examples/autocomplete)
- [Autocomplete recipe](../AGENTS.md#autocomplete)
