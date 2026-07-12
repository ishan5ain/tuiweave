# agentic/chat

`chat` provides a scrolling, auto-following transcript built from width-aware cells.

```go
transcript := chat.New(theme)
assistant := chat.NewAssistant(theme, renderer)
assistant.SetID("answer-1")
transcript.Append(assistant)
transcript.SetSize(width, height)

assistant.Append(delta)
transcript.Invalidate()
```

Cells are pointers retained and mutated by the application. Call `Invalidate` after in-place changes; `Append`, `Replace`, and `SetSize` invalidate automatically. Mouse-wheel scrolling works while blurred, and scrolling away from the bottom disables auto-follow until the bottom is reached again.

See the [chat example](../../examples/chat) and [chat recipe](../../AGENTS.md#viewport--list--table-scrolling-components).

## API highlights

- `New` creates an auto-following transcript.
- `NewUser`, `NewAssistant`, and `NewText` create built-in cells.
- `Append`, `Replace`, `Cells`, and `Invalidate` manage transcript content.
- `Following`, `GotoBottom`, and viewport accessors expose scroll state.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/agentic/chat)
- [Chat example](../../examples/chat)
- [Scrolling conventions](../../AGENTS.md#viewport--list--table-scrolling-components)
