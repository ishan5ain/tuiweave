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
    snaptest.Snap(t, w.View())        // plain-text golden
    // snaptest.SnapStyled(t, ...)    // add when styling itself is the subject
}
```

Snapshot states, not just defaults: focused/blurred, empty/full, truncation
at small sizes. The plain `.golden` file is the artifact to read when judging
whether output is correct.

## Recipes

*(grows one entry per component, starting in Phase 1)*
