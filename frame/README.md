# frame

`frame` supplies themed, domain-neutral decoration around application-owned strings.

```go
opts := frame.PanelOptions{Title: "Jobs", Focused: focused, Padding: 1}
inner := frame.PanelContentRect(area, opts)
m.list.SetSize(inner.Dx(), inner.Dy())
view := frame.Panel(theme, m.list.View(), area.Dx(), opts)
```

`Panel`, `Divider`, and `Badge` are pure composition helpers, not MVU components. They receive explicit widths and return natural-height strings. Choose badge kinds by meaning and let all styling come from theme roles.

To let panel bodies inherit their terminal or parent background, derive the
theme before rendering:

```go
theme := tuiweave.Dark()
theme.SurfaceRaised = lipgloss.NoColor{}
view := frame.Panel(theme, content, width, opts)
```

The border and title keep their semantic foreground roles. Badges remain
filled by their accent or intent role; their fill communicates meaning rather
than serving as a panel background.

See the [frame example](../examples/frame) and [frame recipe](../AGENTS.md#frame).

## API highlights

- `Panel` returns a themed, exact-width natural-height frame.
- `PanelContentRect` calculates the child rectangle for a panel.
- `Divider` and `Badge` provide semantic composition chrome.
- `PanelOptions` controls title, focus border, and padding.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/frame)
- [Frame recipe](../AGENTS.md#frame)
