# gotui — Design

A modular, customizable TUI component library for Go, built on **bubbletea v2**,
**lipgloss v2**, and **ultraviolet**, for building beautiful, consistent
terminal frontends — designed so that humans and coding agents can compose
them reliably.

This document records the architectural decisions, the rationale behind each,
and the conventions that follow from them. [PLAN.md](PLAN.md) is the execution
roadmap; [AGENTS.md](AGENTS.md) is the distilled rulebook agents load when
writing code with the library.

The north star is a small, composable Go vocabulary for terminal interfaces:
make layout, appearance, interaction, and verification predictable for humans
and coding agents. Applications own orchestration and domain state; gotui owns
reusable primitives and optional domain kits. Agentic UIs are a demanding
proving ground, but they do not define the core API.

---

## 1. Decision record

| # | Decision | Choice |
|---|---|---|
| D1 | Audience | **Human- and agent-friendly by design** — both can understand and compose the system |
| D2 | Scope | **Component library**, not a framework |
| D3 | Relationship to bubbles v2 | **From scratch** — uniform conventions from day one |
| D4 | API paradigm | **Pure MVU + opt-in glue utilities** |
| D5 | Theming | **Role-based theme struct** (~18 semantic roles) |
| D6 | Ultraviolet usage | **`uv/layout` wrapped for containers; cell buffers library-internal only** |
| D7 | Markdown | **Glamour now, behind a swappable interface; custom streaming renderer later** |
| D8 | Domain layering | **Domain packages live above generic primitives**; agentic components are the first such layer |
| D9 | Verification | **First-class snapshot harness early; pty capture as a thin script later** |
| D10 | Agent docs | **AGENTS.md + compact catalog + CI-compiled examples** |
| D11 | Driving app | **Applications validate the library** — Pi is the first demanding consumer, not the core boundary |
| D12 | Generality | **Domain-neutral core, layered domain packages** — reusable interaction patterns stay portable |
| D13 | Composition quality | **A small design grammar** — consistency comes from shared contracts and patterns, not visual sameness |
| D14 | Agent operability | **Optional semantic inspection and actions** — make composed UIs understandable and controllable without owning an app framework |

### D1 — Human- and agent-friendly by design

The library is optimized so humans and coding agents can reliably generate
correct, coherent TUIs: a constrained but expressive API surface, explicit
conventions, hard-to-misuse components, runnable examples, and deterministic
verification. Agents are a demanding consumer of this design, not the sole
audience or the reason the library is domain-specific. The verification loop
(D9) and agent-facing docs (D10) make the design grammar observable and usable
by an agent: if an agent cannot understand or inspect what it composed, it
cannot iterate effectively.

### D2/D3 — Component library, from scratch

A better `bubbles`, not a framework: themed widgets that drop into any
bubbletea v2 app. We do not own the event loop, routing, or app shell.
Components are written fresh (not wrappers over bubbles) so theming, sizing
(`SetSize`), and focus conventions are uniform from day one. The cost is
owning deceptively hard internals (list virtualization, text editing, unicode
width); the benefit is that consistency is structural rather than retrofitted.

### D4 — Pure MVU + glue utilities

Components are plain bubbletea v2 models. A declarative component-tree DSL was
rejected for two reasons: (a) owning message routing/focus/layout *is* a
framework, contradicting D2; (b) agents have deep training priors on bubbletea
idioms and zero priors on a novel DSL — **familiar-verbose beats novel-terse**
for agent-friendly code. Known MVU agent failure modes (forgetting to reassign
the model after `Update`, dropping a `Cmd`, unwired focus) are addressed by
opt-in glue utilities — a focus manager, layout helpers, delegation helpers —
plus explicit rules in AGENTS.md, not by hiding the loop.

"Pure MVU" applies to core components. Streaming transcript cells and
tool-call blocks are a deliberate second contract: mutable content objects
owned by the application, invalidated by the transcript when their rendered
content changes. Keeping that distinction explicit avoids treating
pointer-based streaming as a violation of the core component model.

### D5 — Role-based theme struct

One flat Go struct whose fields are **semantic roles** (`Surface`, `TextMuted`,
`Accent`, `Danger`, `BorderFocused`, …), not per-component knobs. Components
derive their lipgloss styles exclusively from roles and never touch raw colors.

- Rejected: flat per-component style struct — grows unboundedly (every
  component × every state wants a field) and merely *enables* consistency
  where roles *enforce* it.
- Rejected: full web-style token architecture (primitive ramps → roles →
  derived styles) — over-built for the terminal's tiny style space
  (fg/bg/bold/italic/underline/borders). lipgloss's colorprofile writer
  already handles truecolor → 256 → 16 degradation; only the semantic layer
  is worth building.
- Semantic roles are also the vocabulary agents already speak (Tailwind,
  Radix, shadcn saturate their training data) — "destructive actions use
  `Danger`" is far more reliable than choosing among 40 color fields.

Roles live in the root `gotui` package (see §3). Component-level overrides,
if ever needed, arrive later as an escape hatch (`WithStyle(fn)`), not as the
primary API.

### D6 — Ultraviolet: layout at the surface, buffers inside

Division of labor in the stack: bubbletea v2 is the runtime (`View() tea.View`
is string content — there is no direct buffer interface for models); lipgloss
is styling + string composition; ultraviolet is the engine under both — cell
grid, damage-tracking diff renderer, geometry, and a Cassowary constraint
solver (`ultraviolet/layout`, a ratatui port).

Key facts that shaped the decision:

- **Cell-level diff rendering comes for free** through bubbletea v2 — it
  parses the View string into a UV buffer and diffs it. Optimized wire output
  alone never justifies touching UV directly.
- **`uv/layout` is just geometry.** Solving constraints yields
  `[]uv.Rectangle`; nothing forces cell buffers. Rectangles can size ordinary
  string-rendering components.
- `uv/layout` is essentially one-dimensional flexbox with nesting:
  `Len/Min/Max/Percent/Ratio/Fill` constraints plus CSS-`justify-content`-like
  `Flex` strategies. **The layout engine we planned to build already exists.**

Therefore: `gotui/layout` wraps `uv/layout` as the flexbox-like container
system — geometry only in the public API. The current wrapper is intentionally
thin and uses public type aliases; this keeps call sites simple but means the
Ultraviolet type identity is not a perfectly sealed compatibility boundary.
UV cell buffers are reserved as a
**library-internal** tool for overlay/modal/z-order compositing (Phase 2),
where lipgloss string-splicing is genuinely bad. The "agents don't know novel
APIs" argument does not apply internally — agents consuming gotui never see UV.

Since ultraviolet is v0.x and bubbletea v2 is beta, the thinner our direct UV
surface, the cheaper every upgrade. `layout.Rect` aliases `uv.Rectangle`
(which is `image.Rectangle`). Consumers import `gotui/layout`, but the aliases
mean Ultraviolet type identity can still affect compatibility during upgrades.

### D7 — Markdown: glamour now, custom later

Pi parity requires streaming markdown (progressive assistant responses, code
blocks, syntax highlighting) — the hardest rendering problem in the driving
app. Strategy: ship parity on **glamour**, but **behind our own interface**,
so a purpose-built incremental renderer can replace it later without touching
app code.

*Implemented in Phase 3* as `markdown.Renderer` in `agentic/markdown`:
glamour v2 with all theme roles mapped into its stylesheet (including chroma
syntax-highlighting colors), a renderer cached per wrap width, and streaming
handled by re-rendering the in-progress message — the chat assistant cell
caches by (source length, width) today; the roadmap replaces that with an
explicit source revision so replacements cannot reuse a same-length render.
`Sprint`
degrades to raw source on error; transcripts must not fail on bad markdown.
The revision must change for replacements as well as appends; append-only
streaming is a useful optimization, not a hidden correctness requirement.

### D8 — Domain packages live above generic primitives

Generic primitives (`list`, `textinput`, `viewport`, `dialog`, …) live in
top-level packages. Domain components such as chat transcripts, tool-call
blocks, diff viewers, and permission prompts live in `gotui/agentic/...` and
are built on those primitives.

Agentic components are reusable because they consume plain Go types, not
agent-backend clients. The same layering can support future domains such as
database consoles, file browsers, monitoring dashboards, and developer tools.
Backend clients, persistence, and application workflows live in app repos.

### D9 — Verification: snapshot harness first, capture later

`gotui/snaptest` is a first-class deliverable built in Phase 0, because
components are string-rendering with explicit sizes and can therefore be
rendered **deterministically** — no pty, no event loop, no timing flake:

- `Snap`: plain-text golden (ANSI stripped) — layout/content, perfectly
  legible in a git diff, the feedback format agents thrive on.
- `SnapCells` (since Phase 1): the rendered view parsed into a UV cell grid
  and dumped as role-labeled style runs, e.g.
  `" gotui " [fg=TextInverted bg=Accent bold]` — the artifact for asserting
  *which role* styles what. Combining graphemes are preserved in the dump;
  `WithRoles(theme)` maps colors back to role names.
- `SnapStyled`: raw-ANSI golden for byte-exact styling regressions.
- `-update` flag regenerates; failures print readable line diffs.
- Doubles as the library's own regression suite.

Overlay composition uses the same grapheme-preserving cell boundary as the
snapshot harness, so modal content does not lose combining marks when it
replaces padded base cells.

Rejected as the core loop: `x/exp/teatest` — experimental dependency,
whole-program granularity, ANSI-laden goldens agents can't usefully read.
(Fine for occasional program-level smoke tests.) Live pty capture (vhs /
`tmux capture-pane`) is deferred to a later application-validation phase as a
small script once there is a real app to capture — a tool, not a platform.
Snapshot limits are accepted and documented: the current harness verifies
`View`, not the full `Update` loop. Deterministic interaction scenarios are the
planned complement for message sequences, commands, and state transitions.

### D10 — Agent docs: AGENTS.md + CI-compiled examples

For a library whose primary consumer is an agent, the conventions docs **are the
product's user interface**. AGENTS.md stays compact: explicit rules
("colors come from theme roles, never literals", "size from layout rects",
"reassign model, collect cmds") plus pointers into the runnable example apps
(`examples/statusbar`, `examples/demo`, `examples/chat`). Drift is
neutralized by making examples real packages that compile and snapshot-test
in CI — API changes that stale the examples break the build. The compact
`AGENT-CATALOG.md` is the routing index; it points to the detailed recipes
without forcing every task to load the entire conventions file.

### D11 — Applications validate the library

A custom Go TUI for the Pi coding agent is the first demanding consumer using
Pi's JSON-RPC interface. It should drive real API improvements, but the app and
its protocol remain outside gotui. Other applications, including non-agentic
tools, are equally important validation targets; gaps found while building any
of them flow back as gotui issues.

### D12 — Generality through layered scope

The library has five conceptual layers:

1. **Foundations:** theme roles, layout, sizing, rendering, and snapshot testing.
2. **Interaction primitives:** inputs, lists, tables, viewports, focus, dialogs,
   scrolling, and other reusable terminal behaviors.
3. **Composition utilities:** frames, panes, toolbars, menus, tabs, and other
   structural helpers that combine primitives without owning a product domain.
4. **Domain packages:** agentic or other specialized components built on the
   lower layers.
5. **Applications:** event routing, backend protocols, persistence, and
   product-specific workflows.

When deciding where code belongs, prefer the lowest layer that can express the
behavior without introducing domain assumptions. Add a component to the core
when it represents a recurring terminal interaction pattern, not merely because
one application currently needs it.

The initial Phase 4 framing slice is deliberately a pure composition utility,
not a component-tree abstraction: `gotui/frame` accepts a theme, content, and
available width, then returns strings for the application to compose. Panels
have natural height and exact requested width; the app remains responsible for
height allocation, state, and event routing. This keeps decoration reusable
without hiding the root model's layout or MVU decisions.

### D13 — Composition quality: a small design grammar

The library should provide a recognizable vocabulary for terminal UI design:
theme roles, layout constraints, sizing contracts, focus behavior, selection
semantics, scrolling statistics, modal routing, and common empty/loading/error
states. Components may be visually distinct, but they should behave predictably
when composed with one another.

This is how gotui balances customization and consistency:

- **Consistency comes from contracts and shared semantics**, not from forcing
  every application into one visual arrangement.
- **Customization comes from composition and explicit extension points**, not
  from every application reimplementing focus, sizing, truncation, or state
  handling locally.
- **Agent-friendliness comes from legibility**: package names, APIs, recipes,
  examples, and snapshots should make the right composition discoverable.

For a proposed abstraction, ask whether it adds a reusable word to this grammar
or merely hides application-specific decisions behind a new name.

Agent-friendliness has two sides. The authoring side is covered by package
names, recipes, examples, and deterministic snapshots. The operating side is
the next priority: an agent should be able to inspect semantic state, discover
available actions, and replay an interaction without parsing ANSI output or
pretending to be a human typing at coordinates.

### D14 — Optional agent operability

Coding agents increasingly do more than generate source: they run applications,
inspect state, approve or deny operations, and iterate from test or runtime
feedback. `View() string` is excellent for humans and visual goldens, but it is
not a stable machine interface.

The next layer should therefore provide optional, framework-free capabilities:

- **Inspection**: stable component IDs, bounds, focus, selection, scroll state,
  visible labels, lifecycle state, and child relationships.
- **Actions**: stable action IDs such as `list.next`, `dialog.confirm`, or
  `transcript.scroll_bottom`, independent of key bindings.
- **Scenarios**: deterministic message sequences with rendered and semantic
  checkpoints, emitted commands, and state transitions.

These capabilities should be interfaces or small utility packages, not a
component tree or application event loop. MCP, JSON-RPC, or a particular coding
agent can adapt to them at the application boundary. Core gotui should expose
the vocabulary without owning the transport or backend protocol.

The first scenario implementation deliberately stays smaller than the full
operability layer. `snaptest.RunScenario` accepts explicit, named messages,
captures the initial and post-message `tea.View.Content`, and records whether
each update emitted a command. It does not execute commands implicitly: timers,
I/O, batching, and command-to-message policies belong in the application test
where they can be made deterministic. This gives the library a useful replay
boundary without pretending that an opaque `tea.Cmd` has a stable identity.

The initial inspection slice is similarly data-only: `gotui/inspect.Node`
represents IDs, bounds, focus, status, selection, scrolling, attributes, and
children; component reports provide local state while the application assembles
and positions the tree. Inspection intentionally omits raw input values and
rendered content by default so the application remains responsible for privacy
and disclosure decisions.

Action metadata follows the same ownership boundary. Components advertise local
IDs and accept `inspect.ActionMsg`; `inspect.Bind` qualifies reported IDs with
the application-owned node ID. Tree-level dispatch remains application-owned,
so gotui does not invent a router or a global focus model.

---

## 2. How the decisions interlock

The snapshot harness works *because* components are string-rendering with
explicit sizes (D4); sizes come from `uv/layout` rectangles (D6); style
assertions in goldens are checkable *because* all color flows through theme
roles (D5); and AGENTS.md rules are enforceable *because* the harness gives CI
a way to catch violations (D9→D10). Each choice load-bears for the others.

## 3. Package architecture

```
github.com/ishansain/gotui
├── gotui            (root) Theme roles, Dark/Light defaults
├── layout/          facade over ultraviolet/layout: Rect, constraints,
│                    Vertical/Horizontal, Sizable, Apply
├── snaptest/        snapshot test harness (Snap, SnapCells, SnapStyled, -update)
├── inspect/         optional semantic UI tree (IDs, bounds, focus, state)
├── action/          shared stable-ID selectable-action definitions
├── frame/            domain-neutral framing and decoration helpers
├── tabs/             focusable sibling-view navigation
├── menu/             focusable vertical action choices
├── toolbar/          focusable horizontal action choices
├── splitpane/        width-aware sibling-view composition
├── stack/             width-aware vertical chrome composition
├── progress/          passive task progress indicator
├── toggle/            focusable boolean setting control
├── button/            focusable single-action control
├── palette/           bounded command-palette foundation
├── autocomplete/      app-owned input plus bounded suggestion window
├── line/              width-aware single-row composition helpers
├── statusbar/  list/  viewport/  textinput/  textarea/  table/  help/  spinner/
│                    generic primitives, one package each
├── scrollbar/       one-column bar for anything implementing Scrollable
├── dialog/          modal confirm box (ResultMsg pattern)
├── overlay/         cell-space compositing — Place/Center (UV internal)
├── focus/           copy-safe (index-only) tab-order manager
├── agentic/          optional domain layer built on the primitives
│   ├── markdown/    Renderer interface + glamour v2 implementation
│   ├── chat/        cell-based streaming transcript (User/Assistant/Text cells)
│   ├── toolcall/    status-aware collapsible tool-call block (chat cell)
│   ├── diffview/    styled unified diffs (Sprint inline + scrollable Model)
│   ├── permission/  numbered permission prompt (ResultMsg pattern)
│   └── usagebar/    model/tokens/cost/context bar
├── examples/
│   ├── statusbar/   canonical single-component wiring
│   ├── demo/        multi-pane app (Phase 2 exit criterion), golden-tested
│   ├── chat/        mock agentic session (Phase 3 exit criterion), golden-tested
│   ├── frame/       framing/decorations/tabs/menu/toolbar/splitpane/stack/progress/toggle/button composition example, golden-tested
│   ├── palette/     filtered command discovery and activation example, golden-tested
│   ├── ops/         non-agentic operations-console pressure test, golden-tested
│   ├── browser/     filterable file list + scrollable preview, golden-tested
│   └── table/       git-status mock: table + diffview + scrollbars, golden-tested
├── .github/workflows/ci.yml   build + vet + test + tidy check
├── AGENTS.md        the agent-facing rulebook
├── DESIGN.md        this document
└── PLAN.md          phased roadmap
```

Naming follows the flat-per-component charm convention (`gotui/list`, not
`gotui/components/list`) because it is what agents expect from `bubbles`.

## 4. Component contract

Every gotui component:

1. Is a plain value type with a `New(...)` constructor taking `gotui.Theme`
   (plus component-specific config).
2. Implements MVU: `Update(tea.Msg) (Self, tea.Cmd)` returning its own
   concrete type (not `tea.Model`), and `View() string`.
3. Implements `SetSize(width, height int)` (the `layout.Sizable` interface).
   Components are bounded by default and render exactly within that box.
   Documented exceptions implement the optional `layout.SizeModeAware`
   contract: `SizeWidthBounded` for width-constrained natural-height panels,
   or `SizeIntrinsic` for components that ignore the assigned box.
4. Derives every style from theme roles at construction/update time — no
   color literals anywhere.
5. Exposes focus as `Focus()`/`Blur()` where interactive.
6. Ships with golden tests (snaptest), coverage in a runnable example app,
   and an AGENTS.md recipe entry.

The root app model composes components, splits its area with `gotui/layout`
on `tea.WindowSizeMsg`, delegates messages, and wraps the final composed
string in `tea.NewView` — standard bubbletea v2, nothing hidden.

Pure composition helpers such as `frame` are not components and therefore do
not implement `Update` or `SetSize`; they receive explicit dimensions from the
app and are covered by focused rendering goldens plus a runnable composition
example.

The current exceptions are `spinner` (a single intrinsic-size glyph) and the
modal panels (`dialog`, `permission`), which are width-bounded and render at
natural content height. `layout.SizeModeOf` defaults ordinary components to
`SizeBounded`, so applications can inspect the policy without special-casing
package names.

**Chat cells are a second, smaller contract** (`agentic/chat.Cell`):
`Render(width int) string`. Cells are *pointers* the app keeps and mutates
as a session progresses (streaming deltas, tool status); the transcript
re-renders them on `Append`/`SetSize`/`Invalidate`. Cells are content, not
components — no `Update`, no focus. Anything can adapt in via `CellFunc`.

## 5. Theme roles

Defined in the root package (`gotui.Theme`), all fields `image/color.Color`
(lipgloss v2's native currency). Role semantics:

| Group | Roles | Used for |
|---|---|---|
| Surfaces | `Surface`, `SurfaceRaised`, `SurfaceSunken` | app background; panels/dialogs; input wells & code blocks |
| Text | `Text`, `TextMuted`, `TextFaint`, `TextInverted` | primary; secondary/placeholders/help; tertiary/separators; text on accent fills |
| Intent | `Accent`, `AccentMuted`, `Success`, `Warning`, `Danger`, `Info` | brand/interactive; subdued accent fills; semantic states |
| Chrome | `Border`, `BorderFocused`, `BorderMuted` | default borders; focused element; subtle dividers |
| Selection | `SelectionBg`, `SelectionFg` | selected rows/items |

18 roles. Adding a role is an API event requiring justification that no
existing role covers the semantics — this is the mechanism that keeps the
theme API from growing per-component.

## 6. Dependency policy

- `charm.land/bubbletea/v2` (v2.0.8), `charm.land/lipgloss/v2` (v2.0.5) —
  beta; pinned, upgraded deliberately. Note: charm's v2 modules live on
  `charm.land` vanity paths.
- `github.com/charmbracelet/ultraviolet` — v0, pseudo-versioned; expect
  churn; confined to `layout` (geometry), `overlay` (compositing), and
  `snaptest` (cell-grid parsing). Nothing else imports it.
- `charm.land/glamour/v2` (v2.0.1) — fenced behind `markdown.Renderer`;
  only `agentic/markdown` imports it.
- No dependency on `x/exp/teatest` in the core verification loop.

## 7. Open questions

- **Module path / publication:** currently `github.com/ishansain/gotui`,
  private-by-circumstance. License (MIT recommended) and publication decision
  before any external consumer.
- **Text editing depth:** `textarea` now has an initial tier-2 slice for
  logical-rune selections, select-all/replacement, and bounded undo/redo on top
  of tier-1 soft wrap, visual-row movement, line joins, and paste. Its display
  geometry now uses grapheme clusters and terminal-cell widths for wide and
  combining characters. It also has whitespace-delimited word movement and a
  bounded model-owned kill/yank ring (`ctrl+w/u/k`, `alt+y`; `ctrl+y` remains
  redo). Consecutive kills coalesce in direction-aware order, and repeated
  yanks rotate through older kills without duplicating the inserted text. Still
  open: IME and further editor-specific commands. The current cell-width audit
  covers ANSI/lipgloss components plus grapheme-safe table, snapshot, and
  overlay cell paths.
  across text-bearing components. `textinput` now has a corresponding initial audit
  slice for cell-aware prompt, placeholder, cursor, and horizontal-window
  geometry. The initial shared/agentic audit also clamps chat-cell headers and
  tool-call output to their assigned width. Dialog and permission panels now apply the same
  bounded-width rule to their outer frame and inner action/content rows.
- **Streaming markdown renderer design:** incremental block parser vs
  full-document reparse with damage hints — decide when glamour's limits are
  measured, not guessed.
- **Inspection/action schema hardening:** evolve the initial data-only schema
  without turning gotui into a framework; preserve stable IDs as APIs mature.
- **Scenario format expansion:** the initial Go-native plain golden is in place;
  decide whether a separate machine-readable event format is needed for richer
  command outcomes and replay.
- **Agentic session ledger:** the initial slice provides optional cell IDs,
  lifecycle states, lookup/replacement, retries, cancellation, and source
  revisions. A future ledger must decide persistence, event ordering, resume,
  and replay across sessions.
