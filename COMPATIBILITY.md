# Compatibility policy

This document is the canonical compatibility contract for tuiweave. It records
the versions the project tests, the public boundaries it intends to stabilize
for v1, and how applications are expected to migrate when those boundaries
change. [DEPENDENCY-TRACKING.md](DEPENDENCY-TRACKING.md) is a dated upstream
watch report, not a support promise.

## Supported versions

| Surface | Supported version | Evidence |
|---|---|---|
| Go | 1.26.0 minimum | `go.mod`; Linux CI at the exact floor |
| Go | current stable | Linux, macOS, and Windows CI |
| Bubble Tea | `charm.land/bubbletea/v2 v2.1.0` | `go.mod`; full CI suite and consumer runtime test |
| Lip Gloss | `charm.land/lipgloss/v2 v2.0.6` | `go.mod`; full CI suite |
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

tuiweave v1 will remain one Go module and will require Go 1.26.0 or newer.
Bubble Tea v2.1.0 and the selected Ultraviolet revision establish that floor.
Splitting `agentic` into another module would add coordinated
release tags, dependency wiring, and example complexity without demonstrated
application demand.

The decision can be revisited in a future major version if applications need a
lower Go floor or the agentic packages require an independent release cadence.
Until then, all top-level and `agentic/...` packages share one module version.

### Go 1.26 decision (2026-10-10)

The next pre-v1 release raises the minimum from Go 1.25.8 to Go 1.26.0 to
adopt [Bubble Tea v2.1.0](https://github.com/charmbracelet/bubbletea/releases/tag/v2.1.0).
The upgrade brings fixes for shutdown signal deadlocks, zero-size terminal
startup and resizing, and terminal queries issued with input disabled. Its
required Ultraviolet update also fixes renderer repainting and input handling.
These runtime fixes justify revising the earlier Go floor; the review found
no documented consumer requirement to retain Go 1.25.8. This does not imply
that every external consumer has been audited.

Consumers must install Go 1.26.0 or newer before adopting the next release.
With `GOTOOLCHAIN=local`, Go 1.25 rejects the upgraded module. Existing tags
retain their original requirements. No tuiweave API migration is required,
and terminal Program Status Protocol reporting remains application-owned and
opt-in through Bubble Tea's `tea.View`.

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
