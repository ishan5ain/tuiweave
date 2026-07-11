package menu

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/inspect"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func newTestMenu(width, height int) Model {
	m := New(tuiweave.Dark())
	m.SetSize(width, height)
	m.SetItems(
		Item{ID: "open", Label: "Open workspace", Description: "Open a workspace"},
		Item{ID: "refresh", Label: "Refresh data", Description: "Reload current data"},
		Item{ID: "delete", Label: "Delete workspace", Description: "Destructive", Disabled: true},
		Item{ID: "quit", Label: "Quit", Description: "Close the application"},
	)
	return m
}

func keyPress(key string) tea.KeyPressMsg {
	if len(key) == 1 {
		return tea.KeyPressMsg{Code: rune(key[0]), Text: key}
	}
	return tea.KeyPressMsg{Text: key}
}

func TestMenuFocused(t *testing.T) {
	m := newTestMenu(24, 4)
	m.Focus()
	snaptest.Snap(t, m.View())
	snaptest.SnapCells(t, m.View(), snaptest.WithRoles(tuiweave.Dark()))
}

func TestMenuBlurred(t *testing.T) {
	m := newTestMenu(24, 4)
	m.Select(1)
	snaptest.Snap(t, m.View())
	snaptest.SnapCells(t, m.View(), snaptest.WithRoles(tuiweave.Dark()))
}

func TestMenuSkipsDisabledItemsAndScrolls(t *testing.T) {
	m := newTestMenu(18, 2)
	m.Focus()
	m, _ = m.Update(keyPress("down"))
	if got := m.SelectedID(); got != "refresh" {
		t.Fatalf("first down selected %q, want refresh", got)
	}
	m, _ = m.Update(keyPress("j"))
	if got := m.SelectedID(); got != "quit" {
		t.Fatalf("second navigation selected %q, want quit", got)
	}
	if got := m.YOffset(); got != 2 {
		t.Fatalf("YOffset = %d, want 2", got)
	}
	snaptest.Snap(t, m.View())
}

func TestMenuActivation(t *testing.T) {
	m := newTestMenu(24, 4)
	m.Focus()
	m, _ = m.Update(keyPress("down"))
	m, cmd := m.Update(keyPress("enter"))
	if m.SelectedID() != "refresh" {
		t.Fatalf("selected after navigation = %q, want refresh", m.SelectedID())
	}
	if cmd == nil {
		t.Fatal("enter returned nil command")
	}
	msg, ok := cmd().(SelectedMsg)
	if !ok || msg.ID != "refresh" || msg.Index != 1 {
		t.Fatalf("activation message = %#v, want refresh at index 1", msg)
	}
}

func TestMenuSemanticActionsAndInspection(t *testing.T) {
	m := newTestMenu(24, 4)
	next, cmd, handled := m.applyAction(inspect.Invoke(ActionSelectPrefix + "quit"))
	if !handled || cmd != nil || next.SelectedID() != "quit" {
		t.Fatalf("semantic selection: handled=%v cmd=%v selected=%q", handled, cmd != nil, next.SelectedID())
	}
	node := next.Inspect()
	if node.Kind != "menu" || node.Selected == nil || node.Selected.Label != "Quit" {
		t.Fatalf("unexpected inspection node: %+v", node)
	}
	if node.Scroll == nil || node.Scroll.Total != 4 {
		t.Fatalf("unexpected scroll metadata: %+v", node.Scroll)
	}
	if len(node.Actions) != 11 {
		t.Fatalf("action count = %d, want 11", len(node.Actions))
	}
	for _, action := range node.Actions {
		if action.ID == ActionSelectPrefix+"delete" && action.Enabled {
			t.Fatal("disabled menu item exposed an enabled action")
		}
	}
}

func TestMenuWidthAndEmptyStates(t *testing.T) {
	for width := 1; width <= 24; width++ {
		m := newTestMenu(width, 3)
		for _, line := range splitLines(m.View()) {
			if got := lipgloss.Width(line); got != width {
				t.Fatalf("width %d rendered line as %d", width, got)
			}
		}
	}
	empty := New(tuiweave.Dark())
	empty.SetSize(12, 2)
	if empty.Selected() != -1 || len(splitLines(empty.View())) != 2 {
		t.Fatalf("empty menu state: selected=%d view=%q", empty.Selected(), empty.View())
	}
}

func TestMenuScenarioGolden(t *testing.T) {
	m := scenarioModel{menu: newTestMenu(24, 4)}
	m.menu.Focus()
	result := snaptest.RunScenario(m,
		snaptest.ScenarioStep{Name: "next enabled action", Msg: keyPress("down")},
		snaptest.ScenarioStep{Name: "activate", Msg: keyPress("enter")},
	)
	snaptest.SnapScenario(t, result)
}

type scenarioModel struct {
	menu Model
}

func (m scenarioModel) Init() tea.Cmd { return nil }

func (m scenarioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(msg)
	return m, cmd
}

func (m scenarioModel) View() tea.View { return tea.NewView(m.menu.View()) }

func splitLines(view string) []string {
	var lines []string
	start := 0
	for i, r := range view {
		if r == '\n' {
			lines = append(lines, view[start:i])
			start = i + 1
		}
	}
	return append(lines, view[start:])
}
