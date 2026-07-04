# gotui — Design

A modular, customizable TUI component library for Go, built on **bubbletea v2**,
**lipgloss v2**, and **ultraviolet**, for building beautiful, consistent
terminal frontends — designed so that coding agents are its primary authors.

This document records the architectural decisions, the rationale behind each,
and the conventions that follow from them. [PLAN.md](PLAN.md) is the execution
roadmap; [AGENTS.md](AGENTS.md) is the distilled rulebook agents load when
writing code with the library.

---

## 1. Decision record

| # | Decision | Choice |
|---|---|---|
| D1 | Audience | **Agent-authored first** — coding agents (Claude Code, Pi) write the apps; humans review |
| D2 | Scope | **Component library**, not a framework |
| D3 | Relationship to bubbles v2 | **From scratch** — uniform conventions from day one |
| D4 | API paradigm | **Pure MVU + opt-in glue utilities** |
| D5 | Theming | **Role-based theme struct** (~18 semantic roles) |
| D6 | Ultraviolet usage | **`uv/layout` wrapped for containers; cell buffers library-internal only** |
| D7 | Markdown | **Glamour now, behind a swappable interface; custom streaming renderer later** |
| D8 | Domain layering | **Agentic components live in the library**, as a subpackage over generic primitives |
| D9 | Verification | **First-class snapshot harness early; pty capture as a thin script later** |
| D10 | Agent docs | **Thin skill file (AGENTS.md) + CI-compiled examples** |
| D11 | Driving app | **Pi coding agent TUI over JSON-RPC** — parity first, then specialization |

### D1 — Agent-authored first

The library is optimized so coding agents can reliably generate correct,
beautiful TUIs: constrained API surface, strong explicit conventions,
hard-to-misuse components. Humans mostly review. The meta-goal is that the Pi
coding agent iterates on **its own TUI** using this library, which raises the
bar on two things treated as afterthoughts elsewhere: the verification loop
(D9) and agent-facing docs (D10) — if an agent can't *see* what it rendered,
it can't iterate.

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
for agent-authored code. Known MVU agent failure modes (forgetting to reassign
the model after `Update`, dropping a `Cmd`, unwired focus) are addressed by
opt-in glue utilities — a focus manager, layout helpers, delegation helpers —
plus explicit rules in AGENTS.md, not by hiding the loop.

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
system — geometry only in the public API. UV cell buffers are reserved as a
**library-internal** tool for overlay/modal/z-order compositing (Phase 2),
where lipgloss string-splicing is genuinely bad. The "agents don't know novel
APIs" argument does not apply internally — agents consuming gotui never see UV.

Since ultraviolet is v0.x and bubbletea v2 is beta, the thinner our direct UV
surface, the cheaper every upgrade. `layout.Rect` aliases `uv.Rectangle`
(which is `image.Rectangle`), so consumers depend on our facade, not UV's path.

### D7 — Markdown: glamour now, custom later

Pi parity requires streaming markdown (progressive assistant responses, code
blocks, syntax highlighting) — the hardest rendering problem in the driving
app. Strategy: ship parity on **glamour**, but **behind our own interface**,
so a purpose-built incremental renderer can replace it later without touching
app code.

*Implemented (Phase 3)* as `markdown.Renderer` in `agentic/markdown`:
glamour v2 with all theme roles mapped into its stylesheet (including chroma
syntax-highlighting colors), a renderer cached per wrap width, and streaming
handled by re-rendering the in-progress message — the chat assistant cell
caches by (source length, width) so only real changes re-render. `Sprint`
degrades to raw source on error; transcripts must not fail on bad markdown.

### D8 — Agentic components live in the library

Generic primitives (`list`, `textinput`, `viewport`, `dialog`, …) in top-level
packages; agentic domain components (chat message list, tool-call blocks, diff
viewer, permission prompts) in `gotui/agentic/...`, built on the primitives.
The family-of-tools goal makes chat UIs the actual reuse target — keeping them
app-side would defer the library's whole point. Domain components consume
plain Go types (messages, tool calls); **agent-backend clients (e.g. Pi's
JSON-RPC) live in app repos**, keeping gotui agent-tool-agnostic.

### D9 — Verification: snapshot harness first, capture later

`gotui/snaptest` is a first-class deliverable built in Phase 0, because
components are string-rendering with explicit sizes and can therefore be
rendered **deterministically** — no pty, no event loop, no timing flake:

- `Snap`: plain-text golden (ANSI stripped) — layout/content, perfectly
  legible in a git diff, the feedback format agents thrive on.
- `SnapCells` (since Phase 1): the rendered view parsed into a UV cell grid
  and dumped as role-labeled style runs, e.g.
  `" gotui " [fg=TextInverted bg=Accent bold]` — the artifact for asserting
  *which role* styles what. `WithRoles(theme)` maps colors back to role names.
- `SnapStyled`: raw-ANSI golden for byte-exact styling regressions.
- `-update` flag regenerates; failures print readable line diffs.
- Doubles as the library's own regression suite.

Rejected as the core loop: `x/exp/teatest` — experimental dependency,
whole-program granularity, ANSI-laden goldens agents can't usefully read.
(Fine for occasional program-level smoke tests.) Live pty capture (vhs /
`tmux capture-pane`) is deferred to Phase 4 as a ~50-line script once there is
a real app to capture — a tool, not a platform. Snapshot limits are accepted
and documented: it verifies `View`, not `Update` behavior; the harness can
step a component through a message sequence before snapshotting to close part
of that gap.

### D10 — Agent docs: AGENTS.md + CI-compiled examples

For a library whose primary consumer is an agent, the conventions doc **is the
product's user interface**. AGENTS.md stays thin (~200 lines): explicit rules
("colors come from theme roles, never literals", "size from layout rects",
"reassign model, collect cmds") plus pointers into the runnable example apps
(`examples/statusbar`, `examples/demo`, `examples/chat`). Drift is
neutralized by making examples real packages that compile and snapshot-test
in CI — API changes that stale the docs break the build.

### D11 — Driving app: Pi TUI

A custom Go TUI for the Pi coding agent using its JSON-RPC interface: first
full parity with the stock Pi TUI, then specialized components. Real needs
drive the API; gaps found while building it flow back as gotui issues. The
long-term target is a family of agentic tools sharing one visual language.

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
├── statusbar/  list/  viewport/  textinput/  table/  help/  spinner/
│                    generic primitives, one package each
├── dialog/          modal confirm box (ResultMsg pattern)
├── overlay/         cell-space compositing — Place/Center (UV internal)
├── focus/           copy-safe (index-only) tab-order manager
├── agentic/
│   ├── markdown/    Renderer interface + glamour v2 implementation
│   ├── chat/        cell-based streaming transcript (User/Assistant/Text cells)
│   ├── toolcall/    status-aware collapsible tool-call block (chat cell)
│   ├── diffview/    styled unified diffs (Sprint inline + scrollable Model)
│   ├── permission/  numbered permission prompt (ResultMsg pattern)
│   └── usagebar/    model/tokens/cost/context bar
├── examples/
│   ├── statusbar/   canonical single-component wiring
│   ├── demo/        multi-pane app (Phase 2 exit criterion), golden-tested
│   └── chat/        mock agentic session (Phase 3 exit criterion), golden-tested
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
3. Implements `SetSize(width, height int)` (the `layout.Sizable` interface)
   and renders exactly within that box.
4. Derives every style from theme roles at construction/update time — no
   color literals anywhere.
5. Exposes focus as `Focus()`/`Blur()` where interactive.
6. Ships with golden tests (snaptest), coverage in a runnable example app,
   and an AGENTS.md recipe entry.

The root app model composes components, splits its area with `gotui/layout`
on `tea.WindowSizeMsg`, delegates messages, and wraps the final composed
string in `tea.NewView` — standard bubbletea v2, nothing hidden.

Documented exceptions to rule 3: `spinner` is intrinsic-size (a single
glyph; `SetSize` exists for the interface and is ignored), and the modal
panels (`dialog`, `permission`) treat their box as an outer bound, rendering
at natural content height.

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
- **Text editing depth:** `textarea` (multi-line editing, kill ring, IME) was
  deferred out of Phase 2; scope it when Pi TUI parity (Phase 4) demands it.
- **Streaming markdown renderer design** (Phase 5): incremental block parser
  vs full-document reparse with damage hints — decide when glamour's limits
  are measured, not guessed.
