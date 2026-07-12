# line

`line` composes exact-width terminal rows while preserving ANSI styling and visible cell widths.

```go
title := line.Fit("service overview", width, line.AlignLeft)
footer := line.Join(width, title, "q quit", line.JoinOptions{Gap: 1})
rule := line.Fill(width, "─")
```

`Truncate` adds an ellipsis, `Fit` also pads and aligns, and `Join` reserves the right side before fitting the left. Use `NoEllipsis` for decorative fills or rules where an ellipsis would be misleading.

See the [frame example](../examples/frame) and [line recipe](../AGENTS.md#line).

## API highlights

- `Truncate` fits text with an ellipsis.
- `Fit` truncates, aligns, and pads to an exact width.
- `Fill` repeats a visible pattern without splitting glyphs.
- `Join` reserves a right-side value and fills the remaining row.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/line)
- [Line recipe](../AGENTS.md#line)
