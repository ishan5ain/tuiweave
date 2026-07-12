# splitpane

`splitpane` composes two width-aware sibling views with optional themed separation and natural-height alignment.

```go
view := splitpane.Horizontal(theme, width, splitpane.Options{Ratio: 60, Gap: 1},
	func(w int) string { return renderList(w) },
	func(w int) string { return renderPreview(w) },
)
```

Callbacks receive their assigned widths; the application still owns component state, sizing, focus, and message routing. `Gap: 0` uses a one-cell divider, while a negative gap removes it.

See the [frame example](../examples/frame) and [splitpane recipe](../AGENTS.md#splitpane).

## API highlights

- `Horizontal` splits a width between two callbacks.
- `Options.Ratio` controls the left-pane percentage.
- `Options.Gap` selects divider, spacing, or no divider behavior.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/splitpane)
- [Splitpane recipe](../AGENTS.md#splitpane)
