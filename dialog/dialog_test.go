package dialog

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/layout"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func newTestDialog() Model {
	d := New(tuiweave.Dark())
	d.ID = "quit"
	d.Title = "Quit tuiweave?"
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
	snaptest.SnapCells(t, d.View(), snaptest.WithRoles(tuiweave.Dark()))
}

func TestDialogButtonNavigationGolden(t *testing.T) {
	d := newTestDialog()
	d, _ = d.Update(key("tab")) // select Cancel
	snaptest.SnapCells(t, d.View(), snaptest.WithRoles(tuiweave.Dark()))
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

func TestDialogResultDeliveryScenarioGolden(t *testing.T) {
	m := dialogScenarioModel{dialog: newTestDialog(), open: true}
	result := snaptest.RunScenario(m,
		snaptest.ScenarioStep{Name: "confirm emits result", Msg: key("enter")},
		snaptest.ScenarioStep{Name: "application delivers result", Msg: ResultMsg{ID: "quit", OK: true}},
	)
	snaptest.SnapScenario(t, result)
}

type dialogScenarioModel struct {
	dialog  Model
	open    bool
	outcome string
}

func (m dialogScenarioModel) Init() tea.Cmd { return nil }

func (m dialogScenarioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if result, ok := msg.(ResultMsg); ok {
		m.open = false
		m.outcome = "confirmed " + result.ID
		return m, nil
	}
	if !m.open {
		return m, nil
	}
	var cmd tea.Cmd
	m.dialog, cmd = m.dialog.Update(msg)
	return m, cmd
}

func (m dialogScenarioModel) View() tea.View {
	if !m.open {
		return tea.NewView(m.outcome)
	}
	return tea.NewView(m.dialog.View())
}

func TestDialogTooSmallRendersNothing(t *testing.T) {
	d := newTestDialog()
	d.SetSize(4, 3)
	if got := d.View(); got != "" {
		t.Errorf("tiny dialog View() = %q, want empty", got)
	}
}

func TestDialogIsWidthBounded(t *testing.T) {
	if got := newTestDialog().SizeMode(); got != layout.SizeWidthBounded {
		t.Fatalf("dialog size mode = %d, want SizeWidthBounded", got)
	}
}

func TestDialogContentStaysWithinWidth(t *testing.T) {
	d := newTestDialog()
	d.Title = "界界 dialog"
	d.Body = "説明 with wide content"
	d.ConfirmLabel = "許可する"
	d.CancelLabel = "キャンセル"
	for width := 7; width <= 30; width++ {
		d.SetSize(width, 10)
		for lineNo, line := range strings.Split(d.View(), "\n") {
			if got := lipgloss.Width(line); got != width {
				t.Fatalf("width %d line %d rendered as %d: %q", width, lineNo+1, got, line)
			}
		}
	}
}
