# Compatibility policy

This document is the canonical compatibility contract for tuiweave. It records
the versions the project tests, the public boundaries it intends to stabilize
for v1, and how applications are expected to migrate when those boundaries
change. [DEPENDENCY-TRACKING.md](DEPENDENCY-TRACKING.md) is a dated upstream
watch report, not a support promise.

## Supported versions

| Surface | Supported version | Evidence |
|---|---|---|
| Go | 1.26.7 minimum | `go.mod`; Linux CI at the exact floor |
| Go | current stable | Linux, macOS, and Windows CI |
| Bubble Tea | `charm.land/bubbletea/v2 v2.1.1` | `go.mod`; full CI suite and consumer runtime test |
| Lip Gloss | `charm.land/lipgloss/v2 v2.1.0` | `go.mod`; full CI suite and writer-boundary tests |
| Glamour | `charm.land/glamour/v2 v2.0.1` | `go.mod`; agentic Markdown tests |
| Ultraviolet | `v0.0.0-20261008173134-6b8d4baf91b4` | `go.mod`; layout, overlay, and SnapCells coverage |
| ANSI utilities | `github.com/charmbracelet/x/ansi v0.11.9` | `go.mod`; rendering and cell-width coverage |

The supported dependency graph is the graph selected by `go.mod`. tuiweave
does not promise compatibility with arbitrary newer upstream versions merely
because Go's module solver can select them. Dependency upgrades are deliberate
compatibility events and include changelog review, the full test suite, and
golden review when rendering changes.

The matrix is reproduced in CI by testing the exact minimum Go version on
Linux and current stable Go on Linux, macOS, and Windows. Locally, validate the
selected graph with:

```sh
go mod tidy -diff
go mod verify
go build ./...
go vet ./...
go test ./...
go test -race ./...
```

## Go version and module boundary

The current supported graph requires Go 1.26.7 or newer, established by
Lip Gloss v2.1.0. The v1 policy keeps tuiweave in one Go module at the declared
compatibility floor, subject to deliberate dependency review.
Splitting `agentic` into another module would add coordinated
release tags, dependency wiring, and example complexity without demonstrated
application demand.

The decision can be revisited in a future major version if applications need a
lower Go floor or the agentic packages require an independent release cadence.
Until then, all top-level and `agentic/...` packages share one module version.

### Go 1.26 decision (2026-10-10)

tuiweave v0.3.0 raises the minimum from Go 1.25.8 to Go 1.26.0 to
adopt [Bubble Tea v2.1.0](https://github.com/charmbracelet/bubbletea/releases/tag/v2.1.0).
The upgrade brings fixes for shutdown signal deadlocks, zero-size terminal
startup and resizing, and terminal queries issued with input disabled. Its
required Ultraviolet update also fixes renderer repainting and input handling.
These runtime fixes justify revising the earlier Go floor; the review found
no documented consumer requirement to retain Go 1.25.8. This does not imply
that every external consumer has been audited.

Consumers must install Go 1.26.0 or newer before adopting v0.3.0.
With `GOTOOLCHAIN=local`, Go 1.25 rejects the upgraded module. Existing tags
retain their original requirements. No tuiweave API migration is required,
and terminal Program Status Protocol reporting remains application-owned and
opt-in through Bubble Tea's `tea.View`.

### Go 1.26.7 decision (2026-11-15)

tuiweave v0.4.0 adopts tagged Lip Gloss v2.1.0, retaining both writer
batching fixes and deliberately raising the Go minimum from 1.26.0 to 1.26.7.
This supersedes the experiment’s original requirement to preserve Go 1.26.0.
Install Go 1.26.7 or newer before upgrading; no tuiweave API migration is
required. CI reads the exact minimum from `go.mod` and sets `GOTOOLCHAIN=local`
so an automatic toolchain upgrade cannot mask an unsupported minimum.

Bubble Tea v2.1.1 also fixes untrusted terminal-size reports that could exhaust
process memory; see [the upstream advisory](https://github.com/charmbracelet/bubbletea/security/advisories/GHSA-qc58-pfg9-7pph).
Applications accepting remote terminal input should prioritize this upgrade.

## Lip Gloss writer update (2026-10-10; historical)

The originally selected Lip Gloss commit included upstream batching between
newlines and pre-sizing of wrapped output buffers. It was pinned by commit
because those changes were not yet in a tagged release. The
[writer experiment](agentic/markdown/WRITER-BATCHING.md) records compatibility
checks, benchmark evidence, and the remaining performance targets in issue #49.
Compared with v2.0.6, the upstream production change is limited to `wrap.go`;
module requirements were unchanged. That adoption required no tuiweave API
migration or Go minimum change, and existing rendering goldens were unchanged.
Later release adoption requires the same dependency review and validation.

The current graph uses tagged v2.1.0 instead of that pseudo-version. The
experiment’s measurements remain historical evidence; they are not new
benchmarks of the tagged release.

## Paste adoption in v0.4.0

Applications that previously forwarded only key messages must also delegate
`tea.PasteMsg` to their focused input or visible palette, reassign the returned
model, and collect its command. Refresh application-owned completion queries,
filters, and layout after the value changes. Keep hidden controls blurred and
route input to the top modal. Paste does not submit or activate a command.

Single-line inputs replace pasted LF/tab with spaces; textarea preserves
logical LF/tab and inserts nonempty paste as one undoable edit. Textarea tabs
render as four cells, or fill a narrower row, while cursor positions remain
rune-based. Programmatic setters retain their existing behavior. See the
[textinput](textinput/README.md#terminal-paste),
[textarea](textarea/README.md#terminal-paste), and
[palette](palette/README.md#terminal-paste) contracts and runnable examples.

## Layout compatibility

### Rectangles

`layout.Rect` is part of the v1 contract as an alias of `image.Rectangle`.
Applications may use the standard `Min`, `Max`, `Dx`, `Dy`, `Inset`, and
`Intersect` geometry surface. The alias is defined directly against the
standard library; Ultraviolet rectangle identity is not part of the contract.

### Constraints

The stable application API is the `layout.Constraint` name plus the `Len`,
`Min`, `Max`, `Percent`, `Ratio`, and `Fill` constructors. tuiweave owns the
sealed constraint representation and translates it to the Ultraviolet solver
only when constructing a `Vertical` or `Horizontal` layout. Ultraviolet
constraint identity is not part of the v1 contract.

Applications following the documented API require no migration:

```go
parts := []layout.Constraint{
	layout.Len(3),
	layout.Fill(1),
}
view := layout.Vertical(parts...)
```

Direct construction of, conversion to, or storage as
`github.com/charmbracelet/ultraviolet/layout.Constraint` is unsupported and no
longer type-compatible. An application that used that pre-v1 implementation
detail must replace Ultraviolet constraint values with tuiweave's constructors.
This boundary was completed in
[Issue #28](https://github.com/ishan5ain/tuiweave/issues/28).

## Release and migration policy

### Before v1

- Minor releases may contain breaking changes when the changelog and release
  notes identify the affected API and provide a migration example.
- Deprecation is preferred before removal when practical, but pre-v1 releases
  do not promise a fixed deprecation window.
- Dependency bumps that alter exported behavior, rendering, input, geometry,
  or the minimum Go version are called out as compatibility changes.
- Applications should pin a tagged tuiweave version or commit rather than use
  `@latest` in reproducible instructions.

### Starting with v1

- tuiweave follows Semantic Versioning for exported APIs and documented
  behavior.
- Backward-compatible features are minor releases; backward-compatible fixes
  are patch releases.
- Exported names are marked with Go's `Deprecated:` documentation before a
  replacement is preferred. Removal or a source-incompatible signature change
  waits for the next major version.
- Release notes identify deprecated APIs and show the supported replacement.
- Security and correctness fixes may change erroneous behavior in a patch
  release, but source compatibility remains the default within a major line.
- Support is limited to the versions declared in this document and exercised
  by CI; no separate long-term-support window is promised.

## Changing this policy

Changes to the minimum Go version, module boundary, supported upstream majors,
or public layout identity require an explicit compatibility decision, release
notes, and corresponding CI evidence. A dependency watch item alone does not
change this contract.
