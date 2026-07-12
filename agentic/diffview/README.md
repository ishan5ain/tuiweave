# agentic/diffview

`diffview` renders unified diffs either inline or inside a scrollable component.

```go
inline := diffview.Sprint(theme, diff, width)

pane := diffview.New(theme)
pane.SetDiff(diff)
pane.SetSize(width, height)
pane, cmd = pane.Update(msg)
```

Use `Sprint` for a chat cell or other natural-height composition and `Model` for a bounded pane. The model follows viewport focus and mouse-wheel behavior and can be paired with `scrollbar.For`.

See the [table example](../../examples/table) and [chat example](../../examples/chat).

## API highlights

- `Sprint` renders a diff inline at an explicit width.
- `New` creates a scrollable diff model.
- `SetDiff` replaces the source content.
- `TotalLines`, `VisibleLines`, and `YOffset` expose scroll statistics.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/agentic/diffview)
- [Table example](../../examples/table)
- [Chat example](../../examples/chat)
