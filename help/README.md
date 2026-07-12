# help

`help` renders a passive, one-line collection of key bindings.

```go
h := help.New(theme)
h.SetBindings(
	help.Binding{Key: "tab", Desc: "focus"},
	help.Binding{Key: "q", Desc: "quit"},
)
h.SetSize(width, 1)
```

Hints are kept whole and dropped from the right as space narrows. The component handles no messages; the application owns the actual key behavior.

See the [demo](../examples/demo) and [package catalog](../AGENT-CATALOG.md#general-purpose-components).

## API highlights

- `New(theme)` constructs a passive hint bar.
- `SetBindings` replaces the displayed `Binding` values.
- `SetSize` controls the available one-row width.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/help)
- [General-purpose components](../AGENT-CATALOG.md#general-purpose-components)
