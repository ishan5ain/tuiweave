# gotui — Agent Catalog

Use this file to route a task quickly. Read [AGENTS.md](AGENTS.md) for hard
rules and recipes, [DESIGN.md](DESIGN.md) for architectural rationale, and
[PLAN.md](PLAN.md) for unfinished work.

## Fast routing

| Task | Start here |
|---|---|
| Build an app shell | `layout`, then `focus`, `statusbar`, `help` |
| Build a file browser | `examples/browser`; compose `textinput`, `list`, `viewport`, `tabs`, and `focus` |
| Compose a multi-pane reference app | `examples/ops` or `examples/browser`; follow the recipes in `AGENTS.md` |
| Frame content or add semantic badges | `frame.Panel`, `frame.Divider`, `frame.Badge` |
| Navigate sibling views | `tabs`; switch app-owned content from `SelectedID()` |
| Define actions shared across surfaces | `action.Item`; keep IDs and disabled state stable |
| Choose or activate an action | `menu`; handle `menu.SelectedMsg` in the app |
| Add a horizontal action strip | `toolbar`; handle `toolbar.SelectedMsg` in the app |
| Compose two sibling panes | `splitpane.Horizontal`; callbacks receive pane widths |
| Stack headers, sections, and footers | `stack.Vertical`; empty sections are omitted |
| Show task completion | `progress`; passive, exact-width, status-aware indicator |
| Add an on/off setting | `toggle`; focusable, semantic, and emits `ChangedMsg` |
| Add one focused action | `button`; enter/space and semantic activation emit `PressedMsg` |
| Search and activate actions | `palette`; filters stable actions and emits `SelectedMsg` |
| Compose a fixed-width row | `line.Fit`, `line.Fill`, `line.Join`; preserves visible cell widths |
| Split panes or rows | `layout.Vertical` / `layout.Horizontal` + `SetSize` |
| Show selectable records | `list` or `table`; add `scrollbar.For` |
| Show long content | `viewport`; forward mouse wheels; add `scrollbar.For` |
| Accept one line | `textinput`; the app owns Enter/submit |
| Accept chat-style text | `textarea`; app owns Enter/send and optional Alt+Enter newline; word movement and kill/yank are built in |
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
| `snaptest` | Verifying rendering or interactions | `Snap`, grapheme-preserving `SnapCells`, `RunScenario`, `SnapScenario` |
| `inspect` | Describing UI semantics for tests/tools/agents | `Inspect()`, `Bind`, `BindAt`, `Group`, `Marshal` |
| `action` | Sharing selectable action definitions and semantic IDs | `Item{ID, Label, Description, Disabled}`; `SelectID`, `ParseSelectID` |
| `focus` | Managing tab order and modal scopes across copied MVU models | `NewManager(n)` or `NewScope(n)`; apply fresh addresses after changes |
| `overlay` | Compositing a modal or popover | `overlay.Center(base, over)` or `Place`; preserves wide/combining graphemes |
| `scrollbar` | Adding a standalone scroll indicator | `scrollbar.For(theme, component)` in a 1-cell layout segment |
| `frame` | Composing reusable decoration around app-owned strings | `Panel(theme, content, width, PanelOptions{...})`; natural height, exact width |

### General-purpose components

| Package | Use when | Important behavior |
|---|---|---|
| `statusbar` | Rendering one-line left/right segments | Drops right segments, then truncates left when narrow |
| `list` | Selecting one-line items | Filtering preserves original-index `Selected()` |
| `table` | Selecting rows with columns | Header and rule consume two rows; fixed/flex columns remain within the assigned width |
| `viewport` | Scrolling pre-rendered content | Mouse wheel works even when blurred |
| `textinput` | Editing one line | Cell-aware prompt, placeholder, cursor, and horizontal window; Enter is not handled |
| `textarea` | Editing wrapped/multiline text | Logical-rune selection, cell-aware wrapping, bounded undo/redo, word movement/deletion, and a bounded kill/yank ring; Enter inserts a newline |
| `help` | Showing key hints | Drops whole hints from the right when narrow |
| `spinner` | Showing activity | Intrinsic-size; start with `Tick`, forward `TickMsg` |
| `dialog` | Confirming or cancelling | App owns visibility; width-bounded titles, bodies, and buttons; result arrives as `ResultMsg` |
| `tabs` | Navigating sibling views | `SetTabs`, `Focus`, `SelectedID`; left/right while focused |
| `menu` | Choosing application-owned actions | `SetItems([]action.Item...)`, `Focus`, `SelectedMsg`; disabled entries are skipped |
| `toolbar` | Rendering horizontal actions | `SetItems([]action.Item...)`, `Focus`, `SelectedMsg`; selected action stays visible when narrow |
| `splitpane` | Composing two width-aware views | `Horizontal(theme, width, Options, left, right)`; natural-height alignment |
| `stack` | Composing vertical app chrome | `Vertical(theme, width, Options, views...)`; exact-width natural-height sections |
| `progress` | Showing task or operation completion | `SetLabel`, `SetPercent`, `SetStatus`; passive and exact-width |
| `toggle` | Editing a boolean setting | `SetLabel`, `SetChecked`, `Focus`; space/enter/x emit `ChangedMsg` |
| `button` | Activating one application-owned action | `SetLabel`, `Focus`, `PressedMsg`; enter/space activate |
| `palette` | Discovering actions by query | `SetItems([]action.Item...)`, `SetQuery`, `SelectedMsg`; filters ID/label/description |
| `line` | Composing width-aware single rows | `Truncate`, `Fit`, `Fill`, `Join`; `JoinOptions{NoEllipsis:true}` for rules |

### Agentic domain components

| Package | Use when | Important behavior |
|---|---|---|
| `agentic/markdown` | Rendering assistant markdown | Depend on `Renderer`; `Sprint` degrades to raw source |
| `agentic/chat` | Rendering a streaming transcript | Append cells, keep pointers, call `Invalidate()` after mutation; cells clip to the assigned cell width |
| `agentic/toolcall` | Rendering tool status/output | `SetID`, `SetStatus`, `Retry`, `Cancel`; headers and output stay within the assigned width |
| `agentic/diffview` | Rendering unified diffs | `Sprint` for cells; `Model` for scrollable panes |
| `agentic/permission` | Requesting approval | Add `Provenance`; width-bounded options/provenance; answers with `ResultMsg` |
| `agentic/usagebar` | Showing model and usage | `SetStats`; context ≥80% warning, ≥95% danger |

Canonical examples are under [`examples/`](examples/): `statusbar` is the
smallest component wiring example, `frame` demonstrates pure composition,
`demo` composes general primitives, `table` combines selection/diff/scrolling,
`palette` demonstrates filtered command discovery and activation, `ops` is the
non-agentic composition pressure test, `browser` is a filterable file-list and
scrollable-preview reference, and `chat` exercises streaming,
permission, toolcall, diff, markdown, usage, and textarea behavior.

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
- Dropping the `tea.Cmd` returned by `menu.Update` on enter; activation emits
  `menu.SelectedMsg` asynchronously.
- Treating a toolbar as the command implementation; `toolbar.SelectedMsg` is
  only an app-owned activation signal.
- Reimplementing pane width arithmetic and height padding at every call site;
  use `splitpane.Horizontal` for the shared composition pattern.
- Concatenating optional headers and footers with manual newlines; use
  `stack.Vertical` so empty sections and width-filled gaps stay deterministic.
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
