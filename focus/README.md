# focus

`focus` provides copy-safe tab order, modal scopes, and nested focus stacks for MVU applications.

```go
fm := focus.NewManager(3)
fm.Next()
fm.Apply(&m.list, &m.preview, &m.input)
```

Managers store indices, never component pointers. Apply fresh component addresses after every focus change. Use `Scope` for one conditional modal and `Stack` for nested layers; the application still owns key routing and visibility.

See the [operations example](../examples/ops) and [focus recipe](../AGENTS.md#focus-tab-order).

## API highlights

- `NewManager` creates a value-type tab-order manager.
- `Scope` isolates a conditional modal group.
- `Stack` manages nested focus layers.
- `Apply` pushes current focus to fresh component addresses.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/focus)
- [Focus recipe](../AGENTS.md#focus-tab-order)
