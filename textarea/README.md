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

## Terminal paste

Forward `tea.PasteMsg` alongside key messages. Only focused inputs accept paste;
`PasteStartMsg` and `PasteEndMsg` are no-ops. Pasted terminal escape sequences and
C0/C1 controls other than LF/tab (including DEL) are removed, and CRLF/CR become
LF. Unicode text is preserved.

LF and tabs remain in the logical value. Each tab renders as four spaces; when
the row is narrower, it fills that row. Wrapping, cursor movement, and selection
use the same display width while logical positions still count one tab rune. Nonempty paste replaces the selection
as one undoable edit; empty or filtered-empty paste preserves selection and
history. Paste does not submit or emit a command.

This normalization applies to terminal paste; programmatic setters keep their
existing behavior. Applications must refresh filters, completion queries, and
layout after a pasted edit just as after a typed edit.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/textarea)
- [Textarea autocomplete example](../examples/textarea-autocomplete)
- [Text input and textarea recipe](../AGENTS.md#textinput--textarea)
