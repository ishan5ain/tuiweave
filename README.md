# tuiweave

`tuiweave` is a composable Go component library for custom terminal interfaces
built on Bubble Tea v2 and Lip Gloss v2. It provides themed primitives for
layout, navigation, scrolling, text editing, overlays, testing, and optional
agentic interfaces while leaving application state and orchestration to you.

This is an independent community project and is not affiliated with Charmbracelet.

> **Pre-v1:** APIs may change in minor releases. Breaking changes are documented
> in the changelog and release notes.

## Install

tuiweave requires Go 1.25.8 or newer and is coupled to Bubble Tea v2.

```sh
go get github.com/ishan5ain/tuiweave@v0.1.0
```

```go
package main

import (
	"fmt"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/statusbar"
)

func main() {
	bar := statusbar.New(tuiweave.Dark())
	bar.SetSize(40, 1)
	bar.SetLeft(statusbar.Segment{Text: "ready", Kind: statusbar.KindSuccess})
	fmt.Println(bar.View())
}
```

## Explore

- General-purpose composition: [`examples/frame`](examples/frame)
- Agentic chat interface: [`examples/chat`](examples/chat)
- All runnable examples: [`examples`](examples)
- Package documentation: [pkg.go.dev](https://pkg.go.dev/github.com/ishan5ain/tuiweave)
- Architecture: [DESIGN.md](DESIGN.md)
- Roadmap: [ROADMAP.md](ROADMAP.md)
- Contributing: [CONTRIBUTING.md](CONTRIBUTING.md)
- Support: [SUPPORT.md](SUPPORT.md)
- Security: [SECURITY.md](SECURITY.md)

The root package supplies semantic theme roles. Domain-neutral components live
in top-level packages such as `layout`, `frame`, `tabs`, `menu`, `viewport`,
`textinput`, and `textarea`; optional domain packages live under `agentic/`.
The [agent catalog](AGENT-CATALOG.md) maps common tasks to packages and examples.

## Development

```sh
gofmt -w .
go mod tidy -diff
go mod verify
go build ./...
go vet ./...
go test ./...
go test -race ./...
govulncheck ./...
```

Rendering changes require intentional snapshot review; see [AGENTS.md](AGENTS.md).

## Known limitations

- The API is pre-v1 and may evolve between minor versions.
- Input-method editor behavior is incomplete.
- Components currently target Bubble Tea v2.
- Layout and rendering internals rely on a pre-v1 Ultraviolet dependency.

Licensed under the [MIT License](LICENSE).
