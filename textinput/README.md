# textinput

`textinput` provides bounded, cell-aware single-line editing.

```go
input := textinput.New(theme)
input.Placeholder = "Search…"
input.SetSize(width, 1)
input.Focus()
input, cmd = input.Update(msg)
```

The component owns cursor movement, editing, and horizontal windowing across wide and combining graphemes. Enter is intentionally not handled: intercept it in the application, read `Value()`, and call `Reset()` after submission.

See the [browser example](../examples/browser) and [text input recipe](../AGENTS.md#textinput--textarea).

## API highlights

- `SetValue`, `Value`, and `Reset` manage application-visible text.
- `SetSize` controls the cell-aware editing window.
- `Focus`, `Blur`, and `Focused` manage keyboard participation.
- `Update` handles editing but leaves Enter/submission to the app.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/textinput)
- [Browser example](../examples/browser)
- [Text input recipe](../AGENTS.md#textinput--textarea)
