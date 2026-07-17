package tabs

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/inspect"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func newTestTabs(width int) Model {
	m := New(tuiweave.Dark())
	m.SetSize(width, 1)
	m.SetTabs(
		Tab{ID: "overview", Label: "Overview"},
		Tab{ID: "logs", Label: "Logs"},
		Tab{ID: "settings", Label: "Settings"},
	)
	return m
}

func keyPress(key string) tea.KeyPressMsg {
	if len(key) == 1 {
		return tea.KeyPressMsg{Code: rune(key[0]), Text: key}
	}
	return tea.KeyPressMsg{Text: key}
}

func TestTabsFocused(t *testing.T) {
	m := newTestTabs(30)
	m.Focus()
	m.Select(1)
	snaptest.Snap(t, m.View())
	snaptest.SnapCells(t, m.View(), snaptest.WithRoles(tuiweave.Dark()))
}

func TestTabsBlurred(t *testing.T) {
	m := newTestTabs(30)
	m.Select(1)
	snaptest.Snap(t, m.View())
	snaptest.SnapCells(t, m.View(), snaptest.WithRoles(tuiweave.Dark()))
}

func TestTabsOverflowKeepsSelectionVisible(t *testing.T) {
	m := New(tuiweave.Dark())
	m.SetSize(16, 1)
	m.SetTabs(
		Tab{ID: "one", Label: "Overview"},
		Tab{ID: "two", Label: "Logs"},
		Tab{ID: "three", Label: "Settings"},
		Tab{ID: "four", Label: "Deploy"},
	)
	m.Focus()
	m.Select(3)
	snaptest.Snap(t, m.View())
	if got := m.SelectedID(); got != "four" {
		t.Fatalf("SelectedID = %q, want four", got)
	}
}

func TestTabsWidthAndEmptyStates(t *testing.T) {
	for width := 1; width <= 24; width++ {
		m := newTestTabs(width)
		if got := lipgloss.Width(m.View()); got != width {
			t.Fatalf("width %d rendered as %d", width, got)
		}
		empty := New(tuiweave.Dark())
		empty.SetSize(width, 1)
		if got := lipgloss.Width(empty.View()); got != width {
			t.Fatalf("empty width %d rendered as %d", width, got)
		}
	}
	if got := newTestTabs(20).View(); got == "" {
		t.Fatal("configured tabs rendered empty")
	}
}

func TestTabsIndexAtWideStripBoundaries(t *testing.T) {
	m := newTestTabs(30)
	tests := []struct {
		name  string
		x     int
		index int
		ok    bool
	}{
		{name: "before strip", x: -1, index: -1},
		{name: "overview start", x: 0, index: 0, ok: true},
		{name: "overview end", x: 9, index: 0, ok: true},
		{name: "first separator", x: 10, index: -1},
		{name: "logs start", x: 11, index: 1, ok: true},
		{name: "logs end", x: 16, index: 1, ok: true},
		{name: "second separator", x: 17, index: -1},
		{name: "settings start", x: 18, index: 2, ok: true},
		{name: "settings end", x: 27, index: 2, ok: true},
		{name: "trailing padding", x: 28, index: -1},
		{name: "past strip", x: 30, index: -1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			index, ok := m.IndexAt(test.x)
			if index != test.index || ok != test.ok {
				t.Fatalf("IndexAt(%d) = (%d, %v), want (%d, %v)", test.x, index, ok, test.index, test.ok)
			}
		})
	}
}

func TestTabsIndexAtNarrowVisibleWindow(t *testing.T) {
	m := New(tuiweave.Dark())
	m.SetSize(16, 1)
	m.SetTabs(
		Tab{ID: "one", Label: "Overview"},
		Tab{ID: "two", Label: "Logs"},
		Tab{ID: "three", Label: "Settings"},
		Tab{ID: "four", Label: "Deploy"},
	)
	m.Select(3)

	for _, x := range []int{0, 7} {
		if index, ok := m.IndexAt(x); !ok || index != 3 {
			t.Fatalf("IndexAt(%d) = (%d, %v), want visible Deploy index 3", x, index, ok)
		}
	}
	if index, ok := m.IndexAt(8); ok || index != -1 {
		t.Fatalf("IndexAt(8) = (%d, %v), want trailing-padding miss", index, ok)
	}

	m.SetSize(3, 1)
	for x := 0; x < 3; x++ {
		if index, ok := m.IndexAt(x); !ok || index != 3 {
			t.Fatalf("truncated IndexAt(%d) = (%d, %v), want index 3", x, index, ok)
		}
	}
}

func TestTabsIndexAtUsesTerminalCellWidths(t *testing.T) {
	m := New(tuiweave.Dark())
	m.SetSize(12, 1)
	m.SetTabs(
		Tab{ID: "wide", Label: "界"},
		Tab{ID: "combining", Label: "e\u0301"},
	)

	// The wide tab occupies four cells (padding + two-cell glyph + padding),
	// followed by one separator. The combining tab occupies three cells.
	for _, test := range []struct {
		x     int
		index int
		ok    bool
	}{
		{x: 3, index: 0, ok: true},
		{x: 4, index: -1},
		{x: 5, index: 1, ok: true},
		{x: 7, index: 1, ok: true},
		{x: 8, index: -1},
	} {
		index, ok := m.IndexAt(test.x)
		if index != test.index || ok != test.ok {
			t.Errorf("IndexAt(%d) = (%d, %v), want (%d, %v)", test.x, index, ok, test.index, test.ok)
		}
	}
}

func TestTabsIndexAtEmptyAndNonRenderingStates(t *testing.T) {
	empty := New(tuiweave.Dark())
	empty.SetSize(10, 1)
	if index, ok := empty.IndexAt(0); ok || index != -1 {
		t.Fatalf("empty IndexAt(0) = (%d, %v)", index, ok)
	}

	for _, size := range []struct{ width, height int }{{0, 1}, {10, 0}} {
		m := newTestTabs(10)
		m.SetSize(size.width, size.height)
		if index, ok := m.IndexAt(0); ok || index != -1 {
			t.Fatalf("size %dx%d IndexAt(0) = (%d, %v)", size.width, size.height, index, ok)
		}
	}
}

func TestTabsNavigationAndReplacement(t *testing.T) {
	m := newTestTabs(30)
	m.Focus()
	m, _ = m.Update(keyPress("right"))
	m, _ = m.Update(keyPress("l"))
	if got := m.SelectedID(); got != "settings" {
		t.Fatalf("after right+l SelectedID = %q, want settings", got)
	}
	m, _ = m.Update(keyPress("home"))
	if got := m.Selected(); got != 0 {
		t.Fatalf("after home Selected = %d, want 0", got)
	}
	m.SetTabs(
		Tab{ID: "settings", Label: "Preferences"},
		Tab{ID: "overview", Label: "Overview"},
	)
	if got := m.SelectedID(); got != "overview" {
		t.Fatalf("replacement SelectedID = %q, want overview after preserving its ID", got)
	}
	m.SelectID("settings")
	m.SetTabs(
		Tab{ID: "overview", Label: "Overview"},
		Tab{ID: "settings", Label: "Preferences"},
	)
	if got := m.SelectedID(); got != "settings" {
		t.Fatalf("replacement did not preserve selected ID: %q", got)
	}
}

func TestTabsSemanticActionsAndInspection(t *testing.T) {
	m := newTestTabs(30)
	next, cmd := m.Update(inspect.Invoke(ActionSelectPrefix + "settings"))
	if cmd != nil {
		t.Fatal("semantic selection returned a command")
	}
	if got := next.SelectedID(); got != "settings" {
		t.Fatalf("semantic selection = %q, want settings", got)
	}
	node := next.Inspect()
	if node.Kind != "tabs" || node.Selected == nil || node.Selected.Label != "Settings" {
		t.Fatalf("unexpected inspection node: %+v", node)
	}
	if len(node.Actions) != 9 {
		t.Fatalf("action count = %d, want 9", len(node.Actions))
	}
}

func TestTabsScenarioGolden(t *testing.T) {
	m := scenarioModel{tabs: newTestTabs(30)}
	m.tabs.Focus()
	result := snaptest.RunScenario(m,
		snaptest.ScenarioStep{Name: "next tab", Msg: keyPress("right")},
		snaptest.ScenarioStep{Name: "select settings", Msg: inspect.Invoke(ActionSelectPrefix + "settings")},
	)
	snaptest.SnapScenario(t, result)
}

type scenarioModel struct {
	tabs Model
}

func (m scenarioModel) Init() tea.Cmd { return nil }

func (m scenarioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.tabs, cmd = m.tabs.Update(msg)
	return m, cmd
}

func (m scenarioModel) View() tea.View { return tea.NewView(m.tabs.View()) }
