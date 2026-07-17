# Changelog

All notable changes are documented here. The project follows Semantic Versioning;
before v1, a minor release may contain breaking changes when clearly called out.

## [Unreleased]

- Polished `examples/ops` as the canonical non-agentic reference application,
  with app-owned mouse routing, role-aware goldens, and a README walkthrough.
- Added named focus, navigation, filtering, scrolling, editing, and narrow-state
  scenarios for list, table, viewport, and text input components.
- Completed the cross-component cell-width audit and separated the remaining
  IME contract work into its own evidence-driven follow-up.
- Added direct regression coverage for shared grapheme protection and named
  modal scenarios with explicit command-result delivery.
- Classified documented one-line components as width-bounded so
  `layout.SizeModeOf` reflects their natural one-row height.
- Stopped `SnapCells` goldens from emitting trailing whitespace for empty rows.
- Added a compile-checked `snaptest` tutorial covering readable, role-aware,
  and scenario goldens with explicit command delivery.
- Added an end-to-end semantic inspection and action-routing workflow in
  `examples/ops`, including explicit command-result delivery.
- Defined the v1 adoption boundary with a readiness checklist and a
  consumer-shaped public API example.
- Added package-level README guides and a documentation coverage guardrail for
  public packages.
- Added a discoverable catalog of eight built-in theme presets with stable IDs,
  lookup, named constructors, and a cycling statusbar example.

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

[Unreleased]: https://github.com/ishan5ain/tuiweave/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/ishan5ain/tuiweave/releases/tag/v0.1.0
