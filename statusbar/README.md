# statusbar

`statusbar` renders passive, one-line status information in themed left and right segments.

```go
bar := statusbar.New(theme)
bar.SetSize(width, 1)
bar.SetLeft(statusbar.Segment{Text: "ready", Kind: statusbar.KindSuccess})
bar.SetRight(statusbar.Segment{Text: "12:4", Kind: statusbar.KindMuted})
```

Use `layout.Len(1)`. At narrow widths the bar drops right segments before truncating the left. Reconstruct it after a theme change and restore application-owned segments and dimensions.

See the [statusbar example](../examples/statusbar) and [recipe](../AGENTS.md#statusbar).

## API highlights

- `New(theme)` constructs a themed passive bar.
- `SetLeft` and `SetRight` replace each side's segments.
- `Segment` and `Kind` select semantic status styling.
- `SetSize` controls the one-row rendering box.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/statusbar)
- [Statusbar recipe](../AGENTS.md#statusbar)
