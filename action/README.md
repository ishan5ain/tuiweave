# action

`action` defines stable, application-owned actions that can be shared by menus, toolbars, palettes, and semantic inspection.

```go
items := []action.Item{
	{ID: "open", Label: "Open workspace"},
	{ID: "delete", Label: "Delete workspace", Disabled: true},
}
```

Keep IDs stable across updates. Components own selection and rendering; the application owns what each action does. `SelectID` and `ParseSelectID` provide the shared `select.<id>` semantic-action format.

See the [frame example](../examples/frame) and the [action recipe](../AGENTS.md#action-definitions).

## API highlights

- `Item` holds a stable ID, label, description, and disabled state.
- `SelectID` and `ParseSelectID` share semantic selection IDs.
- `Find` and `EnabledCount` query action collections.

## Related documentation

- [API reference](https://pkg.go.dev/github.com/ishan5ain/tuiweave/action)
- [Agent catalog](../AGENT-CATALOG.md#package-map)
