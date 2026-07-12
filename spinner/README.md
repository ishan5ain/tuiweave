# spinner

`spinner` renders an intrinsic-size activity indicator driven by Bubble Tea commands.

```go
spin := spinner.New(theme)

func (m model) Init() tea.Cmd { return m.spin.Tick() }

// In Update:
m.spin, cmd = m.spin.Update(msg)
```

Forward `spinner.TickMsg` and collect the next command so animation continues. The spinner reports `layout.SizeIntrinsic`; `SetSize` is present for interface compatibility but does not bound the glyph.

See the [demo](../examples/demo) and [spinner recipe](../AGENTS.md#spinner).

## API highlights

- `New(theme)` creates the intrinsic-size model.
- `Tick` returns the next animation command.
- `TickMsg` advances the frame through `Update`.
- `View` renders the current glyph.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/spinner)
- [Spinner recipe](../AGENTS.md#spinner)
