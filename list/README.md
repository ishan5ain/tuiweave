# list

`list` displays and navigates a bounded set of one-line items.

```go
l := list.New(theme)
l.SetItems("api", "worker", "database")
l.SetSize(width, height)
l.Focus()
l, cmd = l.Update(msg)
```

Keyboard navigation is focus-gated. `SetFilter` performs case-insensitive substring filtering, while `Selected()` remains an index into the original item set. Add a standalone `scrollbar.For` when needed.

See the [browser example](../examples/browser) and [scrolling recipe](../AGENTS.md#viewport--list--table-scrolling-components).

## API highlights

- `SetItems` replaces the original item collection.
- `SetFilter` applies case-insensitive substring filtering.
- `Selected` reports the original-items index.
- `SelectedItem`, `FilteredLen`, and scrolling accessors expose local state.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/list)
- [Browser example](../examples/browser)
- [Scrolling components](../AGENTS.md#viewport--list--table-scrolling-components)
