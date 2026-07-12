# agentic/markdown

`markdown` renders assistant Markdown through a small backend-neutral `Renderer` interface.

```go
r := markdown.NewRenderer(theme)
view := markdown.Sprint(r, source, width)
```

Depend on `Renderer` rather than the concrete Glamour implementation. `Sprint` degrades to raw source if rendering fails, so malformed or partial streaming input remains visible. Reuse the renderer; it caches work by wrap width.

See the [chat example](../../examples/chat) and [architecture rationale](../../DESIGN.md#d7--markdown-glamour-now-custom-later).

## API highlights

- `Renderer` is the backend-neutral rendering interface.
- `NewRenderer` creates the theme-aware implementation.
- `Sprint` renders a source string and falls back to raw source on errors.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/agentic/markdown)
- [Chat example](../../examples/chat)
- [Markdown architecture](../../DESIGN.md#d7--markdown-glamour-now-custom-later)
