# gotui — Agent Catalog

Use this file to route a task quickly. Read [AGENTS.md](AGENTS.md) for hard
rules and recipes, [DESIGN.md](DESIGN.md) for architectural rationale, and
[PLAN.md](PLAN.md) for unfinished work.

## Fast routing

| Task | Start here |
|---|---|
| Build an app shell | `layout`, then `focus`, `statusbar`, `help` |
| Frame content or add semantic badges | `frame.Panel`, `frame.Divider`, `frame.Badge` |
| Navigate sibling views | `tabs`; switch app-owned content from `SelectedID()` |
| Split panes or rows | `layout.Vertical` / `layout.Horizontal` + `SetSize` |
| Show selectable records | `list` or `table`; add `scrollbar.For` |
| Show long content | `viewport`; forward mouse wheels; add `scrollbar.For` |
| Accept one line | `textinput`; the app owns Enter/submit |
| Accept chat-style text | `textarea`; app owns Enter/send and optional Alt+Enter newline |
| Confirm or gate an action | `dialog` or `agentic/permission` + `overlay` |
| Stream a transcript | `agentic/chat` + `agentic/markdown`; keep cell pointers |
| Show tool execution | `agentic/toolcall`; set ID/status and mutate output |
| Show a diff | `agentic/diffview.Sprint` inline or `diffview.Model` in a pane |
| Expose machine-readable UI state | `inspect` + app-owned `Group`/`Bind` tree |
| Test an interaction sequence | `snaptest.RunScenario` + `SnapScenario` |

## Package map

### Foundations and utilities

| Package | Use when | First API / contract |
|---|---|---|
| `gotui` | Choosing semantic colors | `gotui.Dark()` or `gotui.Light()`; never raw colors |
| `layout` | Converting window space to component boxes | `layout.Vertical(...).Apply(area, &components...)`; `SizeModeOf` for sizing exceptions |
| `snaptest` | Verifying rendering or interactions | `Snap`, `SnapCells`, `RunScenario`, `SnapScenario` |
| `inspect` | Describing UI semantics for tests/tools/agents | `Inspect()`, `Bind`, `BindAt`, `Group`, `Marshal` |
| `focus` | Managing tab order across copied MVU models | `focus.NewManager(n)` + `Apply` after every focus change |
| `overlay` | Compositing a modal or popover | `overlay.Center(base, over)` or `Place` |
| `scrollbar` | Adding a standalone scroll indicator | `scrollbar.For(theme, component)` in a 1-cell layout segment |
| `frame` | Composing reusable decoration around app-owned strings | `Panel(theme, content, width, PanelOptions{...})`; natural height, exact width |

### General-purpose components

| Package | Use when | Important behavior |
|---|---|---|
| `statusbar` | Rendering one-line left/right segments | Drops right segments, then truncates left when narrow |
| `list` | Selecting one-line items | Filtering preserves original-index `Selected()` |
| `table` | Selecting rows with columns | Header and rule consume two rows |
| `viewport` | Scrolling pre-rendered content | Mouse wheel works even when blurred |
| `textinput` | Editing one line | Enter is not handled; read `Value()` in the app |
| `textarea` | Editing wrapped/multiline text | Enter inserts a newline; grow with `ContentHeight()` |
| `help` | Showing key hints | Drops whole hints from the right when narrow |
| `spinner` | Showing activity | Intrinsic-size; start with `Tick`, forward `TickMsg` |
| `dialog` | Confirming or cancelling | App owns visibility; result arrives as `ResultMsg` |
| `tabs` | Navigating sibling views | `SetTabs`, `Focus`, `SelectedID`; left/right while focused |

### Agentic domain components

| Package | Use when | Important behavior |
|---|---|---|
| `agentic/markdown` | Rendering assistant markdown | Depend on `Renderer`; `Sprint` degrades to raw source |
| `agentic/chat` | Rendering a streaming transcript | Append cells, keep pointers, call `Invalidate()` after mutation |
| `agentic/toolcall` | Rendering tool status/output | `SetID`, `SetStatus`, `Retry`, `Cancel`; output is capped when expanded |
| `agentic/diffview` | Rendering unified diffs | `Sprint` for cells; `Model` for scrollable panes |
| `agentic/permission` | Requesting approval | Add `Provenance`; options answer with `ResultMsg` |
| `agentic/usagebar` | Showing model and usage | `SetStats`; context ≥80% warning, ≥95% danger |

Canonical examples are under [`examples/`](examples/): `statusbar` is the
smallest component wiring example, `frame` demonstrates pure composition,
`demo` composes general primitives, `table` combines selection/diff/scrolling,
and `chat` exercises streaming, permission, toolcall, diff, markdown, usage,
and textarea behavior.

## Canonical wiring

On every `tea.WindowSizeMsg`, split the window and size components from the
resulting rectangles:

```go
layout.Vertical(
    layout.Len(3),  // header
    layout.Fill(1), // body
    layout.Len(1),  // status
).Apply(layout.NewRect(0, 0, msg.Width, msg.Height),
    &m.header, &m.body, &m.status)
```

Route messages in this order:

1. Modal open: delegate all keys to the modal.
2. Global keys: quit, tab/shift-tab, app actions.
3. Everything else: delegate to every component; blurred components ignore keys.

Always reassign the returned model and collect its command:

```go
m.list, cmd = m.list.Update(msg)
cmds = append(cmds, cmd)
```

## Agent-operable surfaces

Build the semantic tree in the application because the application owns IDs,
visibility, positions, routing, and privacy:

```go
root := inspect.Group("app", "application", inspect.Bounds{Width: w, Height: h},
    inspect.Bind("list", m.list),
    inspect.Bind("input", m.input),
)
data, err := inspect.Marshal(root)
```

Bound nodes qualify local action IDs (`list.next`). Route the tree-level ID in
the app, then deliver the local intent:

```go
next, cmd := m.list.Update(inspect.Invoke("next"))
m.list = next
```

For interaction goldens, commands are recorded but never executed implicitly:

```go
result := snaptest.RunScenario(m,
    snaptest.ScenarioStep{Name: "focus input", Msg: tabMsg},
    snaptest.ScenarioStep{Name: "submit", Msg: enterMsg},
)
snaptest.SnapScenario(t, result)
```

Execute timers, I/O, or command-produced messages explicitly in the test when
that behavior is part of the scenario.

## Common failure modes

- Using a raw hex color instead of a `gotui.Theme` role.
- Importing Ultraviolet outside its three library-internal packages.
- Measuring the terminal or rendering outside the component's assigned box.
- Dropping the model returned by `Update` or its `tea.Cmd`.
- Forgetting to apply focus after the value-type model has been copied.
- Handling Enter inside `textinput`/`textarea` when the app owns submit policy.
- Gating mouse-wheel delegation on focus for a `viewport`-backed component.
- Using `tab` inside `tabs` and stealing the app's global focus key; use
  arrows or `h`/`l` for tab navigation.
- Splicing one pair of side borders around a multiline string; use
  `frame.Panel`, which frames every rendered line.
- Mutating a chat cell without calling `transcript.Invalidate()`.
- Reusing a logical chat/tool ID for a different operation; keep IDs stable for
  retries, not for unrelated events.
- Regenerating goldens without reading the diff and confirming the visual change.

## Validation loop

```sh
go build ./... && go vet ./... && go test ./...
```

After an intentional rendering change:

```sh
go test ./... -update
git diff -- '**/testdata/**'
```

Run the closest example while developing:

```sh
go run ./examples/demo
go run ./examples/chat
go run ./examples/frame
```

## Where new code belongs

- Reusable terminal behavior with no domain assumptions → a top-level package.
- Reusable composition of primitives → a top-level utility package.
- Agent/session/tool behavior → `agentic/...`, using plain Go types.
- Routing, persistence, backend clients, policy, and session orchestration →
  the application.

Prefer an existing package and recipe first. Add a new abstraction only when
the current vocabulary cannot express the behavior cleanly.
