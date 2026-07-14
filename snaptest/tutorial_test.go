package snaptest_test

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/snaptest"
	"github.com/ishan5ain/tuiweave/toggle"
)

const tutorialToggleID = "auto-refresh"

type tutorialModel struct {
	toggle toggle.Model
	notice string
}

func newTutorialModel() tutorialModel {
	setting := toggle.New(tuiweave.Dark())
	setting.ID = tutorialToggleID
	setting.SetLabel("Auto-refresh")
	setting.SetSize(24, 1)
	setting.Focus()
	return tutorialModel{
		toggle: setting,
		notice: "No change delivered",
	}
}

func (m tutorialModel) Init() tea.Cmd { return nil }

func (m tutorialModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if changed, ok := msg.(toggle.ChangedMsg); ok {
		m.notice = fmt.Sprintf("Delivered %s=%t", changed.ID, changed.Checked)
		return m, nil
	}

	var cmd tea.Cmd
	m.toggle, cmd = m.toggle.Update(msg)
	return m, cmd
}

func (m tutorialModel) View() tea.View {
	return tea.NewView(m.toggle.View() + "\n" + m.notice)
}

func TestTutorialInitialView(t *testing.T) {
	m := newTutorialModel()

	snaptest.Snap(t, m.View().Content)
	snaptest.SnapCells(t, m.View().Content,
		snaptest.WithRoles(tuiweave.Dark()))
}

func TestTutorialCommandDelivery(t *testing.T) {
	m := newTutorialModel()

	next, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("activation did not emit a command")
	}
	m = next.(tutorialModel)
	if !m.toggle.Checked() {
		t.Fatal("toggle is unchecked after activation")
	}
	if m.notice != "No change delivered" {
		t.Fatalf("notice changed before command delivery: %q", m.notice)
	}

	result := cmd()
	changed, ok := result.(toggle.ChangedMsg)
	if !ok {
		t.Fatalf("command result has type %T, want toggle.ChangedMsg", result)
	}
	if changed.ID != tutorialToggleID || !changed.Checked {
		t.Fatalf("command result = %+v, want ID %q checked", changed, tutorialToggleID)
	}

	next, cmd = m.Update(changed)
	if cmd != nil {
		t.Fatal("delivering ChangedMsg emitted an unexpected command")
	}
	m = next.(tutorialModel)
	if got, want := m.notice, "Delivered auto-refresh=true"; got != want {
		t.Fatalf("notice = %q, want %q", got, want)
	}
}

func TestTutorialScenario(t *testing.T) {
	result := snaptest.RunScenario(newTutorialModel(),
		snaptest.ScenarioStep{
			Name: "activate toggle",
			Msg:  tea.KeyPressMsg{Code: tea.KeyEnter},
		},
		snaptest.ScenarioStep{
			Name: "deliver changed result",
			Msg:  toggle.ChangedMsg{ID: tutorialToggleID, Checked: true},
		},
	)
	snaptest.SnapScenario(t, result)

	final := result.Model.(tutorialModel)
	if !final.toggle.Checked() {
		t.Fatal("final toggle is unchecked")
	}
	if got, want := final.notice, "Delivered auto-refresh=true"; got != want {
		t.Fatalf("final notice = %q, want %q", got, want)
	}
}
