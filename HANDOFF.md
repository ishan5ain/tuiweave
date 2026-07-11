# gotui — Session Handoff

Updated: 2026-07-10

## Current state

- Branch: `main`
- HEAD at handoff creation: `b2c4486` (`HANDOFF: update session continuity record after 15-commit Phase 5 delivery`)
- At handoff creation, `main` was one local commit ahead of `origin/main`.
- Go package targets: 46 total — root, 29 reusable top-level packages, 6
  agentic packages, 1 internal package, and 9 examples.
- `HANDOFF.md` is tracked as the session-continuity record.

The north star remains: a small, composable, domain-neutral Go vocabulary for
terminal interfaces that is predictable for humans and coding agents. gotui
owns reusable components and optional domain kits; applications own routing,
backends, persistence, policy, and session orchestration.

## Phase status

### Phase 5 — Editing and interaction depth

Complete except for the explicitly deferred textarea IME/editor follow-up.

Completed slices include:

- Textarea tier 2: logical-rune selection, replacement, bounded undo/redo,
  semantic editing actions, word movement/deletion, bounded kill/yank ring,
  kill coalescing, and yank rotation.
- Grapheme/cell-aware rendering for textarea, textinput, shared/agentic cells,
  dialogs, permission panels, tables, overlays, and `snaptest.SnapCells`.
- `autocomplete`: bounded, stable-ID suggestions beside an app-owned
  `textinput`, with disabled handling, semantic actions, inspection, and a
  runnable example.
- `mouse`: standard vertical wheel deltas plus app-owned coordinate helpers;
  viewport wheel scrolling works while blurred.
- `focus.Stack`: copy-safe nested focus layers with `Push`, `Pop`, `Depth`,
  `Index`, and grouped `Apply`; `focus.Scope` remains the simpler one-modal
  convenience.
- Interaction scenarios: browser blurred-wheel routing; ops root → palette →
  nested restart confirmation with explicit cancellation-result delivery and
  narrow `36×12` coverage; demo explicit dialog cancellation delivery.
- Agent discoverability audit: package/example indexes, task routing, recipes,
  nested-modal and command-result guidance, validation caveats, and the agent
  workflow are synchronized across `README.md`, `AGENTS.md`,
  `AGENT-CATALOG.md`, `DESIGN.md`, and `PLAN.md`.

Remaining Phase 5 item:

- Textarea IME behavior and further editor-specific commands, only when a
  concrete reusable requirement justifies them.

### Phase 4 — Composition layer

The initial composition gate is complete: framing, tabs, menus, toolbars,
split panes, stacked chrome, controls, palettes, line helpers, ops, and browser
coverage are in place. Deferred framing/structural/action refinements remain
demand-driven rather than prerequisites.

### Phase 6 — Domain kits and Pi validation

This is the recommended next workstream:

- Scaffold the separate Pi frontend app repository and its JSON-RPC client.
- Keep Pi protocol, persistence, and session lifecycle outside gotui.
- Use the Pi app as a demanding consumer to identify genuinely reusable gaps.
- Review `agentic/` APIs for generalization without adding backend-specific
  clients to this repository.
- Add a small live-capture/pty validation script only when a real app exists.

Do not start by adding Pi-specific abstractions to gotui. First confirm the
separate app-repo boundary and the smallest protocol/client surface it needs.

## Important architecture and usage decisions

1. Colors come only from `gotui.Theme` semantic roles; never add raw color
   literals in components or applications.
2. Applications and components do not import Ultraviolet. Geometry comes from
   `gotui/layout`; UV is confined to `layout`, `overlay`, and `snaptest`.
3. Bounded components receive dimensions only through `SetSize(w, h)` and
   render exactly within the assigned box. `dialog`, `permission`, and
   `spinner` document their size-mode exceptions.
4. Every MVU update reassigns the returned model and collects its `tea.Cmd`.
5. The app owns message routing. A modal layer receives keys before globals or
   background components; nested `focus.Stack` layers must not leak input to
   lower layers.
6. `focus.Stack.Apply` receives fresh `focus.Group` addresses from root to top
   layer. Dialogs/permission prompts may use an empty group because their own
   `Update` methods own internal option selection.
7. Mouse wheel behavior can be component-owned even while blurred. Coordinate
   hit-testing, clicks, releases, and their meaning remain app-owned through
   `mouse.Position`/`mouse.InBounds`.
8. Components emit typed result/selection commands; deterministic scenarios do
   not execute commands implicitly. Tests deliver command-produced messages as
   explicit later steps.
9. Semantic inspection/actions are optional, data-only surfaces. The app
   assembles the tree, owns stable IDs and privacy decisions, and routes
   qualified actions to local component intents.
10. `internal/grapheme` protects combining marks at the UV cell boundary; do
    not duplicate that workaround in application code.

## Package map

```text
gotui/              Theme roles (Dark/Light)
├── action/          Shared selectable-action definitions
├── autocomplete/    Bounded suggestions for app-owned text inputs
├── button/          Focusable single-action control
├── dialog/          Modal confirmation box
├── focus/           Manager, one-modal Scope, nested Stack
├── frame/           Panel, Divider, Badge, PanelContentRect
├── help/            Key-hint bar
├── inspect/         Semantic UI tree and action metadata
├── layout/          Flexbox constraints → rectangles
├── line/            Width-aware single-row helpers
├── list/            Filterable scrolling list
├── menu/            Vertical action menu
├── mouse/           Wheel deltas, coordinates, app-owned hit-testing
├── overlay/         Cell-space modal/popover composition
├── palette/         Filtered command-palette foundation
├── progress/        Passive task indicator
├── scrollbar/       One-column scroll indicator
├── snaptest/        Plain, styled, cell, and scenario goldens
├── spinner/         Tick-driven activity indicator
├── splitpane/       Horizontal pane composition
├── stack/           Vertical section composition
├── statusbar/       One-line left/right bar
├── table/           Columnar table with selection
├── tabs/            Sibling-view navigation
├── textarea/        Cell-aware multiline editor with history/kill ring
├── textinput/       Cell-aware single-line editor
├── toggle/          Boolean setting control
├── toolbar/         Horizontal action strip
├── viewport/        Scrollable content window
├── internal/grapheme/ Shared UV-boundary combining-mark protection
└── agentic/         chat, markdown, toolcall, diffview, permission, usagebar
```

Canonical examples:

```text
examples/statusbar     Smallest layout → SetSize → render wiring
examples/demo          Multi-pane app with modal cancellation
examples/chat          Mock streaming agentic session
examples/table         Git-status table + diff + scrollbars
examples/frame         Panels, tabs, menus, toolbar, controls, composition
examples/palette       Filtered command discovery and activation
examples/autocomplete  App-owned input plus bounded suggestions
examples/ops           Non-agentic console with nested modal routing
examples/browser       File list, filtering, preview, and blurred wheel input
```

## Recommended fresh-session workflow

1. Read `PLAN.md` and check whether the task is Phase 6 work or a concrete
   deferred Phase 5 requirement.
2. Use `AGENT-CATALOG.md` to map the task to a package and canonical example.
3. Read the corresponding `AGENTS.md` recipe and preserve its ownership,
   sizing, focus, routing, and command-handling conventions.
4. Inspect the closest example before creating a new abstraction. Keep domain
   orchestration in the application or separate app repository.
5. Add semantic inspection/actions and a named `snaptest.RunScenario` step when
   the feature is agent-operated or stateful.
6. For intentional visual changes, run the snapshot update command and inspect
   the golden diff. Packages without snapshot flags (`action`, `focus`,
   `inspect`, `layout`, `mouse`) may report the expected
   `flag provided but not defined: -update` error.
7. Finish with the full build, vet, and test gate.

## Validation

The last complete validation passed:

```sh
go build ./...
go vet ./...
go test ./...
```

The repository-wide `go test ./... -update` command passed all
snapshot-bearing packages. Its nonzero status is limited to the known
non-snapshot packages listed above rejecting the `-update` flag.

## Key files

| File | Purpose |
|---|---|
| `PLAN.md` | Checked roadmap and next workstream |
| `DESIGN.md` | Architecture and ownership rationale |
| `AGENTS.md` | Detailed coding recipes and hard rules |
| `AGENT-CATALOG.md` | Task/package/example routing and workflow |
| `README.md` | Public package overview and project orientation |
| `HANDOFF.md` | This continuity record |
