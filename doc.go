// Package gotui is a composable component library for building custom terminal
// UIs on bubbletea v2 and lipgloss v2, designed so that coding agents can
// reliably author and evolve complete frontends with it.
//
// The root package defines the semantic theme roles every component consumes.
// Layout lives in gotui/layout, the snapshot test harness in gotui/snaptest,
// and domain-neutral components live in one package each (e.g.
// gotui/statusbar). Agentic and other domain components live in subpackages
// such as gotui/agentic and are built on the same primitives.
//
// See DESIGN.md for architecture and AGENTS.md for authoring conventions.
package gotui
