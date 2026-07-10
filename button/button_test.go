package button

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/inspect"
	"github.com/ishansain/gotui/snaptest"
)

func newTestButton(width int) Model {
	m := New(gotui.Dark())
	m.ID = "open"
	m.SetLabel("Open workspace")
	m.SetSize(width, 1)
	return m
}

func keyPress(key string) tea.KeyPressMsg {
	if len(key) == 1 {
		return tea.KeyPressMsg{Code: rune(key[0]), Text: key}
	}
	return tea.KeyPressMsg{Text: key}
}

func TestButtonDefaultGolden(t *testing.T) {
	m := newTestButton(28)
	snaptest.Snap(t, m.View())
	snaptest.SnapCells(t, m.View(), snaptest.WithRoles(gotui.Dark()))
}

func TestButtonFocusedGolden(t *testing.T) {
	m := newTestButton(28)
	m.Focus()
	snaptest.Snap(t, m.View())
	snaptest.SnapCells(t, m.View(), snaptest.WithRoles(gotui.Dark()))
}

func TestButtonActivation(t *testing.T) {
	m := newTestButton(28)
	if _, cmd := m.Update(keyPress("enter")); cmd != nil {
		t.Fatal("blurred button activated")
	}
	m.Focus()
	m, cmd := m.Update(keyPress("space"))
	if cmd == nil {
		t.Fatal("focused space returned no activation command")
	}
	msg, ok := cmd().(PressedMsg)
	if !ok || msg.ID != "open" || msg.Label != "Open workspace" {
		t.Fatalf("activation message = %#v, want open workspace", msg)
	}
	_ = m
}

func TestButtonSemanticActionsAndInspection(t *testing.T) {
	m := newTestButton(28)
	next, cmd, handled := m.applyAction(inspect.Invoke(ActionFocus))
	if !handled || cmd != nil || !next.Focused() {
		t.Fatalf("focus action: handled=%v command=%v focused=%v", handled, cmd != nil, next.Focused())
	}
	next, cmd, handled = next.applyAction(inspect.Invoke(ActionActivate))
	if !handled || cmd == nil {
		t.Fatalf("activate action: handled=%v command=%v", handled, cmd != nil)
	}
	node := next.Inspect()
	if node.Kind != "button" || node.Label != "Open workspace" || node.Status != "ready" || !node.Focused {
		t.Fatalf("unexpected inspection node: %+v", node)
	}
	if node.Attributes["id"] != "open" || len(node.Actions) != 3 {
		t.Fatalf("unexpected inspection metadata: %+v", node)
	}
}

func TestButtonDisabled(t *testing.T) {
	m := newTestButton(28)
	m.SetDisabled(true)
	m.Focus()
	if m.Focused() {
		t.Fatal("disabled button accepted focus")
	}
	if _, cmd := m.Update(inspect.Invoke(ActionActivate)); cmd != nil {
		t.Fatal("disabled button emitted activation")
	}
	snaptest.Snap(t, m.View())
}

func TestButtonExactWidth(t *testing.T) {
	for width := 1; width <= 40; width++ {
		m := newTestButton(width)
		m.SetLabel("A very long action label")
		if got := lipgloss.Width(m.View()); got != width {
			t.Fatalf("width %d rendered as %d", width, got)
		}
	}

	m := newTestButton(10)
	m.SetSize(10, 0)
	if got := m.View(); got != "" {
		t.Fatalf("zero-height view = %q, want empty", got)
	}
}

func TestButtonScenarioGolden(t *testing.T) {
	m := scenarioModel{button: newTestButton(28)}
	m.button.Focus()
	result := snaptest.RunScenario(m,
		snaptest.ScenarioStep{Name: "activate", Msg: keyPress("enter")},
		snaptest.ScenarioStep{Name: "activate semantically", Msg: inspect.Invoke(ActionActivate)},
	)
	snaptest.SnapScenario(t, result)
}

type scenarioModel struct {
	button Model
}

func (m scenarioModel) Init() tea.Cmd { return nil }

func (m scenarioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.button, cmd = m.button.Update(msg)
	return m, cmd
}

func (m scenarioModel) View() tea.View { return tea.NewView(m.button.View()) }
