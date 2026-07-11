package autocomplete

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/inspect"
	"github.com/ishansain/gotui/snaptest"
)

func newTestAutocomplete(width, height int) Model {
	m := New(gotui.Dark())
	m.SetItems(
		Item{ID: "git-checkout", Value: "git checkout ", Label: "git checkout", Description: "switch branch"},
		Item{ID: "git-status", Value: "git status", Label: "git status", Description: "show changes"},
		Item{ID: "go-test", Value: "go test ./...", Label: "go test", Description: "run tests", Disabled: true},
		Item{ID: "go-run", Value: "go run ./cmd", Label: "go run", Description: "run a command"},
	)
	m.SetSize(width, height)
	m.SetQuery("g")
	m.Focus()
	return m
}

func keyPress(key string) tea.KeyPressMsg {
	if len(key) == 1 {
		return tea.KeyPressMsg{Code: rune(key[0]), Text: key}
	}
	return tea.KeyPressMsg{Text: key}
}

func TestAutocompleteGolden(t *testing.T) {
	m := newTestAutocomplete(38, 3)
	snaptest.Snap(t, m.View())
	snaptest.SnapCells(t, m.View(), snaptest.WithRoles(gotui.Dark()))
}

func TestAutocompleteNavigationAndActivation(t *testing.T) {
	m := newTestAutocomplete(38, 3)
	m, _ = m.Update(keyPress("down"))
	if got := m.SelectedID(); got != "git-status" {
		t.Fatalf("selected after down = %q, want git-status", got)
	}
	m, cmd := m.Update(keyPress("enter"))
	if cmd == nil {
		t.Fatal("enter returned nil command")
	}
	msg, ok := cmd().(SelectedMsg)
	if !ok || msg.ID != "git-status" || msg.Value != "git status" {
		t.Fatalf("activation message = %#v, want git-status/git status", msg)
	}
	if m.SelectedID() != "git-status" {
		t.Fatalf("activation changed selection to %q", m.SelectedID())
	}
}

func TestAutocompleteSemanticActionsAndInspection(t *testing.T) {
	m := newTestAutocomplete(38, 3)
	next, cmd, handled := m.applyAction(inspect.Invoke(ActionSelectPrefix + "go-run"))
	if !handled || cmd != nil || next.SelectedID() != "go-run" {
		t.Fatalf("semantic selection: handled=%v command=%v selected=%q", handled, cmd != nil, next.SelectedID())
	}
	node := next.Inspect()
	if node.Kind != "autocomplete" || node.Selected == nil || node.Selected.Label != "go run" {
		t.Fatalf("unexpected inspection node: %+v", node)
	}
	if node.Attributes["query"] != "g" || node.Attributes["filtered_count"] != "4" {
		t.Fatalf("unexpected attributes: %+v", node.Attributes)
	}
	if len(node.Actions) != 12 {
		t.Fatalf("action count = %d, want 12", len(node.Actions))
	}
	for _, action := range node.Actions {
		if action.ID == ActionSelectPrefix+"go-test" && action.Enabled {
			t.Fatal("disabled suggestion exposed as enabled")
		}
	}
}

func TestAutocompleteFilteringAndEmptyState(t *testing.T) {
	m := newTestAutocomplete(30, 2)
	m.SetQuery("xyz")
	if m.FilteredLen() != 0 || m.Selected() != -1 || m.View() != "" {
		t.Fatalf("empty match state: count=%d selected=%d view=%q", m.FilteredLen(), m.Selected(), m.View())
	}
	m.SetQuery("go")
	if m.FilteredLen() != 2 || m.SelectedID() != "go-run" {
		t.Fatalf("go match state: count=%d selected=%q", m.FilteredLen(), m.SelectedID())
	}
}

func TestAutocompleteWideAndNarrowContentStaysWithinBox(t *testing.T) {
	for width := 1; width <= 40; width++ {
		m := New(gotui.Dark())
		m.SetItems(
			Item{ID: "wide", Value: "界界", Label: "界界 e\u0301", Description: "説明"},
			Item{ID: "plain", Value: "plain", Label: "plain", Description: "text"},
		)
		m.SetSize(width, 3)
		m.Focus()
		for i, line := range strings.Split(m.View(), "\n") {
			if got := lipgloss.Width(line); got != width {
				t.Fatalf("width %d line %d rendered as %d: %q", width, i+1, got, line)
			}
		}
	}
}

func TestAutocompleteScenarioGolden(t *testing.T) {
	m := newTestAutocomplete(38, 3)
	result := snaptest.RunScenario(scenarioModel{autocomplete: m},
		snaptest.ScenarioStep{Name: "next suggestion", Msg: keyPress("down")},
		snaptest.ScenarioStep{Name: "accept suggestion", Msg: keyPress("enter")},
	)
	snaptest.SnapScenario(t, result)
}

type scenarioModel struct {
	autocomplete Model
}

func (m scenarioModel) Init() tea.Cmd { return nil }

func (m scenarioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.autocomplete, cmd = m.autocomplete.Update(msg)
	return m, cmd
}

func (m scenarioModel) View() tea.View { return tea.NewView(m.autocomplete.View()) }
