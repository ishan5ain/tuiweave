# inspect

`inspect` describes a composed UI as data for tests, tools, and agents, and carries semantic action intents.

```go
root := inspect.Group("app", "application", inspect.Bounds{Width: w, Height: h},
	inspect.Bind("list", m.list),
	inspect.BindAt("input", inspect.FromRect(inputArea), m.input),
)
data, err := inspect.Marshal(root)
```

The application owns stable IDs, visibility, absolute bounds, privacy, and routing. Bound component actions are qualified by node ID; route the target in the app, then deliver the local suffix with `inspect.Invoke`.

See [agent-operable surfaces](../AGENT-CATALOG.md#agent-operable-surfaces).

## API highlights

- `Group`, `Bind`, and `BindAt` build an application-owned semantic tree.
- `Marshal` encodes inspection data.
- `Invoke` creates a local semantic action message.
- Components supply `Inspect()` and `Actions()`; the app supplies IDs and bounds.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/inspect)
- [Agent-operable surfaces](../AGENT-CATALOG.md#agent-operable-surfaces)
