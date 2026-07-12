# snaptest

`snaptest` provides deterministic golden tests for rendered strings and explicit MVU interaction scenarios.

```go
w := widget.New(tuiweave.Dark())
w.SetSize(40, 5)
snaptest.Snap(t, w.View())
snaptest.SnapCells(t, w.View(), snaptest.WithRoles(tuiweave.Dark()))
```

Use `Snap` for readable layout/content goldens and `SnapCells` for theme-role style runs. `RunScenario` records explicit messages and whether commands were emitted; it does not execute commands. Update intentional visual changes with `go test ./... -update`, then read every golden diff.

See [testing a component](../AGENTS.md#testing-a-component).

## API highlights

- `Snap` captures readable plain-text goldens.
- `SnapCells` captures role-labelled style runs.
- `SnapStyled` is available for rare raw-ANSI assertions.
- `RunScenario` and `SnapScenario` capture named interaction checkpoints.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/snaptest)
- [Testing a component](../AGENTS.md#testing-a-component)
