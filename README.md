# gotui

A modular, customizable TUI component library for Go — built on
[bubbletea v2](https://github.com/charmbracelet/bubbletea),
[lipgloss v2](https://github.com/charmbracelet/lipgloss), and
[ultraviolet](https://github.com/charmbracelet/ultraviolet) — for building
beautiful, consistent terminal frontends. Designed **agent-authored first**:
the API, conventions, and verification loop are optimized so coding agents can
reliably write TUIs with it.

**Status: pre-v1.** Phases 0–3 are complete — foundation, proven conventions,
the core primitive set, and the agentic domain layer. APIs still change freely
until v1 (Phase 5).

## What's here

| Package | Purpose |
|---|---|
| `gotui` (root) | Semantic theme roles (`Theme`), `Dark()`/`Light()` defaults |
| `gotui/layout` | Flexbox-like layout: constraints → rectangles → component sizes |
| `gotui/snaptest` | Snapshot test harness: plain-text, raw-ANSI, and role-labeled cell-grid goldens |
| `gotui/statusbar` | One-line status bar with themed left/right segments |
| `gotui/list` | Scrolling list with selection cursor |
| `gotui/viewport` | Scrollable window over pre-rendered content (keys + mouse wheel) |
| `gotui/textinput` | Single-line input: cursor, placeholder, horizontal scroll |
| `gotui/table` | Fixed + flex columns, header, row selection |
| `gotui/help` | One-line key-hint bar |
| `gotui/spinner` | Tick-driven activity indicator |
| `gotui/dialog` | Modal confirm box answering via `ResultMsg` |
| `gotui/overlay` | Cell-space compositing for modals/popovers (UV inside) |
| `gotui/focus` | Copy-safe tab-order manager |
| `agentic/markdown` | Theme-mapped markdown rendering behind a swappable `Renderer` interface (glamour v2 today) |
| `agentic/chat` | Streaming transcript: user/assistant/tool cells, auto-follow |
| `agentic/toolcall` | Status-aware collapsible tool-call block |
| `agentic/diffview` | Styled unified diffs, inline or scrollable |
| `agentic/permission` | Numbered permission prompt (esc = safe default) |
| `agentic/usagebar` | Model / tokens / cost / context bar with thresholds |
| `examples/…` | Runnable apps: `go run ./examples/chat` (mock agentic session), `./examples/demo`, `./examples/statusbar` |

Phases 0–3 are complete — see [PLAN.md](PLAN.md). Next: the Pi coding agent
TUI built on this library (Phase 4, separate app repo), then the custom
streaming markdown renderer (Phase 5).

## Design pillars

- **Role-based theming** — components consume ~18 semantic color roles, never
  raw colors; consistency is structural, not disciplinary.
- **Pure MVU + glue** — components are plain bubbletea v2 models; opt-in
  utilities handle focus, layout, and delegation. No framework, no DSL.
- **Constraint layout** — `ultraviolet/layout`'s Cassowary solver behind a
  small facade: `Len/Min/Max/Percent/Ratio/Fill` plus `Apply` to size
  components straight from the split.
- **Agent-verifiable rendering** — every component snapshot-tests to
  plain-text goldens an agent can read in a git diff.

## Documentation

- [DESIGN.md](DESIGN.md) — architecture, decision record, component contract
- [PLAN.md](PLAN.md) — phased roadmap with exit criteria
- [AGENTS.md](AGENTS.md) — conventions for coding agents building with gotui

## Development

```sh
go build ./... && go vet ./... && go test ./...

# regenerate snapshot goldens after an intentional visual change:
go test ./... -update
```
