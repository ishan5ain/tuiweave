package permission

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/inspect"
	"github.com/ishan5ain/tuiweave/layout"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func newTestPrompt() Model {
	p := New(tuiweave.Dark())
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
	snaptest.SnapCells(t, p.View(), snaptest.WithRoles(tuiweave.Dark()))
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

func TestPermissionResultDeliveryScenarioGolden(t *testing.T) {
	m := permissionScenarioModel{prompt: newTestPrompt(), open: true}
	result := snaptest.RunScenario(m,
		snaptest.ScenarioStep{Name: "select allow always", Msg: key("down")},
		snaptest.ScenarioStep{Name: "confirm emits result", Msg: key("enter")},
		snaptest.ScenarioStep{
			Name: "application delivers result",
			Msg:  ResultMsg{ID: "bash", Choice: 1, Option: "Allow always"},
		},
	)
	snaptest.SnapScenario(t, result)
}

type permissionScenarioModel struct {
	prompt  Model
	open    bool
	outcome string
}

func (m permissionScenarioModel) Init() tea.Cmd { return nil }

func (m permissionScenarioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if result, ok := msg.(ResultMsg); ok {
		m.open = false
		m.outcome = result.ID + ": " + result.Option
		return m, nil
	}
	if !m.open {
		return m, nil
	}
	var cmd tea.Cmd
	m.prompt, cmd = m.prompt.Update(msg)
	return m, cmd
}

func (m permissionScenarioModel) View() tea.View {
	if !m.open {
		return tea.NewView(m.outcome)
	}
	return tea.NewView(m.prompt.View())
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

func TestProvenanceAndSemanticChoice(t *testing.T) {
	p := newTestPrompt()
	p.SetProvenance(Provenance{
		Tool:          "Bash",
		Operation:     "execute",
		Target:        "repo",
		Scope:         "workspace",
		Detail:        "go test ./...",
		Impact:        "runs tests",
		Reversibility: "reversible",
		Policy:        "shell approval",
	})
	if got := p.Provenance(); got.Detail != "go test ./..." || got.Policy != "shell approval" {
		t.Fatalf("provenance = %+v", got)
	}
	node := p.Inspect()
	if node.Status != "awaiting_approval" || node.Attributes["tool"] != "Bash" {
		t.Fatalf("inspection = %+v", node)
	}
	if len(node.Actions) != 3 || node.Actions[1].ID != "choose.2" {
		t.Fatalf("inspection actions = %+v", node.Actions)
	}
	_, cmd := p.Update(inspect.Invoke("choose.2"))
	if res := resultOf(t, cmd); res.Choice != 1 || res.Option != "Allow always" {
		t.Fatalf("semantic choice result = %+v", res)
	}
}

func TestPermissionIsWidthBounded(t *testing.T) {
	if got := newTestPrompt().SizeMode(); got != layout.SizeWidthBounded {
		t.Fatalf("permission size mode = %d, want SizeWidthBounded", got)
	}
}

func TestPermissionContentStaysWithinWidth(t *testing.T) {
	p := newTestPrompt()
	p.Title = "界界 permission"
	p.Body = "説明 with wide content"
	p.SetOptions("許可する", "常に許可", "キャンセル")
	for width := 7; width <= 30; width++ {
		p.SetSize(width, 12)
		for lineNo, line := range strings.Split(p.View(), "\n") {
			if got := lipgloss.Width(line); got != width {
				t.Fatalf("width %d line %d rendered as %d: %q", width, lineNo+1, got, line)
			}
		}
	}
}
