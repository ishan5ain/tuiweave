# gotui — Phased Plan

gotui is an agent-authored-first TUI component library for Go, built on
bubbletea v2, lipgloss v2, and ultraviolet. The full rationale behind every
decision below lives in [DESIGN.md](DESIGN.md). This file is the execution
roadmap.

**North star:** a family of agentic tools sharing one visual language, with a
custom Pi coding agent TUI (JSON-RPC) as the driving app — and, ultimately,
Pi iterating on its own TUI using this library.

---

## Phase 0 — Foundation *(current)*

Repo bootstrap and the three load-bearing pieces every later phase depends on.

- [x] Git repo, Go module (`github.com/ishansain/gotui`), charm v2 dependency stack
- [x] PLAN.md, DESIGN.md, AGENTS.md, README.md
- [x] **Theme package** (root `gotui`): semantic role struct, Dark + Light defaults
- [x] **Layout wrapper** (`gotui/layout`): facade over `ultraviolet/layout`
      (flexbox-like constraints → rectangles), `Sizable` convention, `Apply` helper
- [x] **Snapshot harness v0** (`gotui/snaptest`): deterministic render-to-golden,
      plain-text + styled artifacts, `-update` flag, readable diffs

**Exit criteria:** `go build ./... && go vet ./... && go test ./...` green;
an agent reading AGENTS.md + DESIGN.md can write a conforming component.

## Phase 1 — Prove the conventions

One primitive built end-to-end to validate every convention before mass
component production.

- [x] First primitive: **statusbar** (theme roles, `SetSize`, MVU, golden tests)
- [x] First runnable example under `examples/statusbar/`
- [x] CI (GitHub Actions): build, vet, test, examples compile
- [x] Extend snaptest: styled cell-grid dump via UV buffer (`SnapCells`,
      role-labeled style runs, e.g. `" gotui " [fg=TextInverted bg=Accent bold]`)
- [x] AGENTS.md grows from skeleton to real recipes based on what the statusbar taught

**Exit criteria:** an agent can clone the statusbar pattern to produce a second
primitive without human correction.

## Phase 2 — Core primitives + glue

The generic component set and the opt-in utilities that target known agent
failure modes.

- [x] Components: `list`, `textinput`, `viewport`, `spinner`, `table`, `help`/keybar
- [x] `dialog`/modal + `overlay` compositing (UV cell buffers, library-internal only)
- [x] Glue utilities: focus manager (index-based value type — component
      pointers go stale across MVU copies, a bug the demo caught)
- [x] Golden tests throughout; examples: statusbar + the multi-pane demo cover
      every component (dedicated per-component examples deferred — recipes in
      AGENTS.md point into the demo)
- [ ] Deferred to later phases: `textarea` (multi-line editing), scrollbars,
      list filtering, per-component standalone examples

**Exit criteria (met):** a non-trivial multi-pane demo app composed purely from
gotui components, laid out via `gotui/layout`, fully snapshot-tested —
`examples/demo` (list + viewport + textinput + help + statusbar + spinner +
modal dialog, focus cycling, golden-tested screens).

## Phase 3 — Agentic components

The domain layer (`gotui/agentic/...`) built on the primitives — the actual
reuse target for the family of tools.

- [ ] `MarkdownView`: glamour-backed, **behind a swappable interface**, theme-role
      style mapping, streaming via re-render + completed-block caching
- [ ] Chat message list (streaming assistant/user turns)
- [ ] Tool-call block (collapsible, status-aware)
- [ ] Diff viewer
- [ ] Permission prompt
- [ ] Status/usage bar (model, tokens, cost)

**Exit criteria:** a mock-backed chat demo renders a full agentic session
(streaming markdown, tool calls, diffs) from goldens.

## Phase 4 — Pi TUI parity (separate app repo)

The driving app. The JSON-RPC client lives in the **app repo**, not gotui —
the library stays agent-tool-agnostic; agentic components consume plain Go types.

- [ ] Pi TUI app repo scaffolded; JSON-RPC client for Pi
- [ ] Full parity with the stock Pi TUI, built on gotui
- [ ] Live-capture script (vhs/tmux `capture-pane`, ~50 lines, not a platform)
      so agents can see the running TUI
- [ ] Dogfood loop: Pi modifies its own TUI, verifies via snaptest goldens +
      capture script; gaps found here flow back as gotui issues

**Exit criteria:** daily-drivable Pi TUI; Pi lands a self-authored UI change
verified without a human looking at the screen.

## Phase 5 — Maturity

- [ ] Custom streaming markdown renderer replaces glamour behind the existing
      `MarkdownView` interface (no app-code changes)
- [ ] Theme gallery + palette-derivation helper (base palette → roles)
- [ ] Second tool built on gotui (validates the family-of-tools goal)
- [ ] API review and freeze toward a tagged v1; track bubbletea v2 / lipgloss v2
      out of beta, ultraviolet API settling

**Exit criteria:** two shipping tools on one visual language; v1 tag.

---

## Standing rules (all phases)

- Every component ships with: golden tests, a runnable example, an AGENTS.md recipe.
- Colors only ever come from `gotui.Theme` roles — never literals in components
  or examples.
- Public API stays string/lipgloss-based; ultraviolet types appear only in
  `gotui/layout` geometry (`layout.Rect`) and library internals.
- Dependency policy: bubbletea v2 / lipgloss v2 are beta, ultraviolet is v0 —
  pin versions, upgrade deliberately, keep the direct UV surface thin.
