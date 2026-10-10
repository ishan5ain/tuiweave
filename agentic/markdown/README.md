# agentic/markdown

`markdown` renders assistant Markdown through a small backend-neutral `Renderer` interface.

```go
r := markdown.NewRenderer(theme)
view := markdown.Sprint(r, source, width)
```

Depend on `Renderer` rather than the concrete Glamour implementation. `Sprint`
degrades to raw source if rendering fails, so malformed or partial streaming
input remains visible. Reuse the renderer; it caches renderer construction by
wrap width, not rendered documents. Every changed source still triggers a full
render.

See the [chat example](../../examples/chat) and [architecture rationale](../../DESIGN.md#d7--markdown-glamour-now-custom-later).

## API highlights

- `Renderer` is the backend-neutral rendering interface.
- `NewRenderer` creates the theme-aware implementation.
- `Sprint` renders a source string and falls back to raw source on errors.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/agentic/markdown)
- [Chat example](../../examples/chat)
- [Markdown architecture](../../DESIGN.md#d7--markdown-glamour-now-custom-later)
- [Compatibility and module policy](../../COMPATIBILITY.md)

## Streaming performance

The [streaming investigation](STREAMING-PERFORMANCE.md) documents deterministic
append-and-render benchmarks, raw measurements, profiles, reproduction commands,
and the decision to scope ANSI writer allocation improvements separately.
The [writer batching experiment](WRITER-BATCHING.md) records the first upstream
dependency improvement and the remaining targets.
One streaming operation replays an entire response, and the full matrix is
intentionally expensive. See the report for repeated measurement commands;
for a single pass:

```sh
go test ./agentic/markdown -run '^$' -bench 'Streaming|Completed' \
  -benchmem -benchtime=1x
```
