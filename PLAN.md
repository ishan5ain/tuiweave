# gotui — Phased Plan

gotui is an agent-friendly-by-design, general-purpose TUI component library for Go,
built on bubbletea v2, lipgloss v2, and ultraviolet. The full rationale behind
every decision below lives in [DESIGN.md](DESIGN.md). This file is the
execution roadmap.

**North star:** provide a small, composable Go vocabulary for terminal
interfaces that makes layout, appearance, interaction, and verification
predictable for humans and coding agents. Agentic tools, including a custom Pi
coding agent TUI, are demanding consumers and reference applications rather than
the defining scope.

---

## Immediate priority — agent-operable foundations

The current library is agent-friendly to author: its contracts, examples, and
goldens help an agent discover and compose components. The next highest-value
step is making the resulting UI agent-friendly to inspect, test, and operate.
This is a narrow capability layer, not a component-tree framework, and should
be completed before broadening the composition catalog too far.

- [x] **Size modes (initial slice)**: make bounded, width-bounded, and
      intrinsic behavior explicit through `layout.SizeMode`; ordinary
      components default to bounded, while spinner, dialog, and permission
      declare their documented exceptions.
- [x] **Interaction scenarios (initial slice)**: step a model through explicit named messages,
      capture initial and post-message views, record command emission, and
      snapshot readable checkpoints. Do not execute opaque commands implicitly;
      applications own deterministic command policies.
- [x] **Semantic inspection (initial slice)**: define the data-only
      `gotui/inspect` representation for stable IDs, bounds, focus, selection,
      scroll state, labels, lifecycle state, attributes, and children; add
      reports for the core scrolling/input components and chat.
- [x] **Semantic actions (initial slice)**: define stable action IDs independent
      of key strings, expose them in inspection nodes, and let components handle
      local `inspect.ActionMsg` intents such as `next`, `clear`, and `confirm`.
      `inspect.Bind` qualifies IDs for tree consumers; application routing stays
      outside the core.
- [x] **Agentic session identity (initial slice)**: add optional stable IDs,
      lifecycle states, transcript lookup/replacement, assistant source
      revisions, and tool-call retry/cancel semantics without coupling to a
      backend protocol.
- [x] **Permission provenance (initial slice)**: evolve approval data beyond
      title/body to describe the exact operation, scope, impact, reversibility,
      and policy context while keeping MCP/JSON-RPC adapters outside the core.
- [x] **Agent catalog (initial slice)**: provide a compact, discoverable index
      of packages, recipes, setup sequences, common mistakes, and canonical
      examples in `AGENT-CATALOG.md`.
- [x] **Streaming invalidation correctness**: use an explicit source revision
      for assistant cells so replacement content cannot reuse a same-length
      cached render.

**Initial foundation exit criteria (met):** an application can expose a
deterministic semantic snapshot and action list alongside its human-readable
view; an interaction scenario can replay a focus, selection, scrolling, and
modal flow; and a streamed session can update, cancel, and replace content by
stable identity.

Remaining work in this section extends the foundation with contract hardening
and provenance/catalog refinement.

This work should remain optional at the component boundary. Applications still
own orchestration, persistence, transport, and policy enforcement.

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

**Exit criteria (confirmed during Phase 2):** an agent can clone the statusbar pattern to
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

The Phase 2 deferrals, done library-side **before** the Pi app starts so the
general-purpose composition work begins against a complete component set.

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

## Phase 4 — General-purpose composition layer *(after agent-operable foundations)*

Expand the library from a strong primitive set into a flexible toolkit for
composing complete custom interfaces. These additions must remain domain-neutral
and useful in at least two unrelated application types.

- [ ] Framing and decoration utilities: borders, padding, titles, separators,
      badges, and reusable framed content
- [ ] Structural components: tabs, menus, toolbars, split panes, and stacked
      headers/footers
- [ ] Common controls: toggles, buttons, progress indicators, selectable
      actions, and command-palette foundations
- [ ] Width-aware composition helpers for truncation, alignment, fill zones,
      and graceful narrow-terminal behavior
- [ ] Reference example apps beyond agentic UIs: a file browser, dashboard, or
      operations console
- [ ] Recipes and golden coverage for every new composition pattern
- [ ] Cross-component recipes showing how customization and consistency work
      together
- [ ] Use the semantic inspection, action, and scenario conventions from the
      immediate-priority workstream in the reference composition examples

**Exit criteria:** an agent can assemble a multi-view non-agentic application
from gotui without introducing local copies of common framing, navigation, or
selection behavior, then inspect and exercise that application without parsing
terminal escape sequences.

## Phase 5 — Editing and interaction depth

Make the interaction primitives robust enough for serious editors and repeated
daily use.

- [ ] Textarea tier 2: undo/redo, kill ring, selections, and richer editing
      commands
- [ ] Correct cell-width handling for wide runes and combining characters
- [ ] Autocomplete and command-palette primitives
- [ ] More explicit mouse interaction conventions where bubbletea supports them
- [ ] Focus scopes and nested modal/focus routing utilities
- [ ] Extend interaction scenarios across mouse input, focus scopes, nested
      modals, cancellation, and narrow-terminal behavior
- [ ] Audit APIs and recipes for discoverability by a coding agent starting from
      the package list and AGENTS.md

**Exit criteria:** a small editor or command-driven console can be implemented
with library primitives rather than application-specific editing machinery.

## Phase 6 — Domain kits and Pi validation

Use the expanded composition layer to validate both the original agentic goal
and the broader library boundary. The Pi frontend remains in a separate app
repo; gotui stays independent of Pi's protocol.

- [ ] Pi TUI app repo scaffolded; JSON-RPC client for Pi
- [ ] Pi TUI parity and Zentui-inspired editor/footer composition
- [ ] At least one non-agentic reference application built on gotui
- [ ] Review `agentic/` APIs and add only domain components that generalize
      across agentic applications
- [ ] Live-capture script (vhs/tmux `capture-pane`, ~50 lines, not a platform)
      so agents can see the running TUI
- [ ] Dogfood loop: Pi modifies its own TUI, verifies via snaptest goldens +
      capture script; gaps found here flow back as gotui issues

**Exit criteria:** a daily-drivable Pi TUI and at least one unrelated custom TUI
prove that the core is not coupled to the agentic domain.

## Phase 7 — Maturity and v1

- [ ] Custom streaming markdown renderer replaces glamour behind the existing
      `markdown.Renderer` interface (no app-code changes)
- [ ] Theme gallery + palette-derivation helper (base palette → roles)
- [ ] Examples and theme gallery cover multiple application categories
- [ ] Accessibility and narrow-terminal review across the component set
- [ ] Performance review for large transcripts, tables, and repeated renders
- [ ] API review and freeze toward a tagged v1; track bubbletea v2 / lipgloss v2
      out of beta, ultraviolet API settling

**Exit criteria:** multiple unrelated TUIs share the same foundations, the
public API is reviewed and documented, and gotui reaches a tagged v1.

---

## Standing rules (all phases)

- Every component ships with: golden tests, coverage in a runnable example
  app, and an AGENTS.md recipe.
- New core components must be domain-neutral; domain-specific components live
  in a clearly named subpackage.
- Prefer composition utilities over product-specific convenience components.
- New theme roles require evidence that existing semantic roles cannot express
  the state.
- Preserve the design grammar: new components should reuse established sizing,
  focus, selection, scrolling, modal, and narrow-width conventions.
- Optional inspection and action interfaces must remain framework-free; do not
  move application routing, persistence, policy, or backend protocols into core.
- Interactive components should document the semantic state and actions that
  agents and scenario tests can observe or invoke.
- Make customization explicit and local; applications should not need to fork
  or duplicate core interaction behavior to achieve a distinct visual design.
- Colors only ever come from `gotui.Theme` roles — never literals in components
  or examples.
- Public API stays string/lipgloss-based; ultraviolet types appear only in
  `gotui/layout` geometry (`layout.Rect`) and library internals.
- Dependency policy: bubbletea v2 / lipgloss v2 are beta, ultraviolet is v0 —
  pin versions, upgrade deliberately, keep the direct UV surface thin.
