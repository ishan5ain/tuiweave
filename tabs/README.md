# tabs

`tabs` renders focusable navigation among sibling, application-owned views.

```go
nav := tabs.New(theme)
nav.SetTabs(
	tabs.Tab{ID: "overview", Label: "Overview"},
	tabs.Tab{ID: "logs", Label: "Logs"},
)
nav.SetSize(width, 1)
nav.Focus()
```

Stable IDs preserve selection across `SetTabs` updates. Arrow and vim-style navigation are focus-gated; Tab remains an application-global focus key. Render the selected body separately using `SelectedID()`.

To let the unselected strip inherit its terminal or parent background, pass a
derived theme to `New`:

```go
theme := tuiweave.Dark()
theme.SurfaceRaised = lipgloss.NoColor{}
nav := tabs.New(theme)
```

Focused selection still uses `SelectionBg` and `SelectionFg`; only the raised
bar surface stops emitting a background color.

See the [frame example](../examples/frame) and [tabs recipe](../AGENTS.md#tabs).

## API highlights

- `SetTabs` replaces tabs while preserving selection by stable ID.
- `SelectedID` returns the application-owned view key.
- `Focus`, `Blur`, `Update`, and `View` implement sibling navigation.
- `Tab` defines each stable ID and display label.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/tabs)
- [Frame example](../examples/frame)
- [Tabs recipe](../AGENTS.md#tabs)
