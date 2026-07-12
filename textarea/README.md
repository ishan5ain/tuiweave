# textarea

`textarea` provides bounded multiline editing with soft wrapping, logical-rune selection, history, and kill/yank operations.

```go
ta := textarea.New(theme)
ta.SetSize(width, height)
ta.Focus()
ta, cmd = ta.Update(msg)

pos := ta.CursorPosition()
ta.ReplaceRange(pos, pos, "completion")
```

Coordinates are logical row/rune positions; wrapping and terminal-cell geometry remain internal. Enter inserts a newline, so chat applications should intercept Enter to send and optionally map Alt+Enter to `InsertString("\n")`. `ReplaceRange` is one undoable edit.

See the [textarea autocomplete example](../examples/textarea-autocomplete) and [editing recipe](../AGENTS.md#textinput--textarea).

## API highlights

- `SetValue`, `Value`, and `Reset` manage the editing baseline.
- `CursorPosition` reports logical row/rune coordinates.
- `ReplaceRange` performs one undoable replacement.
- `Undo`, `Redo`, `Yank`, and selection methods expose richer editing workflows.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/textarea)
- [Textarea autocomplete example](../examples/textarea-autocomplete)
- [Text input and textarea recipe](../AGENTS.md#textinput--textarea)
