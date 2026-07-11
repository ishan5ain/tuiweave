// Package tuiweave is a composable component library for building custom terminal
// UIs on bubbletea v2 and lipgloss v2, designed so that humans and coding
// agents can reliably compose and evolve complete frontends with it.
//
// The root package defines the semantic theme roles every component consumes.
// Layout lives in tuiweave/layout, the snapshot test harness in tuiweave/snaptest,
// and domain-neutral components live in one package each (e.g.
// tuiweave/statusbar). Agentic and other domain components live in subpackages
// such as tuiweave/agentic and are built on the same primitives.
//
// See DESIGN.md for architecture and AGENTS.md for authoring conventions.
package tuiweave
