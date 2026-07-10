# gotui — Agent Conventions

Rules for writing code **with** gotui (apps) and **in** gotui (components).
This file is deliberately short; recipes link into the runnable example apps
(`examples/statusbar`, `examples/demo`, `examples/chat`). Architecture
rationale lives in [DESIGN.md](DESIGN.md) — read it before adding a
component; you don't need it to build an app.

## The stack

- Runtime: `charm.land/bubbletea/v2` (MVU; root model's `View()` returns `tea.View`)
- Styling: `charm.land/lipgloss/v2` (components render styled **strings**)
- Layout: `github.com/ishansain/gotui/layout` (flexbox-like constraints → rects)
- Testing: `github.com/ishansain/gotui/snaptest` (golden files)
- Domain packages: `gotui/agentic/…` (markdown, chat, toolcall, diffview,
  permission, usagebar) — optional, backend-agnostic layers built on the
  primitives

The core library is domain-neutral. Its primitives should be suitable for
editors, dashboards, file browsers, forms, operational tools, and agentic UIs.
Keep product-specific orchestration, backend clients, persistence, and session
lifecycle in application repositories.

gotui is agent-friendly by design. Treat its primitives, theme roles, layout
rules, interaction conventions, examples, and snapshots as a small design
grammar: compose from that vocabulary first, then add a new abstraction only
when the existing vocabulary cannot express the intended behavior cleanly.

## Hard rules

1. **Colors come from `gotui.Theme` roles — never literals.** No hex strings,
   no `lipgloss.Color("...")` outside theme definitions. If no role fits,
   stop and flag it; do not improvise a color.
2. **Never import `ultraviolet` in app code or component packages.** Geometry
   comes from `gotui/layout` (`layout.Rect`); UV is a library-internal detail.
   (Inside the library, exactly three packages touch it: `layout`, `overlay`,
   `snaptest`.)
3. **Every bounded component sizes itself only via `SetSize(w, h)`** and must
   render exactly within that box — no measuring the terminal, no guessing.
   Intrinsic components may render at natural size, but must document that
   contract and still expose `SetSize` when they satisfy `layout.Sizable`.
4. **MVU discipline:** always reassign the model returned by `Update` and
   always collect the returned `tea.Cmd`:
   ```go
   m.list, cmd = m.list.Update(msg)
   cmds = append(cmds, cmd)
   ```
   Dropping either is a bug even when it appears to work.
5. **Every change to rendering is verified through snaptest goldens.** After
   an intentional visual change: `go test ./... -update`, then read the golden
   diff in git and confirm it matches your intent before considering the task
   done. Never regenerate goldens to silence a failure you don't understand.
6. **Every new component ships with:** golden tests, coverage in a runnable
   example app under `examples/`, and a recipe entry in this file.
7. **Classify additions before implementing them:** domain-neutral primitives
   belong in a top-level package, reusable composition helpers belong in a
   top-level utility package, and domain-specific behavior belongs under a
   domain package or in the application.
8. **Prefer established interaction patterns:** focus, selection, scrolling,
   modal routing, keyboard handling, empty states, and narrow-width behavior
   should follow existing component conventions unless the component documents
   a deliberate difference.

## Wiring an app (the only layout pattern)

Split the window on every `tea.WindowSizeMsg`; let rects size the components:

```go
case tea.WindowSizeMsg:
    m.width, m.height = msg.Width, msg.Height
    layout.Vertical(
        layout.Len(3),  // header
        layout.Fill(1), // body
        layout.Len(1),  // statusbar
    ).Apply(layout.NewRect(0, 0, msg.Width, msg.Height),
        &m.header, &m.body, &m.status)
```

Compose the final frame with `lipgloss.JoinVertical` / `JoinHorizontal` and
wrap it once, at the root: `return tea.NewView(view)`. Full-screen apps set
`v.AltScreen = true` on the returned view — bubbletea v2 has **no**
`tea.WithAltScreen()` program option (a v1 idiom agents often reach for);
mouse support is also a view field (`v.MouseMode = tea.MouseModeCellMotion`).

## Writing a gotui component

Component contract (full rationale in DESIGN.md §4):

```go
package widget

type Model struct { /* value type; unexported fields */ }

func New(theme gotui.Theme /*, config... */) Model   // derive styles from roles here
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd)  // concrete type, not tea.Model
func (m Model) View() string                         // bounded or documented intrinsic size
func (m *Model) SetSize(width, height int)           // layout.Sizable
func (m *Model) Focus() / Blur()                     // interactive components only
```

One package per component (`gotui/widget`, flat, like bubbles). Agentic
domain components go under `gotui/agentic/<name>` and consume plain Go types —
no agent-backend clients (JSON-RPC etc.) in this repo.

## Testing a component

```go
func TestWidgetDefault(t *testing.T) {
    w := widget.New(gotui.Dark())
    w.SetSize(40, 5)
    snaptest.Snap(t, w.View())        // plain-text golden: layout & content
    snaptest.SnapCells(t, w.View(),   // style-run golden: which role styles what
        snaptest.WithRoles(gotui.Dark()))
    // snaptest.SnapStyled(t, ...)    // raw-bytes golden: rarely needed
}
```

`SnapCells` goldens read like `" gotui " [fg=TextInverted bg=Accent bold]` —
use them to assert roles, e.g. that a selected row uses `SelectionBg`. (Roles
sharing one color label as the first matching Theme field, so `TextInverted`
may appear as `Surface` in the default themes.)

Snapshot states, not just defaults: focused/blurred, empty/full, truncation
at small sizes. The plain `.golden` file is the artifact to read when judging
whether output is correct.

For interaction behavior, use named scenario checkpoints alongside focused
view goldens:

```go
result := snaptest.RunScenario(m,
    snaptest.ScenarioStep{Name: "focus input", Msg: tea.KeyPressMsg{Code: tea.KeyTab}},
    snaptest.ScenarioStep{Name: "submit", Msg: tea.KeyPressMsg{Code: tea.KeyEnter}},
)
snaptest.SnapScenario(t, result)
```

`RunScenario` delivers explicit messages and records whether each update emits
a command; it does not execute commands. Execute timers or I/O explicitly in
the test when their behavior is part of the scenario.

For semantic inspection, components report local state and the app assembles
the tree with stable IDs:

```go
root := inspect.Group("app", "application", inspect.Bounds{Width: w, Height: h},
    inspect.Bind("list", m.list),
    inspect.Bind("input", m.input),
)
data, _ := inspect.Marshal(root)
```

Inspection is data-only. The app owns visibility, layout positions, routing,
privacy decisions, and any transport to an agent or debugging tool.

Action IDs in a bound node are qualified (for example `list.next`). Components
handle the local suffix through `inspect.Invoke("next")`; an application that
receives a tree-level action owns the lookup and routing to the target model.
Prefer semantic actions in tests and automation when available instead of
simulating equivalent key presses.

Two gotchas in hand-written assertions:

- Strip ANSI before `strings.Contains` — renderers style words as separate
  escape-code spans, so matching against raw output fails randomly. Use
  `github.com/charmbracelet/x/ansi`'s `ansi.Strip`.
- A scrolling component's `View()` is only the visible window. To assert on
  full content (e.g. a chat transcript that auto-follows the bottom), render
  the source of truth — `transcript.Cells()` — not the window.

## Recipes

### statusbar

One-line bar with themed segments left and right; lay out with `layout.Len(1)`.
Full wiring: [examples/statusbar](examples/statusbar/main.go).

```go
sb := statusbar.New(theme)
sb.SetSize(width, 1)
sb.SetLeft(
    statusbar.Segment{Text: "gotui", Kind: statusbar.KindAccent}, // "mode" badge; max one per side
    statusbar.Segment{Text: "main.go", Kind: statusbar.KindNormal},
)
sb.SetRight(statusbar.Segment{Text: "12:4", Kind: statusbar.KindMuted})
```

- Kinds map to roles: `KindAccent` = accent-bg badge, `KindMuted` = secondary
  info, `KindSuccess/Warning/Danger/Info` = intent-colored text. Pick by
  meaning, not by color.
- The bar is passive (its `Update` handles nothing) and drops right segments,
  then truncates left, when narrow — don't pre-truncate text yourself.
- Styles are derived from the theme in `New`; to switch themes, construct a
  new statusbar (see `rebuildStatusbar` in the example).

### viewport / list / table (scrolling components)

All three follow the same shape — `SetSize`, `Focus`/`Blur`, vim-ish keys
(`j/k`, `g/G`, `pgup/pgdown`) handled only while focused:

```go
vp := viewport.New(theme); vp.SetContent(text)        // pre-styled ok
l  := list.New(theme);     l.SetItems("a", "b")       // l.SelectedItem()
tb := table.New(theme)
tb.SetColumns(table.Column{Title: "ID", Width: 4}, table.Column{Title: "Name"}) // Width 0 = flex
tb.SetRows([]string{"1", "api"})
```

- They ignore keys when blurred by design — never gate delegation yourself.
- Viewport handles `tea.MouseWheelMsg` even when blurred; forward wheel
  events to it unconditionally.
- Selection styling uses `SelectionBg/Fg` only while focused.

### textinput / textarea

```go
ti := textinput.New(theme)       // single line
ti.Placeholder = "type…"
ti.Focus()                       // cursor renders; keys accepted
// enter is NOT handled: check it in the app and read ti.Value(), then ti.Reset()

ta := textarea.New(theme)        // multi-line, soft-wrapped
// enter inserts a newline INSIDE the textarea — for chat-style "enter sends",
// intercept enter at the app level and offer alt+enter for newlines:
case "enter":     /* read ta.Value(), send, ta.Reset() */
case "alt+enter": ta.InsertString("\n")
```

Grow a chat input with its content by re-splitting the layout after edits:
`layout.Len(min(4, ta.ContentHeight()))` — see [examples/chat](examples/chat/main.go).

### scrollbar

Components never render their own bars. Place one as a 1-column segment and
feed it the component's scroll stats:

```go
layout.Horizontal(layout.Fill(1), layout.Len(1)).Split(area).Assign(&pane, &bar)
m.list.SetSize(pane.Dx(), pane.Dy())
// in render:
lipgloss.JoinHorizontal(lipgloss.Top, m.list.View(), scrollbar.For(theme, m.list))
```

`For` works with every scrolling component (viewport, list, table, chat,
diffview). Note: a table's bar spans its row area, two lines below its top —
prepend two blank lines to align (see [examples/table](examples/table/main.go)).

### list filtering

```go
l.SetFilter(query) // case-insensitive substring; "" clears
l.Selected()       // STILL the original-items index — never remap yourself
l.FilteredLen()    // how many items are displayed
```

The filter input UI is app-owned (a textinput above the list). Navigation,
windowing, and the scrollbar all operate in filtered space automatically.

### spinner

Intrinsic-size; start with `Tick` and forward `TickMsg`:

```go
func (m app) Init() tea.Cmd { return m.spin.Tick() }
case spinner.TickMsg:
    m.spin, cmd = m.spin.Update(msg)
```

### help

`h := help.New(theme); h.SetBindings(help.Binding{Key: "tab", Desc: "focus"}, …)`
— renders whole hints only; drops from the right when narrow.

### focus (tab order)

`focus.Manager` is a **value type holding only an index** — never store
component pointers for focus (they go stale when the MVU model is copied):

```go
m.fm = focus.NewManager(3)              // in the model: fm focus.Manager
case "tab":
    m.fm.Next()
    m.fm.Apply(&m.list, &m.view, &m.input) // fresh addresses, every time
```

### dialog + overlay (modals)

The app owns visibility; the dialog answers via a `ResultMsg` command:

```go
case "ctrl+d":
    m.showDialog = true                  // route ALL keys to the dialog while open
case dialog.ResultMsg:
    m.showDialog = false
    if msg.OK { /* confirmed */ }
// in render:
if m.showDialog { base = overlay.Center(base, m.dlg.View()) }
```

`overlay.Place/Center` composite in cell space — overlays cleanly replace
what's beneath, styles included. Never splice overlay strings manually.

### chat transcript (agentic apps)

The transcript is a stack of **cells** — pointers you keep and mutate as the
session progresses. After mutating a cell in place, call `Invalidate()`
(Append/SetSize do it for you). Full wiring: [examples/chat](examples/chat/main.go).

```go
c := chat.New(theme)
c.Append(chat.NewUser(theme, question))

a := chat.NewAssistant(theme, mdRenderer)  // keep the pointer
a.SetID("assistant-1")                      // stable identity is app-owned
c.Append(a)
// per streamed delta:
a.Append(delta); c.Invalidate()
```

- Auto-follow is built in: stuck to bottom until the user scrolls up,
  re-sticks at bottom. Never call GotoBottom per delta.
- Start a **new** Assistant cell after interleaved content (tool call, diff),
  or the continuation renders above it.
- Adapt anything to a cell with `chat.CellFunc(func(w int) string {...})`
  — e.g. `diffview.Sprint(theme, diff, w)`.
- Identified cells can be found or replaced with `c.Cell(id)` and
  `c.Replace(id, cell)`. Assistant cells expose lifecycle state; tool-call
  blocks support `SetID`, `Retry`, and `Cancel`. Keep IDs stable across retries
  when the logical operation is the same.

### markdown

`markdown.NewRenderer(theme)` is glamour-backed behind the `Renderer`
interface — depend on the interface, never on glamour. `markdown.Sprint`
falls back to raw source on error; transcript UIs should degrade, not fail.

### toolcall

`toolcall.New(theme, name, summary)` returns a `*Block` chat cell: mutate
`SetID`/`SetStatus`/`AppendOutput`/`Retry`/`Cancel`/`Expanded` as the tool progresses. Collapsed
blocks show a line-count hint; expanded output is capped by `MaxOutputLines`.

### permission

Same modal pattern as dialog (app owns visibility, answer arrives as a
`ResultMsg` command), with vertical numbered options; number keys answer
directly, esc picks the **last** option — order options safest-last.

Attach structured provenance when the prompt represents an external operation:

```go
m.perm.SetProvenance(permission.Provenance{
    Tool: "Bash", Operation: "execute", Target: "repo",
    Scope: "workspace", Detail: "go test ./...",
    Impact: "runs tests", Reversibility: "reversible",
    Policy: "shell commands require approval",
})
```

The prompt renders these fields and exposes them through `Inspect()`. Keep the
exact operation and scope visible; the application remains responsible for
policy enforcement and transport.

### usagebar

`u.SetStats(usagebar.Stats{Model, TokensIn, TokensOut, Cost, ContextUsed})` —
context ≥80% styles Warning, ≥95% Danger, automatically.

### Key-routing pattern (multi-component apps)

Order matters; see [examples/demo](examples/demo/main.go) for the full shape:

1. If a modal is open, all keys go to it — nothing else.
2. Global keys next (`ctrl+c`, `tab`/`shift+tab` + `fm.Apply`, app actions).
3. Everything else is delegated to *all* components; blurred ones ignore
   keys themselves.
