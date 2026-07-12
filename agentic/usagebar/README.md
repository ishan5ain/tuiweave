# agentic/usagebar

`usagebar` renders agent-session model, token, cost, and context usage as a passive status line.

```go
u := usagebar.New(theme)
u.SetStats(usagebar.Stats{
	Model: "model-name", TokensIn: 1200, TokensOut: 450,
	Cost: 0.08, ContextUsed: 0.72,
})
u.SetSize(width, 1)
```

Context usage switches to warning at 80% and danger at 95%. The application owns measurement and refresh timing. Reconstruct the component after a theme change and restore the latest `Stats` and dimensions.

See the [chat example](../../examples/chat) and [agentic package map](../../AGENT-CATALOG.md#agentic-domain-components).

## API highlights

- `Stats` holds model, token, cost, and context measurements.
- `SetStats` refreshes all displayed segments and semantic thresholds.
- `Stats` returns the last application-provided snapshot.
- `SetSize`, `Update`, and `View` expose passive statusbar behavior.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/agentic/usagebar)
- [Chat example](../../examples/chat)
- [Agentic package map](../../AGENT-CATALOG.md#agentic-domain-components)
