package dialog

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/snaptest"
)

func newTestDialog() Model {
	d := New(gotui.Dark())
	d.ID = "quit"
	d.Title = "Quit gotui?"
	d.Body = "Unsaved changes will be lost."
	d.SetSize(36, 10)
	return d
}

func key(name string) tea.KeyPressMsg {
	switch name {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "left":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	}
	panic("unsupported key in test helper: " + name)
}

func TestDialogGolden(t *testing.T) {
	d := newTestDialog()
	snaptest.Snap(t, d.View())
	snaptest.SnapCells(t, d.View(), snaptest.WithRoles(gotui.Dark()))
}

func TestDialogButtonNavigationGolden(t *testing.T) {
	d := newTestDialog()
	d, _ = d.Update(key("tab")) // select Cancel
	snaptest.SnapCells(t, d.View(), snaptest.WithRoles(gotui.Dark()))
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

func TestDialogConfirm(t *testing.T) {
	d := newTestDialog()
	d, cmd := d.Update(key("enter"))
	res := resultOf(t, cmd)
	if !res.OK || res.ID != "quit" {
		t.Errorf("ResultMsg = %+v, want OK=true ID=quit", res)
	}
	_ = d
}

func TestDialogCancelViaButtonsAndEsc(t *testing.T) {
	d := newTestDialog()
	d, _ = d.Update(key("left")) // switch to Cancel
	d, cmd := d.Update(key("enter"))
	if res := resultOf(t, cmd); res.OK {
		t.Error("cancel button produced OK=true")
	}

	d2 := newTestDialog()
	_, cmd = d2.Update(key("esc"))
	if res := resultOf(t, cmd); res.OK {
		t.Error("esc produced OK=true")
	}
	_ = d
}

func TestDialogTooSmallRendersNothing(t *testing.T) {
	d := newTestDialog()
	d.SetSize(4, 3)
	if got := d.View(); got != "" {
		t.Errorf("tiny dialog View() = %q, want empty", got)
	}
}
