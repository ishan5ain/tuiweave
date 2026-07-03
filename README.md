# gotui

A modular, customizable TUI component library for Go — built on
[bubbletea v2](https://github.com/charmbracelet/bubbletea),
[lipgloss v2](https://github.com/charmbracelet/lipgloss), and
[ultraviolet](https://github.com/charmbracelet/ultraviolet) — for building
beautiful, consistent terminal frontends. Designed **agent-authored first**:
the API, conventions, and verification loop are optimized so coding agents can
reliably write TUIs with it.

**Status: pre-alpha, Phase 0 (foundation).** APIs will change freely.

## What's here

| Package | Purpose |
|---|---|
| `gotui` (root) | Semantic theme roles (`Theme`), `Dark()`/`Light()` defaults |
| `gotui/layout` | Flexbox-like layout: constraints → rectangles → component sizes |
| `gotui/snaptest` | Snapshot test harness: deterministic render-to-golden with readable diffs |

Components arrive in Phase 1+ (see [PLAN.md](PLAN.md)); agentic domain
components (chat, tool-call blocks, diff viewer, markdown) land under
`gotui/agentic` in Phase 3. The driving application is a custom TUI for the
Pi coding agent.

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
