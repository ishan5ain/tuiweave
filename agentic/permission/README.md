# agentic/permission

`permission` renders a width-bounded approval prompt with structured operation provenance.

```go
p := permission.New(theme)
p.ID = "run-tests"
p.Title = "Run the test suite?"
p.SetProvenance(permission.Provenance{
	Tool: "shell", Operation: "execute", Target: "go test ./...",
})
p.SetSize(width, height)
```

Number keys, navigation plus Enter, and Escape emit `permission.ResultMsg`; Escape chooses the last, safe-default option. The application owns visibility, overlay composition, modal focus, policy, and execution of an approved operation.

See the [chat example](../../examples/chat) and [modal compatibility notes](../../AGENT-CATALOG.md#api-compatibility-notes).

## API highlights

- `Title`, `Body`, and `ID` identify the approval request.
- `SetOptions` replaces choices; `SetProvenance` adds structured context.
- `ResultMsg` reports the selected choice and label.
- `SetSize`, `Update`, and `View` implement the width-bounded modal contract.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/agentic/permission)
- [Chat example](../../examples/chat)
- [Modal compatibility notes](../../AGENT-CATALOG.md#api-compatibility-notes)
