# gotui — Agent Conventions

Rules for writing code **with** gotui (apps) and **in** gotui (components).
This file is deliberately short; recipes link into the runnable example apps
(`examples/statusbar`, `examples/demo`, `examples/chat`, `examples/frame`, `examples/palette`, `examples/ops`, `examples/browser`). Architecture
rationale lives in [DESIGN.md](DESIGN.md) — read it before adding a
component; you don't need it to build an app.

Use [AGENT-CATALOG.md](AGENT-CATALOG.md) to route a task to the right package
and example before reading the detailed recipes below.

## The stack

- Runtime: `charm.land/bubbletea/v2` (MVU; root model's `View()` returns `tea.View`)
- Styling: `charm.land/lipgloss/v2` (components render styled **strings**)
- Layout: `github.com/ishansain/gotui/layout` (flexbox-like constraints → rects)
- Action definitions: `github.com/ishansain/gotui/action` (shared stable-ID action definitions)
- Composition: `github.com/ishansain/gotui/frame` (width-aware themed decoration)
- Navigation: `github.com/ishansain/gotui/tabs` (focusable sibling-view tabs)
- Actions: `github.com/ishansain/gotui/menu` (focusable action choices)
- Toolbars: `github.com/ishansain/gotui/toolbar` (horizontal action strips)
- Split panes: `github.com/ishansain/gotui/splitpane` (width-aware view composition)
- Stacked chrome: `github.com/ishansain/gotui/stack` (headers, sections, footers)
- Progress: `github.com/ishansain/gotui/progress` (passive task indicators)
- Toggles: `github.com/ishansain/gotui/toggle` (focusable boolean settings)
- Buttons: `github.com/ishansain/gotui/button` (focusable single actions)
- Command palettes: `github.com/ishansain/gotui/palette` (filtered action discovery)
- Line composition: `github.com/ishansain/gotui/line` (truncation, alignment, fill zones)
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

Components are bounded by default. Use `layout.SizeModeOf(component)` when a
composition includes an exception: `spinner` is `SizeIntrinsic`, while
`dialog` and `permission` are `SizeWidthBounded` (width-constrained with
natural content height). Do not assume every `SetSize` height is rendered as
rows for those documented modes.

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

### frame

Use `frame` for reusable decoration around app-owned content. It is a pure
string utility, not an MVU component: the app passes widths from its layout
rects and composes the returned strings at the root. Full wiring:
[examples/frame](examples/frame/main.go).

```go
panel := frame.Panel(theme, content, area.Dx(), frame.PanelOptions{
    Title: "Jobs",
    Focused: focused,
    Padding: 1,
})
contentArea := frame.PanelContentRect(area, frame.PanelOptions{Padding: 1})
component.SetSize(contentArea.Dx(), contentArea.Dy())
divider := frame.Divider(theme, area.Dx())
badge := frame.Badge(theme, "healthy", frame.BadgeSuccess)
```

- `Panel` uses `Border` or `BorderFocused`, `SurfaceRaised`, and text roles;
  titles truncate to the available top-border width.
- `Panel` returns a natural-height frame with exactly the requested width. It
  handles multiline and narrow content, but the app still owns height layout.
- `PanelContentRect` returns the exact inner rectangle for a fixed outer panel
  area, accounting for the rounded border and symmetric padding. Use it to
  size bounded children instead of repeating `width-4`/`height-4` arithmetic.
- `Badge` kinds describe meaning (`Accent`, `Muted`, `Success`, `Warning`,
  `Danger`, `Info`); do not select a kind merely to obtain a preferred color.
- Keep borders, titles, and badges in the composition layer. Do not make a
  domain component reimplement them or splice side borders around multiline
  content itself.

### tabs

Use `tabs` for navigation among sibling views; the app owns the selected
view's content. Full composition coverage is in
[examples/frame](examples/frame/main.go).

```go
nav := tabs.New(theme)
nav.SetTabs(
    tabs.Tab{ID: "overview", Label: "Overview"},
    tabs.Tab{ID: "logs", Label: "Logs"},
)
nav.SetSize(area.Dx(), 1)
nav.Focus()
// In the app, render nav.View() and switch content from nav.SelectedID().
```

- `tabs` handles left/right arrows, `h`/`l`, `home`/`g`, and `end`/`G` only
  while focused; leave `tab`/`shift+tab` to the app's global focus manager.
- `SetTabs` preserves selection by stable ID when possible. Keep IDs stable
  across updates so `SelectedID()` and semantic `select.<id>` actions remain
  meaningful.
- It renders one exact-width row, keeps the selected tab visible when narrow,
  and drops earlier tabs from the visible window as selection moves right.
- The tab strip is navigation chrome, not a container for the tab bodies; use
  the app's layout and render the selected view separately.

### action definitions

Use `action.Item` as the application-owned definition when one operation appears
in more than one selectable surface. Keep IDs stable; `Disabled` is shared by
menus, toolbars, and palettes, while each component retains its own renderer,
focus behavior, and `SelectedMsg` type.

Use `action.SelectID(id)` and `action.ParseSelectID(id)` when exposing or
handling semantic `select.<id>` actions; tabs use the same protocol.

```go
commands := []action.Item{
    {ID: "open", Label: "Open workspace"},
    {ID: "refresh", Label: "Refresh data"},
    {ID: "delete", Label: "Delete workspace", Disabled: true},
}
```

### menu

Use `menu` for a vertical set of application-owned actions. Disabled items
remain visible and are skipped by navigation. `enter` emits a typed
`menu.SelectedMsg` command. Full composition coverage is in
[examples/frame](examples/frame/main.go).

```go
actions := menu.New(theme)
actions.SetItems(commands...)
actions.SetSize(area.Dx(), area.Dy())
actions.Focus()

case menu.SelectedMsg:
    // The app owns the operation associated with msg.ID.
```

- `menu` handles up/down, `j`/`k`, page navigation, `home`/`g`, `end`/`G`,
  and enter only while focused; global tab order remains app-owned.
- Keep IDs stable so `SelectedID()`, `SelectedMsg`, and semantic
  `select.<id>` actions remain useful to tests and agents.
- Always reassign the returned model and collect the command from `Update`;
  activation is intentionally a command-producing transition.
- Use `Description` for inspection/action metadata. The menu renders one-line
  labels; application-owned detail belongs beside or below it.

### toolbar

Use `toolbar` for a horizontal strip of application-owned actions. It follows
the menu action contract while keeping the selected action visible when the
available width is narrow. Full composition coverage is in
[examples/frame](examples/frame/main.go).

```go
tools := toolbar.New(theme)
tools.SetItems(commands...)
tools.SetSize(area.Dx(), 1)
tools.Focus()

case toolbar.SelectedMsg:
    // The app owns the operation associated with msg.ID.
```

- `toolbar` handles left/right, `h`/`l`, `home`/`g`, `end`/`G`, and enter only
  while focused; leave `tab`/`shift+tab` to the app's focus manager.
- Keep IDs stable for `SelectedID()` and semantic `select.<id>` actions.
- Disabled actions remain visible, are skipped during navigation, and expose
  disabled semantic actions for inspection.
- The toolbar is one row of action chrome. It does not own a command's
  side-effects or the content that the action changes.

### splitpane

Use `splitpane.Horizontal` when two sibling views need a shared width split,
natural-height alignment, and a themed divider. The callbacks receive their
assigned widths; the app still owns component state and message routing.
Full wiring is in [examples/frame](examples/frame/main.go).

```go
view := splitpane.Horizontal(theme, area.Dx(), splitpane.Options{Gap: 1},
    func(width int) string { return leftView(width) },
    func(width int) string { return rightView(width) },
)
```

- `Ratio` controls the left pane percentage (default 50); `Gap: 0` means a
  one-cell divider, while a negative gap removes the divider.
- The helper aligns both views to the taller natural height and returns an
  exact-width composition. Widths too narrow for both panes render empty.
- Use the callbacks to render width-aware content; do not measure the terminal
  inside a component or hide application routing inside the helper.

### stack

Use `stack.Vertical` for width-aware headers, body sections, separators, and
footers. Empty sections are omitted, and gap rows are width-filled. Full
composition coverage is in [examples/frame](examples/frame/main.go).

```go
view := stack.Vertical(theme, width, stack.Options{Gap: 0},
    func(width int) string { return headerView(width) },
    func(width int) string { return bodyView(width) },
    func(width int) string { return footerView(width) },
)
```

- Set `Divider: true` to insert themed `BorderMuted` rules between sections;
  use `Gap` for blank rows.
- The helper returns natural height and exact requested width. It does not
  allocate component rectangles or route messages; the app still owns those
  decisions.
- Prefer callbacks over manual newline concatenation when sections can be
  conditionally present or have pre-styled, width-aware content.

### progress

Use `progress` for a passive task or operation indicator. The application owns
the work state and updates the value; the component owns themed status styling,
label truncation, and exact-width rendering.

```go
p := progress.New(theme)
p.SetSize(area.Dx(), 1)
p.SetLabel("Indexing")
p.SetPercent(0.72)
p.SetStatus(progress.StatusInfo)
```

- `SetPercent` clamps values to `[0, 1]`; `SetStatus` selects semantic
  `Normal`, `Success`, `Warning`, `Danger`, or `Info` styling.
- The indicator is non-focusable and handles no messages. It is suitable for
  file operations, dashboards, and service panels; use `agentic/usagebar` for
  model/token/cost/context session metrics.
- It renders one exact-width row and gives the label and percentage space back
  to the bar at narrow widths. Do not pre-truncate the label in the app.

### toggle

Use `toggle` for an application-owned boolean setting. It is a focusable
control: space, `x`, and enter change the value and emit a `ChangedMsg`; the
application owns persistence and side effects.

```go
t := toggle.New(theme)
t.ID = "auto-refresh"
t.SetLabel("Auto-refresh")
t.SetChecked(true)
t.SetSize(area.Dx(), 1)
t.Focus()

case toggle.ChangedMsg:
    // Persist or apply msg.Checked for msg.ID in the application.
```

- Register it in the app's `focus.Manager` when it is part of tab order, and
  apply fresh component addresses after every focus change.
- `toggle` exposes `focus`, `blur`, `toggle`, `on`, and `off` semantic actions
  through `Inspect()`. Disabled toggles remain renderable but reject focus and
  state changes.
- It renders one exact-width row and truncates its label at narrow widths; do
  not pre-truncate the setting label in the app.

### button

Use `button` for one application-owned action that needs a focused, visible
activation target. Enter and space emit a typed `PressedMsg`; the application
owns the resulting operation.

```go
b := button.New(theme)
b.ID = "open-workspace"
b.SetLabel("Open workspace")
b.SetSize(area.Dx(), 1)
b.Focus()

case button.PressedMsg:
    // Start the operation associated with msg.ID in the application.
```

- Register the button in the app's `focus.Manager` when it belongs in tab
  order, and apply fresh component addresses after every focus change.
- `button` exposes `focus`, `blur`, and `activate` semantic actions through
  `Inspect()`. Disabled buttons remain visible but reject focus and activation.
- Use `menu` or `toolbar` when the user must choose among several actions; use
  `button` when there is one action at that location.
- It renders one exact-width row and truncates the label inside its brackets;
  do not pre-truncate the label in the app.

### palette

Use `palette` when users need to discover and activate many actions by query.
It combines a text input and filtered action window into one focusable bounded
component; visibility and command side effects remain application-owned.

```go
p := palette.New(theme)
p.SetItems(commands...)
p.SetSize(area.Dx(), area.Dy())
p.Focus()

case palette.SelectedMsg:
    // Run the application operation associated with msg.ID.
```

- Filtering is case-insensitive over stable ID, label, and description. The
  selected index remains an original item index, even while filtered.
- Up/down, `j`/`k`, page keys, home/end, enter, and space follow the established
  selection conventions. Disabled actions remain visible and are skipped.
- Semantic `select.<id>`, `next`, `previous`, `first`, `last`, `clear`, and
  `activate` actions are exposed through `Inspect()` for tests and agents.
- The palette renders an explicit `No matching commands` state and exact-width
  query/result rows. Do not pre-filter or pre-truncate actions in the app.

### line

Use `line` for one-row composition when several values must remain readable and
exactly sized at narrow widths. It preserves ANSI styling while measuring
visible cells, so callers do not need separate byte-length logic.

```go
title := line.Fit("service overview", width, line.AlignLeft)
position := line.Fit("Ln 12", width, line.AlignRight)
footer := line.Join(width, frame.Divider(theme, width), "q quit",
    line.JoinOptions{Gap: 1, NoEllipsis: true})
```

- `Truncate` uses an ellipsis; `Fit` truncates and pads to exactly the requested
  width with left, center, or right alignment.
- `Fill` repeats a visible pattern without splitting a multi-cell glyph. `Join`
  reserves a right-side value and uses the remaining cells as a fill zone;
  narrow joins truncate the left side first.
- Set `NoEllipsis` for decorative rules, dividers, or other patterns where an
  ellipsis would be misleading. The helper does not invent colors or surfaces;
  wrap its result in the app's themed style when needed.

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

For a complete filterable-list plus scrollable-preview composition, see
[examples/browser](examples/browser/main.go).

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
ta.SelectAll()                   // logical-rune selection; inspect with SelectedText()
ta.Undo() / ta.Redo()             // bounded edit history; CanUndo/CanRedo report state
ta.Yank()                         // insert the latest killed text; CanYank reports state
// enter inserts a newline INSIDE the textarea — for chat-style "enter sends",
// intercept enter at the app level and offer alt+enter for newlines:
case "enter":     /* read ta.Value(), send, ta.Reset() */
case "alt+enter": ta.InsertString("\n")
```

- `textinput` and `textarea` calculate display geometry in terminal cells while
  keeping logical cursor positions rune-based. Wide and combining graphemes
  stay intact at prompt, placeholder, cursor, and wrap/window boundaries.

Grow a chat input with its content by re-splitting the layout after edits:
`layout.Len(min(4, ta.ContentHeight()))` — see [examples/chat](examples/chat/main.go).

- `shift+arrow`, `shift+home/end`, and `ctrl+a` select logical runes across
  lines; typing or Enter replaces the selection. `SelectedText()` includes
  logical newlines, while wrapping remains a display concern.
- `ctrl+z`/`ctrl+y` (or `ctrl+shift+z`) undo and redo up to 100 edit states.
  Programmatic `SetValue` and `Reset` establish a fresh history baseline;
  selection changes themselves are not edits.
- `ctrl+left/right` move by whitespace-delimited words; Shift variants select
  those words. `ctrl+w`, `ctrl+u`, and `ctrl+k` kill text into a bounded
  20-entry ring; consecutive kills coalesce in direction-aware order. `alt+y`
  or the semantic `yank` action inserts the latest kill, and repeating it
  rotates through older kills without appending duplicates. `ctrl+y` remains
  redo for compatibility with the history contract.
- `ctrl+delete` kills the next whitespace-delimited word; `alt+backspace` is
  the backward-word-kill alias. Both share the kill ring and undo behavior.
- Textarea display geometry is grapheme- and cell-aware for wide and combining
  characters, while logical selection positions remain rune-based. Richer
  editing commands and IME behavior remain later work; other text-bearing
  components still need a broader cell-width audit.

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

For a modal or conditionally visible group, use `focus.Scope` to save and
restore the parent index while isolating background components:

```go
m.paletteScope = focus.NewScope(1)
m.paletteScope.Enter(m.fm)
m.paletteScope.ApplyBackground(m.fm, &m.tabs, &m.actions)
m.paletteScope.Apply(&m.palette)
// on palette.SelectedMsg:
m.paletteScope.Exit(&m.fm)
m.paletteScope.ApplyBackground(m.fm, &m.tabs, &m.actions)
m.paletteScope.Apply(&m.palette)
```

- A scope stores only indices and state, never component pointers, so it is
  safe inside copied MVU models.
- `ApplyBackground` applies the parent manager while inactive and blurs the
  background while active. `Apply` focuses scoped components while active and
  blurs them while inactive.
- Route all messages to the active scope's components while the scope is open;
  visibility and modal results remain application-owned.

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

### reference compositions

The reference apps are deterministic composition recipes, not application
frameworks. Use them to choose a shape and verify the contracts below before
adding a new helper.

#### operations console

[examples/ops](examples/ops/main.go) is the action-heavy recipe: tabs across
the top, a menu beside a fixed-control operations panel, a command-palette
scope, and a statusbar. Its focus order is tabs → menu → table → toggle →
button; the palette temporarily owns focus through `focus.Scope`.

- Define one `[]action.Item` slice for actions shared by the menu and palette;
  keep IDs and disabled state in the application-owned data.
- Use `frame.PanelContentRect` before sizing bounded children. For a panel with
  a flexible table and fixed controls, split the inner rect with
  `layout.Vertical(layout.Fill(1), layout.Len(1), ...)`.
- Route modal/palette keys first, global keys (`tab`, `shift+tab`, quit, and
  app commands) second, then delegate to every background component and batch
  every returned command.
- Build an app-owned `inspect.Group` containing the tabs, actions, operation
  table, controls, and conditional palette. Exercise focus, palette activation,
  and narrow layout with `snaptest.RunScenario` and screen goldens.

#### file browser

[examples/browser](examples/browser/main.go) is the selection-and-preview
recipe: tabs → filterable file list and scrollable preview → statusbar. It
uses a deterministic mock workspace so tests cover UI behavior without file
system or persistence concerns.

- Keep the focus order explicit: tabs → filter input → list → viewport. `/`
  is an app-level shortcut that moves focus to the filter; `tab` remains the
  global focus key.
- When the query changes, call `list.SetFilter(query)` and continue using
  `list.Selected()` as the original-item index. Sync the preview only when the
  selected entry changes, and reset its viewport to the top for a new entry.
- Forward mouse-wheel messages to the viewport regardless of focus. Put a
  scrollbar in a sibling one-column segment. If that segment sits beside a
  padded panel, reserve the outer bar column before calling
  `frame.PanelContentRect`; size the viewport from the panel's content rect and
  align the bar to the panel's content rows.
- Verify the normal and narrow screens, an inspection tree, filter/selection,
  focus traversal, and preview scrolling with readable goldens and named
  scenario checkpoints.

#### cross-component completion checklist

Before calling a composition complete, confirm that it has:

- one `tea.WindowSizeMsg` layout path using `layout.Rect` and `SetSize`;
- a documented focus order with fresh `focus.Manager.Apply` addresses;
- global-key routing separated from component delegation, with no dropped
  `tea.Cmd` values;
- stable IDs and an app-owned inspection tree when the UI is agent-operated;
- at least one narrow rendering golden and one interaction scenario;
- explicit empty, loading, or unavailable states where the composed view can
  lack content.

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
- Built-in user, assistant, note, and tool-call cells keep headers, padding,
  and output within the assigned terminal-cell width, including narrow widths.
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
