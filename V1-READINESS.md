# v1 readiness checklist

This checklist tracks the boundary between the current pre-v1 vocabulary and a
stable v1 release. It is a decision aid, not a release schedule.

## Compatibility

- [x] Publish the supported Go, Bubble Tea, Lip Gloss, Glamour, and Ultraviolet
      versions as a tested compatibility matrix.
- [x] Keep Go 1.25.8 as the minimum and the optional `agentic` layer in the
      main module; see [COMPATIBILITY.md](COMPATIBILITY.md#go-version-and-module-boundary).
- [x] Keep `layout.Rect` as a standard-library alias and own the
      `layout.Constraint` representation while preserving its constructor API;
      see [COMPATIBILITY.md](COMPATIBILITY.md#layout-compatibility) and
      [Issue #28](https://github.com/ishan5ain/tuiweave/issues/28).
- [x] Document the deprecation, migration, and removal policy for v1; see
      [COMPATIBILITY.md](COMPATIBILITY.md#release-and-migration-policy).

## Component contracts

- [x] Verify every bounded component's `SetSize` and exact-width/height behavior
      across narrow, empty, focused, and blurred states.
- [x] Complete the cell-width expectations for the audited text-bearing
      components; see [Issue #10](https://github.com/ishan5ain/tuiweave/issues/10).
- [ ] Define IME expectations separately with application evidence; see
      [Issue #19](https://github.com/ishan5ain/tuiweave/issues/19).
- [x] Define the accessibility semantics exposed by inspection and document
      what remains application-owned.
- [x] Keep application state, visibility, routing, and side effects outside
      reusable components.

## Verification and operability

- [x] Maintain paired plain-text and role-aware goldens for representative
      theme-owned rendering behavior, with documented exclusions for styling
      pass-through utilities; see [Issue #31](https://github.com/ishan5ain/tuiweave/issues/31).
- [x] Cover important transitions with named scenarios, including emitted
      commands and explicitly delivered command results.
- [x] Demonstrate semantic inspection and action routing end to end in a small,
      deterministic application test.
- [x] Exercise the public API from a consumer-shaped example that does not rely
      on internal packages or undocumented helpers.

## Adoption

- [x] Present one polished non-agentic reference application in the README.
- [x] Give newcomers a short package-learning path and a clear comparison with
      Bubbles and framework-style alternatives.
- [ ] Convert the remaining roadmap areas into scoped issues with acceptance
      tests, non-goals, and an explanation of which applications need them.
