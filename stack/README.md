# stack

`stack` composes width-aware headers, body sections, separators, and footers into an exact-width, natural-height view.

```go
view := stack.Vertical(theme, width, stack.Options{Divider: true},
	func(w int) string { return headerView(w) },
	func(w int) string { return bodyView(w) },
	func(w int) string { return footerView(w) },
)
```

Empty sections are omitted. Use `Gap` for blank rows and `Divider` for themed rules. The application remains responsible for rectangles, component state, and message routing.

See the [frame example](../examples/frame) and [stack recipe](../AGENTS.md#stack).

## API highlights

- `Vertical` composes width-aware section callbacks.
- `Options.Divider` inserts themed rules between non-empty sections.
- `Options.Gap` inserts width-filled blank rows.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/stack)
- [Stack recipe](../AGENTS.md#stack)
