# gotui — Agent Conventions

Rules for writing code **with** gotui (apps) and **in** gotui (components).
This file is deliberately short; each rule links to a runnable example once
one exists. Architecture rationale lives in [DESIGN.md](DESIGN.md) — read it
before adding a component; you don't need it to build an app.

## The stack

- Runtime: `charm.land/bubbletea/v2` (MVU; root model's `View()` returns `tea.View`)
- Styling: `charm.land/lipgloss/v2` (components render styled **strings**)
- Layout: `github.com/ishansain/gotui/layout` (flexbox-like constraints → rects)
- Testing: `github.com/ishansain/gotui/snaptest` (golden files)

## Hard rules

1. **Colors come from `gotui.Theme` roles — never literals.** No hex strings,
   no `lipgloss.Color("...")` outside theme definitions. If no role fits,
   stop and flag it; do not improvise a color.
2. **Never import `ultraviolet` in app code or component packages.** Geometry
   comes from `gotui/layout` (`layout.Rect`); UV is a library-internal detail.
3. **Every component sizes itself only via `SetSize(w, h)`** and must render
   exactly within that box — no measuring the terminal, no guessing.
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
6. **Every new component ships with:** golden tests, a runnable example under
   `examples/<component>/`, and a recipe entry in this file.

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
wrap it once, at the root: `return tea.NewView(view)`.

## Writing a gotui component

Component contract (full rationale in DESIGN.md §4):

```go
package widget

type Model struct { /* value type; unexported fields */ }

func New(theme gotui.Theme /*, config... */) Model   // derive styles from roles here
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd)  // concrete type, not tea.Model
func (m Model) View() string                         // render within the set box
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

### textinput

```go
ti := textinput.New(theme)
ti.Placeholder = "type…"
ti.Focus()                       // cursor renders; keys accepted
// enter is NOT handled: check it in the app and read ti.Value(), then ti.Reset()
```

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

### Key-routing pattern (multi-component apps)

Order matters; see [examples/demo](examples/demo/main.go) for the full shape:

1. If a modal is open, all keys go to it — nothing else.
2. Global keys next (`ctrl+c`, `tab`/`shift+tab` + `fm.Apply`, app actions).
3. Everything else is delegated to *all* components; blurred ones ignore
   keys themselves.
