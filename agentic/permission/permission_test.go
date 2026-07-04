package permission

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/snaptest"
)

func newTestPrompt() Model {
	p := New(gotui.Dark())
	p.ID = "bash"
	p.Title = `Run "go test ./..."?`
	p.Body = "The agent wants to run a shell command."
	p.SetSize(40, 12)
	return p
}

func key(name string) tea.KeyPressMsg {
	switch name {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	}
	if len(name) == 1 {
		return tea.KeyPressMsg{Code: rune(name[0]), Text: name}
	}
	panic("unsupported key in test helper: " + name)
}

func resultOf(t *testing.T, cmd tea.Cmd) ResultMsg {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected a result command, got nil")
	}
	msg, ok := cmd().(ResultMsg)
	if !ok {
		t.Fatalf("command produced %T, want ResultMsg", cmd())
	}
	return msg
}

func TestPromptGolden(t *testing.T) {
	p := newTestPrompt()
	snaptest.Snap(t, p.View())
	snaptest.SnapCells(t, p.View(), snaptest.WithRoles(gotui.Dark()))
}

func TestNavigateAndConfirm(t *testing.T) {
	p := newTestPrompt()
	p, _ = p.Update(key("j"))
	p, cmd := p.Update(key("enter"))
	res := resultOf(t, cmd)
	if res.Choice != 1 || res.Option != "Allow always" || res.ID != "bash" {
		t.Errorf("ResultMsg = %+v, want Choice=1 Allow always ID=bash", res)
	}
	_ = p
}

func TestNumberQuickSelect(t *testing.T) {
	p := newTestPrompt()
	_, cmd := p.Update(key("3"))
	if res := resultOf(t, cmd); res.Choice != 2 || res.Option != "Deny" {
		t.Errorf("ResultMsg = %+v, want Choice=2 Deny", res)
	}
}

func TestEscPicksSafeDefault(t *testing.T) {
	p := newTestPrompt()
	_, cmd := p.Update(key("esc"))
	if res := resultOf(t, cmd); res.Choice != 2 {
		t.Errorf("esc picked choice %d, want last option (2)", res.Choice)
	}
}

func TestOutOfRangeNumberIgnored(t *testing.T) {
	p := newTestPrompt()
	_, cmd := p.Update(key("9"))
	if cmd != nil {
		t.Error("out-of-range number produced a result")
	}
}

func TestCustomOptions(t *testing.T) {
	p := newTestPrompt()
	p.SetOptions("Yes", "No")
	_, cmd := p.Update(key("esc"))
	if res := resultOf(t, cmd); res.Option != "No" {
		t.Errorf("esc with custom options picked %q, want No", res.Option)
	}
}
