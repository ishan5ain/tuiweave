package toolbar

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/inspect"
	"github.com/ishansain/gotui/snaptest"
)

func newTestToolbar(width int) Model {
	m := New(gotui.Dark())
	m.SetSize(width, 1)
	m.SetItems(
		Item{ID: "refresh", Label: "Refresh", Description: "Reload data"},
		Item{ID: "export", Label: "Export", Description: "Export data"},
		Item{ID: "delete", Label: "Delete", Description: "Destructive", Disabled: true},
		Item{ID: "settings", Label: "Settings", Description: "Open settings"},
	)
	return m
}

func keyPress(key string) tea.KeyPressMsg {
	if len(key) == 1 {
		return tea.KeyPressMsg{Code: rune(key[0]), Text: key}
	}
	return tea.KeyPressMsg{Text: key}
}

func TestToolbarFocused(t *testing.T) {
	m := newTestToolbar(36)
	m.Focus()
	m.Select(1)
	snaptest.Snap(t, m.View())
	snaptest.SnapCells(t, m.View(), snaptest.WithRoles(gotui.Dark()))
}

func TestToolbarBlurred(t *testing.T) {
	m := newTestToolbar(36)
	m.Select(1)
	snaptest.Snap(t, m.View())
	snaptest.SnapCells(t, m.View(), snaptest.WithRoles(gotui.Dark()))
}

func TestToolbarSkipsDisabledAndKeepsSelectionVisible(t *testing.T) {
	m := newTestToolbar(14)
	m.Focus()
	m, _ = m.Update(keyPress("right"))
	if got := m.SelectedID(); got != "export" {
		t.Fatalf("first right selected %q, want export", got)
	}
	m, _ = m.Update(keyPress("l"))
	if got := m.SelectedID(); got != "settings" {
		t.Fatalf("second navigation selected %q, want settings", got)
	}
	if !strings.Contains(stripANSI(m.View()), "Settings") {
		t.Fatalf("selected action is not visible: %q", stripANSI(m.View()))
	}
	snaptest.Snap(t, m.View())
}

func TestToolbarActivation(t *testing.T) {
	m := newTestToolbar(36)
	m.Focus()
	m, _ = m.Update(keyPress("right"))
	m, cmd := m.Update(keyPress("enter"))
	if m.SelectedID() != "export" {
		t.Fatalf("selected after navigation = %q, want export", m.SelectedID())
	}
	if cmd == nil {
		t.Fatal("enter returned nil command")
	}
	msg, ok := cmd().(SelectedMsg)
	if !ok || msg.ID != "export" || msg.Index != 1 {
		t.Fatalf("activation message = %#v, want export at index 1", msg)
	}
}

func TestToolbarSemanticActionsAndInspection(t *testing.T) {
	m := newTestToolbar(36)
	next, cmd, handled := m.applyAction(inspect.Invoke(ActionSelectPrefix + "settings"))
	if !handled || cmd != nil || next.SelectedID() != "settings" {
		t.Fatalf("semantic selection: handled=%v cmd=%v selected=%q", handled, cmd != nil, next.SelectedID())
	}
	node := next.Inspect()
	if node.Kind != "toolbar" || node.Selected == nil || node.Selected.Label != "Settings" {
		t.Fatalf("unexpected inspection node: %+v", node)
	}
	if node.Scroll == nil || node.Scroll.Total != 4 || node.Scroll.Visible == 0 {
		t.Fatalf("unexpected scroll metadata: %+v", node.Scroll)
	}
	if len(node.Actions) != 11 {
		t.Fatalf("action count = %d, want 11", len(node.Actions))
	}
	for _, action := range node.Actions {
		if action.ID == ActionSelectPrefix+"delete" && action.Enabled {
			t.Fatal("disabled toolbar action exposed as enabled")
		}
	}
}

func TestToolbarWidthAndEmptyStates(t *testing.T) {
	for width := 1; width <= 28; width++ {
		m := newTestToolbar(width)
		if got := lipgloss.Width(m.View()); got != width {
			t.Fatalf("width %d rendered as %d", width, got)
		}
		empty := New(gotui.Dark())
		empty.SetSize(width, 1)
		if got := lipgloss.Width(empty.View()); got != width {
			t.Fatalf("empty width %d rendered as %d", width, got)
		}
	}
}

func TestToolbarScenarioGolden(t *testing.T) {
	m := scenarioModel{toolbar: newTestToolbar(36)}
	m.toolbar.Focus()
	result := snaptest.RunScenario(m,
		snaptest.ScenarioStep{Name: "next action", Msg: keyPress("right")},
		snaptest.ScenarioStep{Name: "activate", Msg: keyPress("enter")},
	)
	snaptest.SnapScenario(t, result)
}

type scenarioModel struct {
	toolbar Model
}

func (m scenarioModel) Init() tea.Cmd { return nil }

func (m scenarioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.toolbar, cmd = m.toolbar.Update(msg)
	return m, cmd
}

func (m scenarioModel) View() tea.View { return tea.NewView(m.toolbar.View()) }

func stripANSI(value string) string {
	return ansi.Strip(value)
}
