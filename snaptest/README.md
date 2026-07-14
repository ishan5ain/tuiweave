# snaptest tutorial

`snaptest` turns rendered strings and explicit MVU transitions into deterministic,
reviewable golden files. The compile-checked fixture in
[`tutorial_test.go`](tutorial_test.go) uses only public APIs and demonstrates the
complete workflow.

## API highlights

- `Snap` captures readable plain-text layout and content.
- `SnapCells` captures terminal-cell style runs; `WithRoles` labels theme colors.
- `SnapStyled` captures raw ANSI bytes for the rare cases that require them.
- `RunScenario` and `SnapScenario` capture named MVU checkpoints without
  executing commands.

## 1. Build a deterministic fixture

Give every bounded component a fixed size and explicit focus state. Keep
application-owned state outside the component, and preserve the MVU contract by
reassigning every model returned by `Update` and returning its command:

```go
func newTutorialModel() tutorialModel {
	setting := toggle.New(tuiweave.Dark())
	setting.ID = "auto-refresh"
	setting.SetLabel("Auto-refresh")
	setting.SetSize(24, 1)
	setting.Focus()
	return tutorialModel{toggle: setting, notice: "No change delivered"}
}

func (m tutorialModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if changed, ok := msg.(toggle.ChangedMsg); ok {
		m.notice = fmt.Sprintf("Delivered %s=%t", changed.ID, changed.Checked)
		return m, nil
	}

	var cmd tea.Cmd
	m.toggle, cmd = m.toggle.Update(msg)
	return m, cmd
}
```

This fixture makes command delivery observable: activating the toggle changes
the component immediately and emits `toggle.ChangedMsg`, while the app-owned
notice changes only after that result is delivered back to `Update`.

## 2. Snapshot the view and its semantic styles

Use `Snap` for the artifact humans should read first. It strips ANSI and trims
trailing spaces, so its `.golden` file describes layout and content clearly.
Use `SnapCells` with `WithRoles` alongside it when styling matters; the
`.cells.golden` file labels runs with theme roles instead of brittle color
values.

```go
func TestTutorialInitialView(t *testing.T) {
	m := newTutorialModel()

	snaptest.Snap(t, m.View().Content)
	snaptest.SnapCells(t, m.View().Content,
		snaptest.WithRoles(tuiweave.Dark()))
}
```

Prefer these two artifacts for component rendering. `SnapStyled` also writes a
raw ANSI `.styled.golden`, but exact escape-byte assertions are rarely useful;
reserve it for behavior that cannot be expressed by readable layout and
role-aware cell runs.

## 3. Execute command results explicitly

An MVU command is a function that produces a later message. Deterministic tests
should inspect that boundary: call the command, assert its semantic result, and
then deliver the result to the application model.

```go
next, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
m = next.(tutorialModel)
if cmd == nil || !m.toggle.Checked() {
	t.Fatal("activation did not update the toggle and emit a command")
}

changed, ok := cmd().(toggle.ChangedMsg)
if !ok || changed.ID != "auto-refresh" || !changed.Checked {
	t.Fatalf("unexpected command result: %+v", changed)
}

next, cmd = m.Update(changed)
m = next.(tutorialModel)
if cmd != nil || m.notice != "Delivered auto-refresh=true" {
	t.Fatalf("result was not delivered: notice=%q command=%v", m.notice, cmd)
}
```

The complete
[`TestTutorialCommandDelivery`](tutorial_test.go) also asserts that the notice
does not change before delivery. Keep assertions about IDs, values, selection,
or other final state semantic; do not infer them from rendered text.

## 4. Record a named scenario

`RunScenario` delivers each listed message and records the initial view plus a
checkpoint after every step. It reports whether `Update` returned a command,
but deliberately does not execute that command. Add the produced result as its
own named step when delivery belongs in the behavior under test:

```go
result := snaptest.RunScenario(newTutorialModel(),
	snaptest.ScenarioStep{
		Name: "activate toggle",
		Msg:  tea.KeyPressMsg{Code: tea.KeyEnter},
	},
	snaptest.ScenarioStep{
		Name: "deliver changed result",
		Msg:  toggle.ChangedMsg{ID: "auto-refresh", Checked: true},
	},
)
snaptest.SnapScenario(t, result)

final := result.Model.(tutorialModel)
if !final.toggle.Checked() || final.notice != "Delivered auto-refresh=true" {
	t.Fatal("unexpected final state")
}
```

The scenario golden shows `command: yes` after activation and `command: no`
after explicit delivery. Named steps make failures readable; semantic assertions
on `result.Model` keep correctness checks independent of presentation.

## 5. Update narrowly, then review every artifact

Generate only this tutorial's goldens with a targeted command:

```sh
go test ./snaptest -run TestTutorial -update
```

Then run the package normally and inspect the diff:

```sh
go test ./snaptest
git diff -- snaptest/testdata
```

Read every changed `.golden`, `.cells.golden`, and `.scenario.golden` file
before accepting it. A passing `-update` run proves only that files were
written; it does not prove the rendering or transition is correct. Never use a
broad regeneration to silence a snapshot failure you do not understand.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/snaptest)
- [Testing a component](../AGENTS.md#testing-a-component)
