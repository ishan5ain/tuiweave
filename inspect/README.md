# inspect

`inspect` describes a composed UI as data for tests, automation, debugging, and
assistive tooling, and carries semantic action intents. It is a semantic state
contract, not an accessibility renderer, transport, event loop, or application
framework.

```go
root := inspect.Group("app", "application", inspect.Bounds{Width: w, Height: h},
	inspect.Bind("list", m.list),
	inspect.BindAt("input", inspect.FromRect(inputArea), m.input),
)
data, err := inspect.Marshal(root)
```

The application owns stable IDs, visibility, absolute bounds, privacy, and
routing. Bound component actions are qualified by node ID; route the target in
the app, then deliver the local suffix with `inspect.Invoke`.

See [agent-operable surfaces](../AGENT-CATALOG.md#agent-operable-surfaces).

## API highlights

- `Group`, `Bind`, and `BindAt` build an application-owned semantic tree.
- `Marshal` encodes inspection data.
- `Invoke` creates a local semantic action message.
- Components supply `Inspect()` and `Actions()`; the app supplies IDs and bounds.

## Node contract

Component reports are local and data-only:

| Field | Stable meaning |
|---|---|
| `ID` | Application identity after `Bind`/`BindAt`. A component-owned correlation ID may be overwritten when bound. |
| `Kind` | Stable component kind such as `list`, `dialog`, or `diffview`. |
| `Bounds` | Component-local size from `SetSize`; `BindAt` replaces it with application-owned absolute terminal-cell bounds. |
| `Label` | Human-readable summary, not an identity or source-of-truth value. |
| `Focused` | Whether the component currently owns local keyboard interaction. Always-active modal choices report focused when included in the visible tree. |
| `Status` | Component-specific lifecycle or semantic state. Consumers must interpret it with `Kind`. |
| `Selected` | Current logical selection when the component has one. `Index` follows the component's public selection index contract; filtered components may retain an original-item index while `Count` describes the current matched population. |
| `Scroll` | Total logical lines/items, visible capacity, and zero-based offset. It describes state, not screen origin. |
| `Actions` | Local semantic intents. `Bind` qualifies non-empty IDs as `<node-id>.<local-id>`. |
| `Attributes` | Small string-valued facts documented by component kind. They are not a duplicate domain model. |
| `Children` | Ordered application- or component-owned semantic descendants. Visibility and privacy filtering still belong to the application. |

Absent optional fields mean “not applicable or intentionally undisclosed,” not
an empty value. In particular, text inputs and textareas report shape, cursor,
selection, and history facts without copying user-entered text. Viewport,
diffview, and chat reports omit rendered content. Applications should omit an
entire node when the current inspection consumer must not know that it exists.

## Action contract

Action IDs are stable local protocol names. Common IDs have consistent meaning:

| ID | Meaning |
|---|---|
| `focus`, `blur` | Change local keyboard focus where supported. |
| `next`, `previous`, `first`, `last` | Move a logical selection. |
| `select.<stable-id>` | Select an application-defined item or tab. Disabled items remain discoverable with `enabled: false`. |
| `activate` | Emit the same typed result command as keyboard activation. The app must deliver the resulting message through `Update`. |
| `clear` | Clear the component-owned query or editable value. |
| `scroll_up`, `scroll_down`, `scroll_top`, `scroll_bottom` | Move a scroll window independently of keyboard focus. |

Editing controls and modal components add documented domain-local actions such
as `undo`, `redo`, `yank`, `select_all`, `confirm`, `cancel`, and `choose.N`.
`Enabled` is a snapshot of whether the action is currently permitted; an
enabled action may be idempotent. Applications must resolve the action against
the current visible tree and reject unknown, hidden, or disabled actions before
forwarding the local suffix. Components safely preserve state for unavailable
actions, but that is not a substitute for application-boundary validation.

`snaptest.RunScenario` records whether an update emitted a command but does not
execute it. Tests that cover activation must call the command, assert its typed
message, and explicitly deliver that message back to the application model.

## Interactive component inventory

| Kind | Semantic state | Additional action families |
|---|---|---|
| `autocomplete` | query, original/matched counts, selected ID, selection, scroll | selection, clear, activate |
| `button` | disabled, correlation ID, ready/disabled status | activate |
| `chat` | focus, following, cell count, scroll, identified child cells | scroll |
| `dialog` | title/body, active answer, awaiting-input status | next/previous, confirm/cancel |
| `diffview` | focus and scroll; diff text is omitted | scroll |
| `list` | item/matched counts, selection, filter, scroll | selection |
| `menu` | item/enabled counts, selected ID, selection, scroll | selection, activate |
| `palette` | query, item/matched counts, selected ID, selection, scroll | selection, clear, activate |
| `permission` | title/body, provenance, active choice, awaiting-approval status | `choose.N` |
| `table` | row/column counts, selected-row summary, scroll | selection |
| `tabs` | tab count, selected ID, visible start, selection | selection |
| `textarea` | content shape, logical cursor, selection size, undo/redo capability | clear, selection, history, yank |
| `textinput` | empty state, placeholder, logical cursor; value is omitted | clear |
| `toolcall` | identity, lifecycle, summary, attempt, output-line count; output is omitted | cancel/retry through `ApplyAction` |
| `toggle` | checked, disabled, correlation ID, on/off status | toggle/on/off |
| `toolbar` | item/enabled counts, selected ID, selection, horizontal window | selection, activate |
| `viewport` | focus and scroll; content is omitted | scroll |

Passive inspected components such as `progress` expose semantic state but no
actions. Other passive rendering helpers need no inspection surface.

## Application ownership and routing

The application must:

1. Include only currently visible and disclosure-safe nodes.
2. Assign stable IDs and absolute bounds from the same layout rectangles used
   to call `SetSize`.
3. Look up the qualified action in the current tree and require `Enabled`.
4. Strip exactly the target node's prefix and forward `inspect.Invoke(localID)`
   to that component, or use its documented local method when it is not an MVU
   model (for example, `toolcall.Block.ApplyAction`).
5. Reassign the returned model, retain the returned `tea.Cmd`, and let Bubble
   Tea deliver its typed result normally.
6. Keep persistence, side effects, modal visibility, focus stacks, privacy,
   transport, and authorization outside reusable components.

See [`examples/consumer`](../examples/consumer) for a minimal public-API tree
and qualified action router, and [`examples/ops`](../examples/ops) for nested
modal visibility, disabled/hidden rejection, semantic activation commands, and
explicit result delivery.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/inspect)
- [Agent-operable surfaces](../AGENT-CATALOG.md#agent-operable-surfaces)
