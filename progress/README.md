# progress

`progress` renders a passive, exact-width task completion indicator.

```go
p := progress.New(theme)
p.SetSize(width, 1)
p.SetLabel("Indexing")
p.SetPercent(0.72)
p.SetStatus(progress.StatusInfo)
```

Percentages are clamped to `[0, 1]`. Choose `Normal`, `Success`, `Warning`, `Danger`, or `Info` by meaning. Narrow rendering gives label and percentage space back to the bar; do not pre-truncate them.

See the [frame example](../examples/frame) and [progress recipe](../AGENTS.md#progress).

## API highlights

- `SetLabel` updates the task label.
- `SetPercent` clamps completion to `[0, 1]`.
- `SetStatus` selects semantic intent styling.
- `SetSize` and `View` implement exact-width passive rendering.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/progress)
- [Progress recipe](../AGENTS.md#progress)
