package toggle

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/inspect"
	"github.com/ishansain/gotui/snaptest"
)

func newTestToggle(width int) Model {
	m := New(gotui.Dark())
	m.ID = "auto-refresh"
	m.SetLabel("Auto-refresh")
	m.SetSize(width, 1)
	return m
}

func keyPress(key string) tea.KeyPressMsg {
	if len(key) == 1 {
		return tea.KeyPressMsg{Code: rune(key[0]), Text: key}
	}
	return tea.KeyPressMsg{Text: key}
}

func TestToggleDefaultGolden(t *testing.T) {
	m := newTestToggle(28)
	snaptest.Snap(t, m.View())
	snaptest.SnapCells(t, m.View(), snaptest.WithRoles(gotui.Dark()))
}

func TestToggleFocusedGolden(t *testing.T) {
	m := newTestToggle(28)
	m.Focus()
	m.SetChecked(true)
	snaptest.Snap(t, m.View())
	snaptest.SnapCells(t, m.View(), snaptest.WithRoles(gotui.Dark()))
}

func TestToggleChangesAndEmitsMessage(t *testing.T) {
	m := newTestToggle(28)
	if next, cmd := m.Update(keyPress("space")); cmd != nil || next.Checked() {
		t.Fatalf("blurred toggle changed: checked=%v command=%v", next.Checked(), cmd != nil)
	}
	m.Focus()
	m, cmd := m.Update(keyPress("space"))
	if !m.Checked() || cmd == nil {
		t.Fatalf("space: checked=%v command=%v, want checked with command", m.Checked(), cmd != nil)
	}
	msg, ok := cmd().(ChangedMsg)
	if !ok || msg.ID != "auto-refresh" || !msg.Checked {
		t.Fatalf("change message = %#v, want auto-refresh checked", msg)
	}
	m, cmd = m.Update(keyPress("enter"))
	if m.Checked() || cmd == nil {
		t.Fatalf("enter: checked=%v command=%v, want unchecked with command", m.Checked(), cmd != nil)
	}
}

func TestToggleSemanticActionsAndInspection(t *testing.T) {
	m := newTestToggle(28)
	next, cmd, handled := m.applyAction(inspect.Invoke(ActionOn))
	if !handled || cmd == nil || !next.Checked() {
		t.Fatalf("on action: handled=%v command=%v checked=%v", handled, cmd != nil, next.Checked())
	}
	next, cmd, handled = next.applyAction(inspect.Invoke(ActionOff))
	if !handled || cmd == nil || next.Checked() {
		t.Fatalf("off action: handled=%v command=%v checked=%v", handled, cmd != nil, next.Checked())
	}
	node := next.Inspect()
	if node.Kind != "toggle" || node.Label != "Auto-refresh" || node.Status != "off" {
		t.Fatalf("unexpected inspection node: %+v", node)
	}
	if node.Attributes["id"] != "auto-refresh" || node.Attributes["checked"] != "false" {
		t.Fatalf("unexpected attributes: %+v", node.Attributes)
	}
	if len(node.Actions) != 5 || node.Actions[3].Enabled == false {
		t.Fatalf("unexpected actions: %+v", node.Actions)
	}
}

func TestToggleDisabled(t *testing.T) {
	m := newTestToggle(28)
	m.SetChecked(true)
	m.SetDisabled(true)
	m.Focus()
	if m.Focused() {
		t.Fatal("disabled toggle accepted focus")
	}
	next, cmd := m.Update(inspect.Invoke(ActionToggle))
	if next.Checked() != m.Checked() || cmd != nil {
		t.Fatalf("disabled toggle changed: checked=%v command=%v", next.Checked(), cmd != nil)
	}
	snaptest.Snap(t, m.View())
}

func TestToggleExactWidth(t *testing.T) {
	for width := 1; width <= 40; width++ {
		m := newTestToggle(width)
		m.SetLabel("A very long setting label")
		m.SetChecked(true)
		if got := lipgloss.Width(m.View()); got != width {
			t.Fatalf("width %d rendered as %d", width, got)
		}
	}

	m := newTestToggle(10)
	m.SetSize(10, 0)
	if got := m.View(); got != "" {
		t.Fatalf("zero-height view = %q, want empty", got)
	}
}

func TestToggleScenarioGolden(t *testing.T) {
	m := scenarioModel{toggle: newTestToggle(28)}
	m.toggle.Focus()
	result := snaptest.RunScenario(m,
		snaptest.ScenarioStep{Name: "turn on", Msg: keyPress("space")},
		snaptest.ScenarioStep{Name: "turn off", Msg: inspect.Invoke(ActionOff)},
	)
	snaptest.SnapScenario(t, result)
}

type scenarioModel struct {
	toggle Model
}

func (m scenarioModel) Init() tea.Cmd { return nil }

func (m scenarioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.toggle, cmd = m.toggle.Update(msg)
	return m, cmd
}

func (m scenarioModel) View() tea.View { return tea.NewView(m.toggle.View()) }
