# table

`table` displays selectable records in fixed and flexible columns.

```go
t := table.New(theme)
t.SetColumns(
	table.Column{Title: "ID", Width: 4},
	table.Column{Title: "Name"}, // width 0 is flexible
)
t.SetRows([]string{"1", "api"}, []string{"2", "worker"})
t.SetSize(width, height)
```

Keyboard navigation is focus-gated. The header and rule consume two rows; columns shrink to stay inside the assigned width, and cell text is truncated by visible terminal width. Align a standalone scrollbar with the row area.

See the [table example](../examples/table) and [scrolling recipe](../AGENTS.md#viewport--list--table-scrolling-components).

## API highlights

- `SetColumns` defines fixed and flexible columns.
- `SetRows` replaces the record set.
- `Selected` and scrolling accessors expose selection state.
- `SetSize`, `Update`, and `View` implement the bounded component contract.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/table)
- [Table example](../examples/table)
- [Scrolling components](../AGENTS.md#viewport--list--table-scrolling-components)
