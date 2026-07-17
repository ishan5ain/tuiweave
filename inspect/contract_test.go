package inspect_test

import (
	"slices"
	"sort"
	"testing"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/action"
	chat "github.com/ishan5ain/tuiweave/agentic/chat"
	diffview "github.com/ishan5ain/tuiweave/agentic/diffview"
	permission "github.com/ishan5ain/tuiweave/agentic/permission"
	toolcall "github.com/ishan5ain/tuiweave/agentic/toolcall"
	"github.com/ishan5ain/tuiweave/autocomplete"
	"github.com/ishan5ain/tuiweave/button"
	"github.com/ishan5ain/tuiweave/dialog"
	"github.com/ishan5ain/tuiweave/inspect"
	"github.com/ishan5ain/tuiweave/list"
	"github.com/ishan5ain/tuiweave/menu"
	"github.com/ishan5ain/tuiweave/palette"
	"github.com/ishan5ain/tuiweave/progress"
	"github.com/ishan5ain/tuiweave/table"
	"github.com/ishan5ain/tuiweave/tabs"
	"github.com/ishan5ain/tuiweave/textarea"
	"github.com/ishan5ain/tuiweave/textinput"
	"github.com/ishan5ain/tuiweave/toggle"
	"github.com/ishan5ain/tuiweave/toolbar"
	"github.com/ishan5ain/tuiweave/viewport"
)

var (
	_ inspect.Inspectable = autocomplete.Model{}
	_ inspect.Inspectable = button.Model{}
	_ inspect.Inspectable = chat.Model{}
	_ inspect.Inspectable = dialog.Model{}
	_ inspect.Inspectable = diffview.Model{}
	_ inspect.Inspectable = list.Model{}
	_ inspect.Inspectable = menu.Model{}
	_ inspect.Inspectable = palette.Model{}
	_ inspect.Inspectable = permission.Model{}
	_ inspect.Inspectable = progress.Model{}
	_ inspect.Inspectable = (*toolcall.Block)(nil)
	_ inspect.Inspectable = table.Model{}
	_ inspect.Inspectable = tabs.Model{}
	_ inspect.Inspectable = textarea.Model{}
	_ inspect.Inspectable = textinput.Model{}
	_ inspect.Inspectable = toggle.Model{}
	_ inspect.Inspectable = toolbar.Model{}
	_ inspect.Inspectable = viewport.Model{}
)

type componentContract struct {
	name       string
	component  inspect.Inspectable
	actionIDs  []string
	attributes []string
	selection  bool
	scroll     bool
	unbounded  bool
}

func TestPublicComponentInspectionContract(t *testing.T) {
	for _, contract := range componentContracts() {
		t.Run(contract.name, func(t *testing.T) {
			node := contract.component.Inspect()
			if node.Kind != contract.name {
				t.Fatalf("kind = %q, want %q", node.Kind, contract.name)
			}
			if !contract.unbounded && (node.Bounds.Width <= 0 || node.Bounds.Height <= 0) {
				t.Fatalf("non-positive local bounds: %+v", node.Bounds)
			}
			if (node.Selected != nil) != contract.selection {
				t.Fatalf("selection presence = %v, want %v", node.Selected != nil, contract.selection)
			}
			if (node.Scroll != nil) != contract.scroll {
				t.Fatalf("scroll presence = %v, want %v", node.Scroll != nil, contract.scroll)
			}
			gotActions := make([]string, len(node.Actions))
			seen := make(map[string]bool, len(node.Actions))
			for i, action := range node.Actions {
				gotActions[i] = action.ID
				if action.ID == "" || action.Label == "" || seen[action.ID] {
					t.Fatalf("invalid action at %d: %+v", i, action)
				}
				seen[action.ID] = true
			}
			if !slices.Equal(gotActions, contract.actionIDs) {
				t.Fatalf("action IDs = %v, want %v", gotActions, contract.actionIDs)
			}
			gotAttributes := make([]string, 0, len(node.Attributes))
			for key := range node.Attributes {
				gotAttributes = append(gotAttributes, key)
			}
			sort.Strings(gotAttributes)
			wantAttributes := append([]string(nil), contract.attributes...)
			sort.Strings(wantAttributes)
			if !slices.Equal(gotAttributes, wantAttributes) {
				t.Fatalf("attribute keys = %v, want %v", gotAttributes, wantAttributes)
			}
		})
	}
}

func TestDisabledItemSelectionActionsRemainDiscoverable(t *testing.T) {
	for _, contract := range componentContracts() {
		if contract.name != "autocomplete" && contract.name != "menu" && contract.name != "palette" && contract.name != "toolbar" {
			continue
		}
		node := contract.component.Inspect()
		found := false
		for _, semanticAction := range node.Actions {
			if semanticAction.ID == action.SelectID("disabled") {
				found = true
				if semanticAction.Enabled {
					t.Errorf("%s disabled selection action is enabled", contract.name)
				}
			}
		}
		if !found {
			t.Errorf("%s omitted disabled selection action", contract.name)
		}
	}
}

func componentContracts() []componentContract {
	theme := tuiweave.Dark()

	completion := autocomplete.New(theme)
	completion.SetSize(24, 3)
	completion.SetItems(
		autocomplete.Item{ID: "open", Value: "open", Label: "Open"},
		autocomplete.Item{ID: "disabled", Value: "delete", Label: "Delete", Disabled: true},
	)

	press := button.New(theme)
	press.ID = "save"
	press.SetLabel("Save")
	press.SetSize(12, 1)

	transcript := chat.New(theme)
	transcript.SetSize(24, 2)
	transcript.Append(chat.CellFunc(func(int) string { return "one\ntwo\nthree" }))

	confirm := dialog.New(theme)
	confirm.Title = "Continue?"
	confirm.Body = "Review the operation."
	confirm.SetSize(30, 8)

	diff := diffview.New(theme)
	diff.SetSize(30, 2)
	diff.SetDiff("--- a/file\n+++ b/file\n-old\n+new")

	items := list.New(theme)
	items.SetSize(20, 2)
	items.SetItems("one", "two", "three")

	commands := menu.New(theme)
	commands.SetSize(24, 2)
	commands.SetItems(
		action.Item{ID: "open", Label: "Open"},
		action.Item{ID: "disabled", Label: "Delete", Disabled: true},
	)

	commandPalette := palette.New(theme)
	commandPalette.SetSize(30, 3)
	commandPalette.SetItems(
		action.Item{ID: "open", Label: "Open"},
		action.Item{ID: "disabled", Label: "Delete", Disabled: true},
	)

	approval := permission.New(theme)
	approval.ID = "deploy"
	approval.Title = "Deploy?"
	approval.Body = "Production"
	approval.SetSize(36, 10)
	tool := toolcall.New(theme, "shell", "Run tests")
	tool.SetID("tool-1")
	tool.AppendOutput("ok")

	completionProgress := progress.New(theme)
	completionProgress.SetSize(24, 1)
	completionProgress.SetLabel("Deploy")
	completionProgress.SetPercent(0.5)

	grid := table.New(theme)
	grid.SetSize(24, 4)
	grid.SetColumns(table.Column{Title: "Name"})
	grid.SetRows([]string{"api"}, []string{"worker"})

	navigation := tabs.New(theme)
	navigation.SetSize(24, 1)
	navigation.SetTabs(tabs.Tab{ID: "overview", Label: "Overview"}, tabs.Tab{ID: "logs", Label: "Logs"})

	area := textarea.New(theme)
	area.SetSize(24, 3)
	area.SetValue("hello")

	input := textinput.New(theme)
	input.SetSize(24, 1)
	input.SetValue("hello")

	setting := toggle.New(theme)
	setting.ID = "refresh"
	setting.SetLabel("Auto-refresh")
	setting.SetSize(24, 1)

	tools := toolbar.New(theme)
	tools.SetSize(30, 1)
	tools.SetItems(
		action.Item{ID: "open", Label: "Open"},
		action.Item{ID: "disabled", Label: "Delete", Disabled: true},
	)

	window := viewport.New(theme)
	window.SetSize(24, 2)
	window.SetContent("one\ntwo\nthree")

	return []componentContract{
		{name: "autocomplete", component: completion, actionIDs: []string{"focus", "blur", "next", "previous", "first", "last", "clear", "activate", "select.open", "select.disabled"}, attributes: []string{"item_count", "filtered_count", "query", "selected_id"}, selection: true, scroll: true},
		{name: "button", component: press, actionIDs: []string{"focus", "blur", "activate"}, attributes: []string{"disabled", "id"}},
		{name: "chat", component: transcript, actionIDs: []string{"focus", "blur", "scroll_up", "scroll_down", "scroll_top", "scroll_bottom"}, attributes: []string{"cell_count", "following"}, scroll: true},
		{name: "dialog", component: confirm, actionIDs: []string{"confirm", "cancel", "next", "previous"}, attributes: []string{"body"}, selection: true},
		{name: "diffview", component: diff, actionIDs: []string{"focus", "blur", "scroll_up", "scroll_down", "scroll_top", "scroll_bottom"}, scroll: true},
		{name: "list", component: items, actionIDs: []string{"focus", "blur", "next", "previous", "first", "last"}, attributes: []string{"item_count", "filtered_count"}, selection: true, scroll: true},
		{name: "menu", component: commands, actionIDs: []string{"focus", "blur", "next", "previous", "first", "last", "activate", "select.open", "select.disabled"}, attributes: []string{"item_count", "enabled_count", "selected_id"}, selection: true, scroll: true},
		{name: "palette", component: commandPalette, actionIDs: []string{"focus", "blur", "next", "previous", "first", "last", "clear", "activate", "select.open", "select.disabled"}, attributes: []string{"item_count", "filtered_count", "query", "selected_id"}, selection: true, scroll: true},
		{name: "permission", component: approval, actionIDs: []string{"choose.1", "choose.2", "choose.3"}, attributes: []string{"body"}, selection: true},
		{name: "progress", component: completionProgress, attributes: []string{"percent", "show_percent"}},
		{name: "table", component: grid, actionIDs: []string{"focus", "blur", "next", "previous", "first", "last"}, attributes: []string{"column_count", "row_count"}, selection: true, scroll: true},
		{name: "tabs", component: navigation, actionIDs: []string{"focus", "blur", "next", "previous", "first", "last", "select.overview", "select.logs"}, attributes: []string{"tab_count", "visible_start", "selected_id"}, selection: true},
		{name: "textarea", component: area, actionIDs: []string{"focus", "blur", "clear", "undo", "redo", "yank", "select_all", "clear_selection"}, attributes: []string{"empty", "logical_lines", "content_height", "cursor_row", "cursor_column", "has_selection", "selected_runes", "can_undo", "can_redo"}},
		{name: "textinput", component: input, actionIDs: []string{"focus", "blur", "clear"}, attributes: []string{"empty", "placeholder", "cursor"}},
		{name: "toolcall", component: tool, actionIDs: []string{"cancel", "retry"}, attributes: []string{"summary", "attempt", "output_lines"}, unbounded: true},
		{name: "toggle", component: setting, actionIDs: []string{"focus", "blur", "toggle", "on", "off"}, attributes: []string{"checked", "disabled", "id"}},
		{name: "toolbar", component: tools, actionIDs: []string{"focus", "blur", "next", "previous", "first", "last", "activate", "select.open", "select.disabled"}, attributes: []string{"item_count", "enabled_count", "selected_id"}, selection: true, scroll: true},
		{name: "viewport", component: window, actionIDs: []string{"focus", "blur", "scroll_up", "scroll_down", "scroll_top", "scroll_bottom"}, scroll: true},
	}
}
