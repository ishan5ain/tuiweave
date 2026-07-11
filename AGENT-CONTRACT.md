# gotui Agent Contract

This is the short workflow contract for agents building applications with
gotui. The library's API and design truth live in the gotui repository:
[`AGENTS.md`](AGENTS.md), [`AGENT-CATALOG.md`](AGENT-CATALOG.md),
[`DESIGN.md`](DESIGN.md), and the runnable examples. Do not copy or fork that
rulebook into an application.

## Before editing

- Inspect the repository, `go.mod`, current tests, and the closest gotui
  example. Record the gotui version or commit being used.
- For a greenfield app, map components, layout rectangles, focus order, and
  application-owned state before writing the shell.
- For a migration, inventory rendering, input, scrolling, focus, and domain
  behavior. Classify each concern as reusable gotui behavior, application
  behavior, or a candidate gotui API gap.
- For a review, verify theme roles, `SetSize`/layout ownership, MVU model and
  command handling, focus reapplication, mouse bounds, modal results, narrow
  rendering, and snapshot/scenario coverage.

## Task contract

Every implementation task states:

```text
Goal:
Reference gotui example:
Components:
Application-owned state:
Files allowed:
Behavior to preserve:
Acceptance tests:
Verification commands:
Out of scope:
```

Keep migration slices vertically usable and give each slice explicit file
ownership. Do not parallelize edits to shared model, render, or update files
unless isolated worktrees are used. Preserve behavior before intentionally
changing visuals, and review every golden diff.

## Compatibility and gaps

Use the actual API names and contracts from the pinned gotui revision:

- `dialog` answers with `dialog.ResultMsg`; there is no `dialog.Message`.
- `autocomplete` uses `SetItems`, `SetQuery`, and `SelectedMsg`; there is no
  `SetSuggestions`.
- `scrollbar.For(theme, component)` renders the bar; there is no
  `scrollbar.Model` or `scrollbar.ForViewport`.
- `focus.Manager` is a value type. Apply fresh component addresses after each
  copied model update; do not store component pointers in the manager.
- `textarea.CursorPosition` and `textarea.ReplaceRange` use logical rune
  coordinates. Do not synchronize a completion buffer from `Value()` alone
  when the cursor may be in the middle of the text.
- Layout produces `layout.Rect`; the application owns rectangles and calls
  bounded components' `SetSize` on every window-size layout pass.

Do not invent local forks or guessed compatibility wrappers. If a reusable
application need cannot be expressed by the pinned API, report it as a
candidate gotui API gap with a focused example and acceptance test. A gotui
change requires explicit authorization; a local `replace` directive is only
for development and must not replace the pinned dependency in committed app
configuration.

## Verification

Run the app's build, vet, tests, snapshot update when intentionally changing
rendering, and the read-only audit from
[`gotui-agent-kit`](https://github.com/ishansain/gotui-agent-kit). Read the
golden diff before handoff. Include failures, deferred gaps, and the exact
gotui revision in the result.
