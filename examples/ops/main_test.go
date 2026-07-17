package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/button"
	"github.com/ishan5ain/tuiweave/dialog"
	"github.com/ishan5ain/tuiweave/inspect"
	"github.com/ishan5ain/tuiweave/palette"
	"github.com/ishan5ain/tuiweave/snaptest"
	"github.com/ishan5ain/tuiweave/toggle"
)

func sized(t *testing.T) model {
	return sizedAt(t, 80, 22)
}

func sizedAt(t *testing.T, width, height int) model {
	t.Helper()
	m := newModel()
	next, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return next.(model)
}

func TestOpsScreenGolden(t *testing.T) {
	m := sized(t)
	snaptest.Snap(t, m.render())
	snaptest.SnapCells(t, m.render(), snaptest.WithRoles(tuiweave.Dark()))
}

func TestOpsTabAndFocusComposition(t *testing.T) {
	m := sized(t)
	if m.fm.Index() != 0 || !m.tabs.Focused() {
		t.Fatalf("initial focus index=%d tabs=%v", m.fm.Index(), m.tabs.Focused())
	}
	m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyRight})
	if m.tabs.SelectedID() != "jobs" || m.rows.SelectedRow()[0] != "nightly-backup" {
		t.Fatalf("tab selection: tab=%q row=%v", m.tabs.SelectedID(), m.rows.SelectedRow())
	}
	for i := 1; i <= 4; i++ {
		m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
		if m.fm.Index() != i {
			t.Fatalf("tab %d focus index = %d", i, m.fm.Index())
		}
	}
	if !m.openLogs.Focused() {
		t.Fatal("focus cycle did not reach open logs button")
	}
}

func TestOpsCommandPaletteGolden(t *testing.T) {
	m := sized(t)
	m, _ = update(t, m, tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})
	if !m.showPalette {
		t.Fatal("ctrl+p did not open command palette")
	}
	snaptest.Snap(t, m.render())
	snaptest.SnapCells(t, m.render(), snaptest.WithRoles(tuiweave.Dark()))
}

func TestOpsMouseRoutesRootControls(t *testing.T) {
	m := sized(t)
	if got := m.View().MouseMode; got != tea.MouseModeCellMotion {
		t.Fatalf("mouse mode = %v, want cell motion", got)
	}

	m, _ = update(t, m, click(m.actionsArea.Min.X, m.actionsArea.Min.Y+1))
	if m.fm.Index() != 1 || !m.actions.Focused() || m.actions.SelectedID() != "drain" {
		t.Fatalf("action click: focus=%d focused=%v selected=%q", m.fm.Index(), m.actions.Focused(), m.actions.SelectedID())
	}

	m, _ = update(t, m, click(m.rowsArea.Min.X, m.rowsArea.Min.Y+3))
	if m.fm.Index() != 2 || !m.rows.Focused() || m.rows.Selected() != 1 {
		t.Fatalf("table click: focus=%d focused=%v selected=%d", m.fm.Index(), m.rows.Focused(), m.rows.Selected())
	}

	var cmd tea.Cmd
	m, cmd = update(t, m, click(m.autoRefreshArea.Min.X, m.autoRefreshArea.Min.Y))
	if m.fm.Index() != 3 || !m.autoRefresh.Focused() || m.autoRefresh.Checked() || cmd == nil {
		t.Fatalf("toggle click: focus=%d focused=%v checked=%v command=%v", m.fm.Index(), m.autoRefresh.Focused(), m.autoRefresh.Checked(), cmd != nil)
	}
	if m.notice != "ready" {
		t.Fatalf("toggle changed notice before command delivery: %q", m.notice)
	}
	changed, ok := cmd().(toggle.ChangedMsg)
	if !ok || changed.ID != "auto-refresh" || changed.Checked {
		t.Fatalf("toggle click result = %#v", changed)
	}
	m, _ = update(t, m, changed)
	if m.notice != "auto-refresh false" {
		t.Fatalf("delivered toggle notice = %q", m.notice)
	}

	m, cmd = update(t, m, click(m.openLogsArea.Min.X, m.openLogsArea.Min.Y))
	if m.fm.Index() != 4 || !m.openLogs.Focused() || cmd == nil {
		t.Fatalf("button click: focus=%d focused=%v command=%v", m.fm.Index(), m.openLogs.Focused(), cmd != nil)
	}
	pressed, ok := cmd().(button.PressedMsg)
	if !ok || pressed.ID != "open-logs" {
		t.Fatalf("button click result = %#v", pressed)
	}
	m, _ = update(t, m, pressed)
	if m.notice != "opened logs" {
		t.Fatalf("delivered button notice = %q", m.notice)
	}
}

func TestOpsMouseDoesNotReachBackgroundUnderOverlays(t *testing.T) {
	m := sized(t)
	m, _ = update(t, m, tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})
	m, cmd := update(t, m, click(m.autoRefreshArea.Min.X, m.autoRefreshArea.Min.Y))
	if cmd != nil || m.fm.Depth() != 1 || !m.commands.Focused() || !m.autoRefresh.Checked() {
		t.Fatalf("palette background click: command=%v depth=%d palette=%v checked=%v", cmd != nil, m.fm.Depth(), m.commands.Focused(), m.autoRefresh.Checked())
	}

	m, _ = update(t, m, palette.SelectedMsg{ID: "restart", Label: "Restart service"})
	m, cmd = update(t, m, click(m.actionsArea.Min.X, m.actionsArea.Min.Y+2))
	if cmd != nil || m.fm.Depth() != 2 || !m.showConfirm || m.actions.SelectedID() != "restart" {
		t.Fatalf("confirmation background click: command=%v depth=%d confirmation=%v selected=%q", cmd != nil, m.fm.Depth(), m.showConfirm, m.actions.SelectedID())
	}
}

func TestOpsCommandPaletteActivation(t *testing.T) {
	m := sized(t)
	m, _ = update(t, m, tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})
	if !m.fm.Active() || m.fm.Depth() != 1 || !m.commands.Focused() || m.tabs.Focused() {
		t.Fatalf("palette scope: active=%v depth=%d palette=%v tabs=%v", m.fm.Active(), m.fm.Depth(), m.commands.Focused(), m.tabs.Focused())
	}
	for _, r := range "refresh" {
		m, _ = update(t, m, tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	m, cmd := update(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("palette enter returned no command")
	}
	if m.showPalette == false {
		// The command is intentionally not executed by this test; palette
		// visibility closes when the resulting message is delivered below.
		t.Fatal("palette closed before SelectedMsg was delivered")
	}
	selected, ok := cmd().(palette.SelectedMsg)
	if !ok || selected.ID != "refresh" {
		t.Fatalf("palette message = %#v, want refresh", selected)
	}
	next, _ := m.Update(selected)
	m = next.(model)
	if m.showPalette || m.notice != "ran Refresh data" {
		t.Fatalf("palette result: open=%v notice=%q", m.showPalette, m.notice)
	}
	if m.fm.Active() || m.fm.Depth() != 0 || !m.tabs.Focused() || m.commands.Focused() {
		t.Fatalf("restored focus: active=%v depth=%d tabs=%v palette=%v", m.fm.Active(), m.fm.Depth(), m.tabs.Focused(), m.commands.Focused())
	}
}

func TestOpsNestedConfirmationCancellation(t *testing.T) {
	m := sized(t)
	m, _ = update(t, m, tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})
	m, _ = update(t, m, palette.SelectedMsg{ID: "restart", Label: "Restart service"})
	if !m.showPalette || !m.showConfirm || m.fm.Depth() != 2 || m.commands.Focused() {
		t.Fatalf("nested confirmation: palette=%v confirm=%v depth=%d paletteFocused=%v", m.showPalette, m.showConfirm, m.fm.Depth(), m.commands.Focused())
	}

	m, cmd := update(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("confirmation escape returned no result command")
	}
	result, ok := cmd().(dialog.ResultMsg)
	if !ok || result.ID != "restart" || result.OK {
		t.Fatalf("confirmation result = %#v, want restart cancellation", result)
	}
	m, _ = update(t, m, result)
	if !m.showPalette || m.showConfirm || m.fm.Depth() != 1 || !m.commands.Focused() || m.tabs.Focused() {
		t.Fatalf("after cancellation: palette=%v confirm=%v depth=%d paletteFocused=%v tabsFocused=%v", m.showPalette, m.showConfirm, m.fm.Depth(), m.commands.Focused(), m.tabs.Focused())
	}
}

func TestOpsInspectionGolden(t *testing.T) {
	m := sized(t)
	m, _ = update(t, m, tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})
	node := m.Inspect()
	data, err := inspect.Marshal(node)
	if err != nil {
		t.Fatal(err)
	}
	snaptest.Snap(t, string(data))
}

func TestOpsSemanticActionRouting(t *testing.T) {
	m := sized(t)
	m, _ = update(t, m, tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})

	commands := requireInspectNode(t, m.Inspect(), "commands")
	if commands.Bounds != inspect.FromRect(m.commandsArea) {
		t.Fatalf("commands bounds = %+v, want %+v", commands.Bounds, inspect.FromRect(m.commandsArea))
	}
	if !inspectActionEnabled(commands, "commands.select.restart") {
		t.Fatal("commands.select.restart is not enabled")
	}
	t.Run("palette-open", func(t *testing.T) {
		snaptest.Snap(t, marshalInspection(t, m))
	})

	var cmd tea.Cmd
	m, cmd = update(t, m, inspect.Invoke("commands.select.restart"))
	if cmd != nil || m.commands.SelectedID() != "restart" {
		t.Fatalf("semantic select: command=%v selected=%q", cmd != nil, m.commands.SelectedID())
	}
	m, cmd = update(t, m, inspect.Invoke("commands.activate"))
	if cmd == nil {
		t.Fatal("commands.activate returned no command")
	}
	if m.showConfirm {
		t.Fatal("confirmation opened before SelectedMsg was delivered")
	}
	selected, ok := cmd().(palette.SelectedMsg)
	if !ok || selected.ID != "restart" {
		t.Fatalf("commands.activate result = %#v, want restart selection", selected)
	}
	m, _ = update(t, m, selected)
	if !m.showPalette || !m.showConfirm {
		t.Fatalf("delivered restart selection: palette=%v confirmation=%v", m.showPalette, m.showConfirm)
	}

	confirm := requireInspectNode(t, m.Inspect(), "confirm-restart")
	if confirm.Bounds != inspect.FromRect(m.confirmArea) {
		t.Fatalf("confirmation bounds = %+v, want %+v", confirm.Bounds, inspect.FromRect(m.confirmArea))
	}
	for _, id := range []string{"confirm-restart.confirm", "confirm-restart.cancel"} {
		if !inspectActionEnabled(confirm, id) {
			t.Fatalf("%s is not enabled", id)
		}
	}
	t.Run("confirmation-visible", func(t *testing.T) {
		snaptest.Snap(t, marshalInspection(t, m))
	})

	m, cmd = update(t, m, inspect.Invoke("confirm-restart.cancel"))
	if cmd == nil || !m.showConfirm {
		t.Fatalf("semantic cancel: command=%v confirmation=%v", cmd != nil, m.showConfirm)
	}
	result, ok := cmd().(dialog.ResultMsg)
	if !ok || result.ID != "restart" || result.OK {
		t.Fatalf("semantic cancel result = %#v", result)
	}
	m, _ = update(t, m, result)
	if !m.showPalette || m.showConfirm || m.notice != "restart cancelled" {
		t.Fatalf("cancel result: palette=%v confirmation=%v notice=%q", m.showPalette, m.showConfirm, m.notice)
	}

	// Repeat from a clean model so the positive result covers the same
	// command-delivery boundary as cancellation.
	m = sized(t)
	m, _ = update(t, m, tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})
	m, _ = update(t, m, inspect.Invoke("commands.select.restart"))
	m, cmd = update(t, m, inspect.Invoke("commands.activate"))
	selected, ok = cmd().(palette.SelectedMsg)
	if !ok {
		t.Fatalf("repeat activation result = %#v", selected)
	}
	m, _ = update(t, m, selected)
	m, cmd = update(t, m, inspect.Invoke("confirm-restart.confirm"))
	if cmd == nil {
		t.Fatal("semantic confirm returned no command")
	}
	result, ok = cmd().(dialog.ResultMsg)
	if !ok || result.ID != "restart" || !result.OK {
		t.Fatalf("semantic confirm result = %#v", result)
	}
	m, _ = update(t, m, result)
	if m.showPalette || m.showConfirm || m.notice != "restarted service" {
		t.Fatalf("confirm result: palette=%v confirmation=%v notice=%q", m.showPalette, m.showConfirm, m.notice)
	}

	// The dispatcher must reject actions that are not in the current visible
	// tree, disabled actions, and unknown IDs without changing state.
	assertRejectedSemanticAction(t, sized(t), "commands.select.restart")

	m = sized(t)
	m, _ = update(t, m, tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})
	assertRejectedSemanticAction(t, m, "commands.select.delete")
	assertRejectedSemanticAction(t, m, "commands.unknown")
}

func TestOpsScenarioGolden(t *testing.T) {
	result := snaptest.RunScenario(sized(t),
		snaptest.ScenarioStep{Name: "focus actions", Msg: tea.KeyPressMsg{Code: tea.KeyTab}},
		snaptest.ScenarioStep{Name: "navigate action", Msg: tea.KeyPressMsg{Code: tea.KeyDown}},
		snaptest.ScenarioStep{Name: "open commands", Msg: tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl}},
		snaptest.ScenarioStep{Name: "request restart", Msg: palette.SelectedMsg{ID: "restart", Label: "Restart service"}},
		snaptest.ScenarioStep{Name: "request cancellation", Msg: tea.KeyPressMsg{Code: tea.KeyEscape}},
		snaptest.ScenarioStep{Name: "deliver cancellation", Msg: dialog.ResultMsg{ID: "restart", OK: false}},
		snaptest.ScenarioStep{Name: "close commands", Msg: palette.SelectedMsg{ID: "refresh", Label: "Refresh data"}},
	)
	snaptest.SnapScenario(t, result)
}

func TestOpsNarrowScenarioGolden(t *testing.T) {
	result := snaptest.RunScenario(sizedAt(t, 36, 12),
		snaptest.ScenarioStep{Name: "open commands", Msg: tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl}},
		snaptest.ScenarioStep{Name: "open restart confirmation", Msg: palette.SelectedMsg{ID: "restart", Label: "Restart service"}},
		snaptest.ScenarioStep{Name: "cancel confirmation", Msg: dialog.ResultMsg{ID: "restart", OK: false}},
	)
	snaptest.SnapScenario(t, result)
}

func TestOpsMouseScenarioGolden(t *testing.T) {
	m := sized(t)
	result := snaptest.RunScenario(m,
		snaptest.ScenarioStep{Name: "select drain action", Msg: click(m.actionsArea.Min.X, m.actionsArea.Min.Y+1)},
		snaptest.ScenarioStep{Name: "select worker row", Msg: click(m.rowsArea.Min.X, m.rowsArea.Min.Y+3)},
		snaptest.ScenarioStep{Name: "toggle auto refresh", Msg: click(m.autoRefreshArea.Min.X, m.autoRefreshArea.Min.Y)},
		snaptest.ScenarioStep{Name: "deliver toggle result", Msg: toggle.ChangedMsg{ID: "auto-refresh", Checked: false}},
		snaptest.ScenarioStep{Name: "open commands", Msg: tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl}},
		snaptest.ScenarioStep{Name: "ignore background click", Msg: click(m.openLogsArea.Min.X, m.openLogsArea.Min.Y)},
	)
	snaptest.SnapScenario(t, result)
}

func click(x, y int) tea.MouseClickMsg {
	return tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft}
}

func update(t *testing.T, m model, msg tea.Msg) (model, tea.Cmd) {
	t.Helper()
	next, cmd := m.Update(msg)
	return next.(model), cmd
}

func marshalInspection(t *testing.T, m model) string {
	t.Helper()
	data, err := inspect.Marshal(m.Inspect())
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func requireInspectNode(t *testing.T, root inspect.Node, id string) inspect.Node {
	t.Helper()
	if root.ID == id {
		return root
	}
	for _, child := range root.Children {
		if node := findInspectNode(child, id); node.ID != "" {
			return node
		}
	}
	t.Fatalf("inspection node %q not found", id)
	return inspect.Node{}
}

func findInspectNode(node inspect.Node, id string) inspect.Node {
	if node.ID == id {
		return node
	}
	for _, child := range node.Children {
		if found := findInspectNode(child, id); found.ID != "" {
			return found
		}
	}
	return inspect.Node{}
}

func inspectActionEnabled(node inspect.Node, id string) bool {
	for _, action := range node.Actions {
		if action.ID == id {
			return action.Enabled
		}
	}
	return false
}

func assertRejectedSemanticAction(t *testing.T, m model, id string) {
	t.Helper()
	before := marshalInspection(t, m)
	next, cmd := m.Update(inspect.Invoke(id))
	if cmd != nil {
		t.Fatalf("rejected action %q returned a command", id)
	}
	got := next.(model)
	if after := marshalInspection(t, got); after != before {
		t.Fatalf("rejected action %q mutated inspection state", id)
	}
}
