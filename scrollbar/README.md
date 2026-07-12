# scrollbar

`scrollbar` renders a standalone one-column indicator for any component implementing `Scrollable`.

```go
m.list.SetSize(pane.Dx(), pane.Dy())
view := lipgloss.JoinHorizontal(
	lipgloss.Top,
	m.list.View(),
	scrollbar.For(theme, m.list),
)
```

Allocate the bar its own one-cell layout segment; scrolling components do not render bars internally. Table row content begins two lines below its header, so prepend two blank rows when aligning a table scrollbar.

See the [table example](../examples/table) and [scrollbar recipe](../AGENTS.md#scrollbar).

## API highlights

- `For(theme, component)` derives the bar from a `Scrollable` component.
- `Vertical` renders a bar from explicit total, visible, and offset values.
- Allocate the bar in a one-column layout segment.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/scrollbar)
- [Scrollbar recipe](../AGENTS.md#scrollbar)
