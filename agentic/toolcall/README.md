# agentic/toolcall

`toolcall` renders a mutable, status-aware tool execution block that implements `chat.Cell`.

```go
block := toolcall.New(theme, "shell", "Run tests")
block.SetID("tool-1")
block.SetStatus(toolcall.StatusRunning)
block.AppendOutput("go test ./...\n")
transcript.Append(block)
```

Keep IDs stable across retries and call the containing transcript's `Invalidate` after mutating status or output. `Retry` advances the attempt and resets execution state; `Cancel` records cancellation. The application owns actual tool execution and policy.

See the [chat example](../../examples/chat) and [agentic package map](../../AGENT-CATALOG.md#agentic-domain-components).

## API highlights

- `New` creates a named tool block with a summary.
- `SetID`, `SetStatus`, `Retry`, and `Cancel` manage lifecycle state.
- `AppendOutput` streams output into the block.
- `Render` implements the width-aware chat-cell contract.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/agentic/toolcall)
- [Chat example](../../examples/chat)
- [Agentic package map](../../AGENT-CATALOG.md#agentic-domain-components)
