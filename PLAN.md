# gotui — Phased Plan

gotui is an agent-authored-first TUI component library for Go, built on
bubbletea v2, lipgloss v2, and ultraviolet. The full rationale behind every
decision below lives in [DESIGN.md](DESIGN.md). This file is the execution
roadmap.

**North star:** a family of agentic tools sharing one visual language, with a
custom Pi coding agent TUI (JSON-RPC) as the driving app — and, ultimately,
Pi iterating on its own TUI using this library.

---

## Phase 0 — Foundation *(complete)*

Repo bootstrap and the three load-bearing pieces every later phase depends on.

- [x] Git repo, Go module (`github.com/ishansain/gotui`), charm v2 dependency stack
- [x] PLAN.md, DESIGN.md, AGENTS.md, README.md
- [x] **Theme package** (root `gotui`): semantic role struct, Dark + Light defaults
- [x] **Layout wrapper** (`gotui/layout`): facade over `ultraviolet/layout`
      (flexbox-like constraints → rectangles), `Sizable` convention, `Apply` helper
- [x] **Snapshot harness v0** (`gotui/snaptest`): deterministic render-to-golden,
      plain-text + styled artifacts, `-update` flag, readable diffs

**Exit criteria (met):** `go build ./... && go vet ./... && go test ./...`
green; an agent reading AGENTS.md + DESIGN.md can write a conforming component.

## Phase 1 — Prove the conventions *(complete)*

One primitive built end-to-end to validate every convention before mass
component production.

- [x] First primitive: **statusbar** (theme roles, `SetSize`, MVU, golden tests)
- [x] First runnable example under `examples/statusbar/`
- [x] CI (GitHub Actions): build, vet, test, examples compile
- [x] Extend snaptest: styled cell-grid dump via UV buffer (`SnapCells`,
      role-labeled style runs, e.g. `" gotui " [fg=TextInverted bg=Accent bold]`)
- [x] AGENTS.md grows from skeleton to real recipes based on what the statusbar taught

**Exit criteria (met in Phase 2):** an agent can clone the statusbar pattern to
produce a second primitive without human correction — the Phase 2 component set
was produced this way.

## Phase 2 — Core primitives + glue *(complete)*

The generic component set and the opt-in utilities that target known agent
failure modes.

- [x] Components: `list`, `textinput`, `viewport`, `spinner`, `table`, `help`/keybar
- [x] `dialog`/modal + `overlay` compositing (UV cell buffers, library-internal only)
- [x] Glue utilities: focus manager (index-based value type — component
      pointers go stale across MVU copies, a bug the demo caught)
- [x] Golden tests throughout; examples: statusbar + the multi-pane demo cover
      every component (dedicated per-component examples deferred — recipes in
      AGENTS.md point into the demo)
Deferred out of this phase: `textarea`, scrollbars, list filtering,
per-component standalone examples — all picked up in **Phase 3.5** below.

**Exit criteria (met):** a non-trivial multi-pane demo app composed purely from
gotui components, laid out via `gotui/layout`, fully snapshot-tested —
`examples/demo` (list + viewport + textinput + help + statusbar + spinner +
modal dialog, focus cycling, golden-tested screens).

## Phase 3 — Agentic components *(complete)*

The domain layer (`gotui/agentic/...`) built on the primitives — the actual
reuse target for the family of tools.

- [x] `agentic/markdown`: glamour v2-backed, **behind the swappable `Renderer`
      interface**, theme-role style mapping (incl. chroma), per-width renderer
      cache; streaming re-render with caching in the chat assistant cell
- [x] `agentic/chat`: cell-based transcript (user/assistant/note cells +
      `CellFunc` adapter), streaming via mutate-and-`Invalidate`, auto-follow
      with unstick/re-stick
- [x] `agentic/toolcall`: status-aware collapsible block (pending/running/✓/✗,
      output cap with "+N more")
- [x] `agentic/diffview`: styled unified diffs — `Sprint` for inline cells,
      `Model` for scrollable panes
- [x] `agentic/permission`: vertical numbered options, number quick-select,
      esc = safest (last) option
- [x] `agentic/usagebar`: model/tokens/cost/context with threshold styling

**Exit criteria (met):** `examples/chat` renders a full mock agentic session —
streaming markdown, permission-gated tool call, inline diff, live usage bar —
from goldens (mid-stream permission overlay + final screen) and verified live
in a pty.

## Phase 3.5 — Deferred primitives *(complete)*

The Phase 2 deferrals, done library-side **before** the Pi app starts so
Phase 4 begins against a complete component set.

- [x] **`gotui/scrollbar`**: standalone package —
      `Vertical(theme, height, total, visible, offset)` (track `BorderMuted`,
      thumb `Border`) + `For(theme, s)` over the `Scrollable` interface
      (`TotalLines/VisibleLines/YOffset`), implemented by viewport, list,
      table, chat, and diffview. Apps place the bar as a 1-column layout
      segment — components stay unaware of it.
- [x] **List filtering**: `SetFilter(query)` — case-insensitive substring,
      display-only state. `Selected()` keeps returning the index into the
      **original** items (mapping maintained internally); navigation,
      windowing, and scrollbar stats operate in filtered space. The app owns
      the filter input UI.
- [x] **`gotui/textarea`**: logical lines (`[][]rune`) with soft-wrap display
      (character-level, deterministic), cursor mapping through wrapped rows
      including the exact-multiple wrap-boundary case, visual-row up/down,
      `ContentHeight()` for content-driven growth, `InsertString` (also the
      app's newline hook, e.g. alt+enter), paste with embedded newlines via
      `Key.Text`. Tier-1 keys: arrows/home/end/ctrl+a/e, backspace/delete
      (joining at boundaries), ctrl+u/k/w. Deferred tier 2: undo, kill ring,
      selections, IME, wide-rune (CJK) column math.
- [x] **Example coverage gaps closed**: `examples/table` (mock git-status:
      table + diffview.Model + scrollbars on both panes, golden-tested);
      `examples/chat` input swapped to textarea (enter sends, alt+enter
      newline, input grows to 4 rows via `ContentHeight`). Every public
      component now appears in ≥1 example app.
- [x] Recipes for scrollbar/filtering/textarea; docs updated.

**Exit criteria (met):** textarea golden-tested at wrap edges and
paste-with-newlines; every scrolling component wears a scrollbar via one
interface; a filtered list reports original-index selection under test;
every public component appears in at least one example app; suite green.

## Phase 4 — Pi TUI parity *(next — separate app repo)*

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
      `markdown.Renderer` interface (no app-code changes)
- [ ] Theme gallery + palette-derivation helper (base palette → roles)
- [ ] Second tool built on gotui (validates the family-of-tools goal)
- [ ] API review and freeze toward a tagged v1; track bubbletea v2 / lipgloss v2
      out of beta, ultraviolet API settling

**Exit criteria:** two shipping tools on one visual language; v1 tag.

---

## Standing rules (all phases)

- Every component ships with: golden tests, coverage in a runnable example
  app, and an AGENTS.md recipe.
- Colors only ever come from `gotui.Theme` roles — never literals in components
  or examples.
- Public API stays string/lipgloss-based; ultraviolet types appear only in
  `gotui/layout` geometry (`layout.Rect`) and library internals.
- Dependency policy: bubbletea v2 / lipgloss v2 are beta, ultraviolet is v0 —
  pin versions, upgrade deliberately, keep the direct UV surface thin.
