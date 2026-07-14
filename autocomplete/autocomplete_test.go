package autocomplete

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/inspect"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func newTestAutocomplete(width, height int) Model {
	m := New(tuiweave.Dark())
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
	snaptest.SnapCells(t, m.View(), snaptest.WithRoles(tuiweave.Dark()))
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
		m := New(tuiweave.Dark())
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

func TestAutocompleteUnicodeNarrowRowsStayWithinBox(t *testing.T) {
	items := []Item{
		{ID: "wide", Value: "界界界", Label: "界界界", Description: "説明"},
		{ID: "combining", Value: "e\u0301", Label: "e\u0301 cluster", Description: "accent"},
		{ID: "emoji", Value: "👩‍💻", Label: "👩‍💻 developer", Description: "coding"},
		{ID: "plain", Value: "ordinary", Label: "ordinary", Description: "text"},
	}
	for _, width := range []int{0, 1, 2, 3, 4, 8, 16} {
		for _, focused := range []bool{false, true} {
			m := New(tuiweave.Dark())
			m.SetSize(width, 3)
			m.SetItems(items...)
			if focused {
				m.Focus()
				m.Select(2)
			}

			view := m.View()
			if width == 0 {
				if view != "" {
					t.Fatalf("width=0 focused=%v rendered %q", focused, view)
				}
				continue
			}
			for row, line := range strings.Split(view, "\n") {
				if got := ansi.StringWidth(ansi.Strip(line)); got != width {
					t.Fatalf("width=%d focused=%v row=%d rendered width=%d: %q", width, focused, row, got, line)
				}
			}
		}
	}
}

func TestAutocompleteUnicodeTruncationGolden(t *testing.T) {
	m := New(tuiweave.Dark())
	m.SetSize(12, 3)
	m.SetItems(
		Item{ID: "wide", Value: "界e\u0301", Label: "界e\u0301", Description: "説明"},
		Item{ID: "emoji", Value: "👩‍💻 developer", Label: "👩‍💻 developer", Description: "coding"},
		Item{ID: "long", Value: "a very long wide 名称", Label: "a very long wide 名称", Description: "詳細"},
	)
	m.Focus()
	m.Select(1)
	snaptest.Snap(t, m.View())
	snaptest.SnapCells(t, m.View(), snaptest.WithRoles(tuiweave.Dark()))
}

func TestAutocompleteNarrowEmojiDoesNotSplit(t *testing.T) {
	m := New(tuiweave.Dark())
	m.SetSize(3, 2)
	m.SetItems(Item{ID: "emoji", Value: "👩‍💻", Label: "👩‍💻"})
	m.Focus()
	m.Select(0)

	view := ansi.Strip(m.View())
	if strings.Contains(view, "👩") || strings.Contains(view, "💻") {
		t.Fatalf("narrow emoji suggestion was split instead of truncated: %q", view)
	}
	for row, line := range strings.Split(view, "\n") {
		if got := ansi.StringWidth(line); got != 3 {
			t.Fatalf("row %d width = %d, want 3: %q", row, got, line)
		}
	}
}

func TestAutocompleteUnicodeSelectionScroll(t *testing.T) {
	m := New(tuiweave.Dark())
	m.SetSize(14, 2)
	m.SetItems(
		Item{ID: "one", Value: "界 one", Label: "界 one"},
		Item{ID: "two", Value: "e\u0301 two", Label: "e\u0301 two"},
		Item{ID: "three", Value: "👩‍💻 three", Label: "👩‍💻 three"},
		Item{ID: "four", Value: "終 four", Label: "終 four"},
	)
	m.Focus()
	m, _ = m.Update(keyPress("G"))
	if got := m.Selected(); got != 3 {
		t.Fatalf("selected = %d, want 3", got)
	}
	if got := m.YOffset(); got != 2 {
		t.Fatalf("YOffset = %d, want 2", got)
	}
	for row, line := range strings.Split(ansi.Strip(m.View()), "\n") {
		if got := ansi.StringWidth(line); got != 14 {
			t.Fatalf("row %d width = %d, want 14", row, got)
		}
	}
}

func TestAutocompleteUnicodeQueryMatchesByPrefix(t *testing.T) {
	m := New(tuiweave.Dark())
	m.SetItems(
		Item{ID: "wide", Value: "界面", Label: "界面"},
		Item{ID: "combining", Value: "e\u0301clair", Label: "e\u0301clair"},
	)

	m.SetQuery("界")
	if got := m.SelectedID(); got != "wide" {
		t.Fatalf("wide query selected %q, want wide", got)
	}
	m.SetQuery("e\u0301")
	if got := m.SelectedID(); got != "combining" {
		t.Fatalf("combining query selected %q, want combining", got)
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
