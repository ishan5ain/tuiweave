# gotui

`gotui` is a composable Go toolkit for building custom terminal user
interfaces. It provides reliable primitives for layout, styling, interaction,
scrolling, text editing, overlays, and rendering, with optional domain packages
for agentic and other specialized applications.

It is built on [bubbletea v2](https://github.com/charmbracelet/bubbletea),
[lipgloss v2](https://github.com/charmbracelet/lipgloss), and
[ultraviolet](https://github.com/charmbracelet/ultraviolet). It is designed
**agent-friendly by design**: the API, conventions, examples, and verification
loop make the component system easy for both humans and coding agents to
understand, compose, and evolve.

The north star is a **small, composable Go vocabulary for terminal interfaces**:
make layout, appearance, interaction, and verification predictable for humans
and coding agents. Applications own orchestration and domain state; gotui owns
reusable primitives and optional domain kits. Agentic UIs are a demanding
proving ground, not the boundary of the core library.

**Status: pre-v1.** Phases 0–3.5 are complete: foundation, core interaction
primitives, and the first agentic domain layer. APIs still change freely while
agent-operable foundations and the general-purpose composition layer evolve.

## What's here

The library is organized in layers. The core stays domain-neutral; domain
packages are built on top of it and applications own their event routing,
backend protocols, and session lifecycle.

| Package | Purpose |
|---|---|
| `gotui` (root) | Semantic theme roles (`Theme`), `Dark()`/`Light()` defaults |
| `gotui/layout` | Flexbox-like layout: constraints → rectangles → component sizes |
| `gotui/snaptest` | Snapshot test harness: plain-text, raw-ANSI, role-labeled cell-grid, and interaction-scenario goldens |
| `gotui/inspect` | Optional semantic UI tree and action metadata: IDs, bounds, focus, selection, scrolling, actions, and children |
| `gotui/statusbar` | One-line status bar with themed left/right segments |
| `gotui/list` | Scrolling list with selection cursor and filtering (original-index selection) |
| `gotui/viewport` | Scrollable window over pre-rendered content (keys + mouse wheel) |
| `gotui/textinput` | Single-line input: cursor, placeholder, horizontal scroll |
| `gotui/textarea` | Multi-line input: soft wrap, visual-row cursor, content-driven height |
| `gotui/scrollbar` | One-column scroll indicator for any `Scrollable` component |
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
| `examples/…` | Runnable apps: `go run ./examples/chat` (mock agentic session), `./examples/table` (git-status mock), `./examples/demo`, `./examples/statusbar` |

Agentic packages are important reference implementations, not the boundary of
the library. The same primitives should support editors, dashboards, file
browsers, operational tools, forms, and other custom TUIs. See [PLAN.md](PLAN.md)
for the next phases of general-purpose evolution.

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
- **Composable layers** — primitives and composition utilities remain
  domain-neutral; agentic and application-specific packages build on them.
- **Agent-friendly composition** — the package structure, contracts, recipes,
  and examples provide a small vocabulary that coding agents can use without
  inventing inconsistent local patterns.

Agent-friendliness is an architectural quality, not a product specialization.
The library should be easy to discover, difficult to misuse, flexible enough for
distinct visual designs, and explicit about the interaction conventions that
make those designs feel coherent.

In practice, a gotui UI should support a complete lifecycle:

- **Discover** the right package, recipe, and example.
- **Compose** it from explicit sizing, theme, focus, and state contracts.
- **Verify** it through readable rendering snapshots today and deterministic
  interaction scenarios as the roadmap expands.
- **Operate** it through optional semantic inspection and stable actions.
- **Recover** from loading, failure, cancellation, and narrow-terminal states.

Discovery, composition, rendering snapshots, deterministic scenarios, semantic
inspection/actions, and structured approval provenance now have initial support;
the roadmap focuses on hardening those contracts while expanding the
composition vocabulary.

## Scope boundaries

`gotui` owns reusable rendering and interaction components. It does not own an
application event loop, backend protocol, persistence layer, or product-specific
workflow. A package belongs in the core when it expresses a reusable terminal
interaction pattern; it belongs under `agentic/` or in an application when it
depends on a domain model or service.

## Documentation

- [DESIGN.md](DESIGN.md) — architecture, decision record, component contract
- [PLAN.md](PLAN.md) — phased roadmap with exit criteria
- [AGENTS.md](AGENTS.md) — conventions for coding agents building with gotui
- [AGENT-CATALOG.md](AGENT-CATALOG.md) — compact package and task routing index

## Development

```sh
go build ./... && go vet ./... && go test ./...

# regenerate snapshot goldens after an intentional visual change:
go test ./... -update
```
