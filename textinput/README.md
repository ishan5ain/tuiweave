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

## Terminal paste

Forward `tea.PasteMsg` alongside key messages. Only focused inputs accept paste;
`PasteStartMsg` and `PasteEndMsg` are no-ops. Pasted terminal escape sequences and
C0/C1 controls other than LF/tab (including DEL) are removed, and CRLF/CR become
LF. Unicode text is preserved.

Each LF or tab becomes one space, without trimming or collapsing other spaces.
Paste inserts at the cursor and does not submit or emit a command.

This normalization applies to terminal paste; programmatic setters keep their
existing behavior. Applications must refresh filters, completion queries, and
layout after a pasted edit just as after a typed edit.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/textinput)
- [Browser example](../examples/browser)
- [Text input recipe](../AGENTS.md#textinput--textarea)
