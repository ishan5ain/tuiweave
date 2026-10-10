# Changelog

All notable changes are documented here. The project follows Semantic Versioning;
before v1, a minor release may contain breaking changes when clearly called out.

## [Unreleased]

### Performance

- Pinned Lip Gloss's upstream writer batching and wrap-buffer sizing commits
  to reduce repeated Markdown rendering overhead. Downstream consumers resolve
  the unreleased `v2.0.7-0.20260901023042-868e8b5c4fb6` commit through Go's
  minimal version selection. Revisit the pin when a v2.0.7-or-later tag contains
  both fixes and preserves the Go 1.26.0 floor, after compatibility validation.
  Public APIs, Go minimum, and existing rendering goldens are unchanged; the
  [writer experiment](agentic/markdown/WRITER-BATCHING.md) records measurements
  and remaining targets in issue #49.

### Added

- Writer-boundary compatibility coverage for split ANSI/UTF-8 input, newline
  style/link restoration, CRLF, nested writers, and writes after close, plus
  direct `Wrap` output comparisons with v2.0.6.

## [0.3.0] - 2026-10-10

### Compatibility

- Raised the minimum Go version from 1.25.8 to 1.26.0 and upgraded Bubble Tea
  to v2.1.0 for shutdown, terminal-size fallback, and input-disabled query fixes.
  Consumers must upgrade their Go toolchain before adopting v0.3.0; no
  tuiweave API changes are required. Earlier release tags keep their original
  Go requirements.
- Updated the required Ultraviolet, ANSI utilities, rune-width, and terminfo
  pins selected by Bubble Tea v2.1.0, and aligned CI and the compatibility matrix
  with the new graph.

### Changed

- Documented and regression-tested derived themes that use
  `lipgloss.NoColor{}` for inherited raised surfaces while preserving
  selection, accent, and intent fills.

### Fixed

- Preserved regional-indicator flag graphemes when table selection repaints
  under wcwidth terminals by updating `x/ansi` and adding renderer-transition
  regression coverage.

## [0.2.0] - 2026-07-17

### Compatibility

- Replaced the public Ultraviolet `layout.Constraint` alias with a
  tuiweave-owned sealed representation. Applications using `layout.Len`,
  `Min`, `Max`, `Percent`, `Ratio`, and `Fill` require no migration. Pre-v1
  applications that directly constructed, converted, or stored Ultraviolet
  constraint values must replace them with tuiweave's constructors.
- Defined the tested compatibility and migration policy, retained the agentic
  packages in the main module at the Go 1.25.8 floor, and made `layout.Rect`
  depend directly on the standard-library rectangle identity.

### Added

- Added a discoverable catalog of eight built-in theme presets with stable IDs,
  lookup, named constructors, and a cycling statusbar example.
- Added cell-aware `tabs.IndexAt` local-coordinate hit-testing and used it for
  mouse-selectable tabs in the canonical operations example.
- Defined the cross-component inspection and accessibility contract, added a
  consumer-shaped qualified-action example, completed diffview inspection,
  and aligned modal focus and toolbar selection-ID reporting.
- Added a compile-checked `snaptest` tutorial covering readable, role-aware,
  and scenario goldens with explicit command delivery.
- Added package-level README guides, a documentation coverage guardrail, a v1
  readiness checklist, and a consumer-shaped public API example.

### Changed

- Classified documented one-line components as width-bounded so
  `layout.SizeModeOf` reflects their natural one-row height.
- Polished `examples/ops` as the canonical non-agentic reference application,
  with app-owned mouse routing, semantic action routing, role-aware goldens,
  and a README walkthrough.
- Reconciled the v1 roadmap around scoped IME and streaming-Markdown
  investigations, recording completed foundations and deferring speculative
  APIs until applications provide evidence.
- Updated the indirect `golang.org/x/net` dependency and pinned newer
  `actions/checkout` and `actions/setup-go` CI revisions.

### Fixed and hardened

- Completed the cross-component cell-width audit and added direct regression
  coverage for shared grapheme protection across narrow rendering paths.
- Added named modal scenarios with explicit command-result delivery and named
  focus, navigation, filtering, scrolling, editing, and narrow-state scenarios
  for list, table, viewport, and text input components.
- Completed representative paired plain-text and role-aware golden coverage,
  with documented exclusions for styling pass-through utilities.
- Stopped `SnapCells` goldens from emitting trailing whitespace for empty rows.
- Separated the remaining IME contract work into its own evidence-driven
  follow-up.

## [0.1.0] - 2026-07-11

Initial public release, including semantic themes; constraint layout; frames,
stacks, split panes, tabs, menus, toolbars, palettes, buttons, toggles, progress,
status bars, help, and spinners; scrollable viewports, lists, and tables; text
input, textarea editing/history/selection/completion APIs, and autocomplete;
mouse, focus, overlays, dialogs, inspection/actions, and snapshot/scenario test
helpers; optional markdown, chat, tool-call, diff, permission, and usage packages;
and runnable general-purpose and agentic examples.

Known limitations: pre-v1 API evolution, incomplete IME behavior, Bubble Tea v2
coupling, and internal reliance on a pre-v1 Ultraviolet dependency.

[Unreleased]: https://github.com/ishan5ain/tuiweave/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/ishan5ain/tuiweave/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/ishan5ain/tuiweave/releases/tag/v0.1.0
