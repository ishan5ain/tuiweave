package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/button"
	"github.com/ishansain/gotui/snaptest"
	"github.com/ishansain/gotui/toggle"
	"github.com/ishansain/gotui/toolbar"
)

func TestFrameExampleGolden(t *testing.T) {
	m := newModel()
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 68, Height: 13})
	snaptest.Snap(t, m.render())
}

func TestFrameFocusCyclesAllInteractiveComponents(t *testing.T) {
	m := newModel()
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 68, Height: 13})
	assertFocus(t, m, 0, true, false, false, false, false)

	m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	assertFocus(t, m, 1, false, true, false, false, false)
	m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	assertFocus(t, m, 2, false, false, true, false, false)

	// The focused toolbar is a live component, not decorative output.
	m, cmd := update(t, m, tea.KeyPressMsg{Code: 'l', Text: "l"})
	if cmd != nil || m.tools.SelectedID() != "export" {
		t.Fatalf("toolbar navigation: cmd=%v selected=%q", cmd != nil, m.tools.SelectedID())
	}
	m, cmd = update(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("toolbar enter returned no activation command")
	}
	msg, ok := cmd().(toolbar.SelectedMsg)
	if !ok || msg.ID != "export" {
		t.Fatalf("toolbar activation = %#v, want export", msg)
	}
	if _, cmd = update(t, m, msg); cmd != nil {
		t.Fatal("showcase produced a command while handling toolbar activation")
	}

	m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	assertFocus(t, m, 3, false, false, false, true, false)
	m, cmd = update(t, m, tea.KeyPressMsg{Code: ' ', Text: " "})
	if cmd == nil || m.autoRefresh.Checked() {
		t.Fatalf("toggle space: checked=%v command=%v, want unchecked with command", m.autoRefresh.Checked(), cmd != nil)
	}
	toggleMsg, ok := cmd().(toggle.ChangedMsg)
	if !ok || toggleMsg.ID != "auto-refresh" || toggleMsg.Checked {
		t.Fatalf("toggle change = %#v, want auto-refresh unchecked", toggleMsg)
	}
	m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	assertFocus(t, m, 4, false, false, false, false, true)
	m, cmd = update(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil || m.open.Focused() == false {
		t.Fatalf("button enter: focused=%v command=%v, want activation command", m.open.Focused(), cmd != nil)
	}
	buttonMsg, ok := cmd().(button.PressedMsg)
	if !ok || buttonMsg.ID != "open-workspace" {
		t.Fatalf("button activation = %#v, want open-workspace", buttonMsg)
	}

	// Reverse traversal reaches every interactive control as well.
	m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	assertFocus(t, m, 3, false, false, false, true, false)
	m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	assertFocus(t, m, 2, false, false, true, false, false)
	m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	assertFocus(t, m, 1, false, true, false, false, false)
	m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	assertFocus(t, m, 0, true, false, false, false, false)
	m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	assertFocus(t, m, 4, false, false, false, false, true)
}

func TestFrameToolbarFocusedGolden(t *testing.T) {
	m := newModel()
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 68, Height: 13})
	m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	snaptest.Snap(t, m.render())
	snaptest.SnapCells(t, m.render(), snaptest.WithRoles(gotui.Dark()))
}

func update(t *testing.T, m model, msg tea.Msg) (model, tea.Cmd) {
	t.Helper()
	next, cmd := m.Update(msg)
	return next.(model), cmd
}

func assertFocus(t *testing.T, m model, index int, nav, actions, tools, autoRefresh, open bool) {
	t.Helper()
	if m.fm.Index() != index || m.nav.Focused() != nav || m.actions.Focused() != actions || m.tools.Focused() != tools || m.autoRefresh.Focused() != autoRefresh || m.open.Focused() != open {
		t.Fatalf("focus index=%d nav=%v actions=%v tools=%v auto-refresh=%v open=%v; want index=%d nav=%v actions=%v tools=%v auto-refresh=%v open=%v", m.fm.Index(), m.nav.Focused(), m.actions.Focused(), m.tools.Focused(), m.autoRefresh.Focused(), m.open.Focused(), index, nav, actions, tools, autoRefresh, open)
	}
}
