package palette

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

func newTestPalette(width, height int) Model {
	m := New(tuiweave.Dark())
	m.SetItems(
		Item{ID: "open", Label: "Open workspace", Description: "Choose a workspace"},
		Item{ID: "format", Label: "Format document", Description: "Run the formatter"},
		Item{ID: "delete", Label: "Delete workspace", Description: "Destructive action", Disabled: true},
		Item{ID: "settings", Label: "Open settings", Description: "Edit preferences"},
	)
	m.SetSize(width, height)
	return m
}

func keyPress(key string) tea.KeyPressMsg {
	if len(key) == 1 {
		return tea.KeyPressMsg{Code: rune(key[0]), Text: key}
	}
	return tea.KeyPressMsg{Text: key}
}

func TestPaletteDefaultGolden(t *testing.T) {
	m := newTestPalette(44, 5)
	snaptest.Snap(t, m.View())
	snaptest.SnapCells(t, m.View(), snaptest.WithRoles(tuiweave.Dark()))
}

func TestPaletteFocusedFilteredGolden(t *testing.T) {
	m := newTestPalette(44, 5)
	m.Focus()
	m.SetQuery("open")
	snaptest.Snap(t, m.View())
	snaptest.SnapCells(t, m.View(), snaptest.WithRoles(tuiweave.Dark()))
}

func TestPaletteFilteringAndActivation(t *testing.T) {
	m := newTestPalette(44, 5)
	m.Focus()
	m, _ = m.Update(keyPress("f"))
	m, _ = m.Update(keyPress("o"))
	if m.Query() != "fo" || m.FilteredLen() != 1 || m.SelectedID() != "format" {
		t.Fatalf("filtered palette: query=%q count=%d selected=%q", m.Query(), m.FilteredLen(), m.SelectedID())
	}
	m, cmd := m.Update(keyPress("enter"))
	if cmd == nil {
		t.Fatal("enter returned no activation command")
	}
	msg, ok := cmd().(SelectedMsg)
	if !ok || msg.ID != "format" || msg.Label != "Format document" {
		t.Fatalf("activation message = %#v, want format", msg)
	}
	if m.SelectedID() != "format" {
		t.Fatalf("selection changed during activation: %q", m.SelectedID())
	}
}

func TestPaletteSkipsDisabledAndSemanticSelection(t *testing.T) {
	m := newTestPalette(44, 3)
	m.Focus()
	m, _ = m.Update(keyPress("down"))
	if m.SelectedID() != "format" {
		t.Fatalf("first down selected %q, want format", m.SelectedID())
	}
	m, _ = m.Update(keyPress("down"))
	if m.SelectedID() != "settings" {
		t.Fatalf("second down selected %q, want settings", m.SelectedID())
	}
	next, cmd, handled := m.applyAction(inspect.Invoke(ActionSelectPrefix + "open"))
	if !handled || cmd != nil || next.SelectedID() != "open" {
		t.Fatalf("semantic selection: handled=%v command=%v selected=%q", handled, cmd != nil, next.SelectedID())
	}
	next, cmd, handled = next.applyAction(inspect.Invoke(ActionSelectPrefix + "delete"))
	if handled || cmd != nil || next.SelectedID() != "open" {
		t.Fatalf("disabled semantic selection: handled=%v command=%v selected=%q", handled, cmd != nil, next.SelectedID())
	}
}

func TestPaletteInspection(t *testing.T) {
	m := newTestPalette(44, 5)
	m.Focus()
	m.SetQuery("settings")
	node := m.Inspect()
	if node.Kind != "palette" || node.Selected == nil || node.Selected.Label != "Open settings" {
		t.Fatalf("unexpected inspection node: %+v", node)
	}
	if node.Scroll == nil || node.Scroll.Total != 1 || node.Scroll.Visible != 4 {
		t.Fatalf("unexpected scroll metadata: %+v", node.Scroll)
	}
	if node.Attributes["query"] != "settings" || node.Attributes["filtered_count"] != "1" {
		t.Fatalf("unexpected attributes: %+v", node.Attributes)
	}
	if len(node.Actions) != 12 {
		t.Fatalf("action count = %d, want 12", len(node.Actions))
	}
}

func TestPaletteEmptyStateAndExactWidth(t *testing.T) {
	for width := 1; width <= 52; width++ {
		m := newTestPalette(width, 5)
		m.SetQuery("no-match")
		for _, line := range splitLines(m.View()) {
			if got := lipgloss.Width(line); got != width {
				t.Fatalf("width %d rendered line as %d: %q", width, got, line)
			}
		}
	}
	m := newTestPalette(20, 0)
	if got := m.View(); got != "" {
		t.Fatalf("zero-height view = %q, want empty", got)
	}
}

func TestPaletteUnicodeNarrowRowsStayWithinBox(t *testing.T) {
	items := []Item{
		{ID: "wide", Label: "界界界", Description: "説明"},
		{ID: "combining", Label: "e\u0301 cluster", Description: "accent"},
		{ID: "emoji", Label: "👩‍💻 developer", Description: "coding"},
		{ID: "plain", Label: "ordinary", Description: "text"},
	}
	for _, width := range []int{0, 1, 2, 3, 4, 8, 16} {
		for _, focused := range []bool{false, true} {
			m := New(tuiweave.Dark())
			m.SetSize(width, 4)
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

func TestPaletteUnicodeTruncationGolden(t *testing.T) {
	m := New(tuiweave.Dark())
	m.SetSize(12, 4)
	m.SetItems(
		Item{ID: "wide", Label: "界e\u0301", Description: "説明"},
		Item{ID: "emoji", Label: "👩‍💻 developer", Description: "coding"},
		Item{ID: "long", Label: "a very long wide 名称", Description: "詳細"},
	)
	m.Focus()
	m.Select(1)
	snaptest.Snap(t, m.View())
	snaptest.SnapCells(t, m.View(), snaptest.WithRoles(tuiweave.Dark()))
}

func TestPaletteNarrowEmojiDoesNotSplit(t *testing.T) {
	m := New(tuiweave.Dark())
	m.SetSize(3, 2)
	m.SetItems(Item{ID: "emoji", Label: "👩‍💻"})
	m.Focus()
	m.Select(0)

	view := ansi.Strip(m.View())
	if strings.Contains(view, "👩") || strings.Contains(view, "💻") {
		t.Fatalf("narrow emoji item was split instead of truncated: %q", view)
	}
	for row, line := range strings.Split(view, "\n") {
		if got := ansi.StringWidth(line); got != 3 {
			t.Fatalf("row %d width = %d, want 3: %q", row, got, line)
		}
	}
}

func TestPaletteUnicodeSelectionScroll(t *testing.T) {
	m := New(tuiweave.Dark())
	m.SetSize(14, 3) // query plus two visible result rows
	m.SetItems(
		Item{ID: "one", Label: "界 one"},
		Item{ID: "two", Label: "e\u0301 two"},
		Item{ID: "three", Label: "👩‍💻 three"},
		Item{ID: "four", Label: "終 four"},
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

func TestPaletteScenarioGolden(t *testing.T) {
	m := newTestPalette(44, 5)
	m.Focus()
	result := snaptest.RunScenario(scenarioModel{palette: m},
		snaptest.ScenarioStep{Name: "search format", Msg: keyPress("f")},
		snaptest.ScenarioStep{Name: "search formatter", Msg: keyPress("o")},
		snaptest.ScenarioStep{Name: "activate", Msg: keyPress("enter")},
	)
	snaptest.SnapScenario(t, result)
}

func splitLines(value string) []string {
	var lines []string
	start := 0
	for i, r := range value {
		if r == '\n' {
			lines = append(lines, value[start:i])
			start = i + 1
		}
	}
	return append(lines, value[start:])
}

type scenarioModel struct {
	palette Model
}

func (m scenarioModel) Init() tea.Cmd { return nil }

func (m scenarioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.palette, cmd = m.palette.Update(msg)
	return m, cmd
}

func (m scenarioModel) View() tea.View { return tea.NewView(m.palette.View()) }
