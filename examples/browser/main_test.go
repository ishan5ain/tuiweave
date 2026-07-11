package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/charmbracelet/x/ansi"
	"github.com/ishan5ain/tuiweave/inspect"
	"github.com/ishan5ain/tuiweave/mouse"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func sized(t *testing.T, width, height int) model {
	t.Helper()
	m := newModel()
	next, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return next.(model)
}

func key(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func update(t *testing.T, m model, msg tea.Msg) (model, tea.Cmd) {
	t.Helper()
	next, cmd := m.Update(msg)
	return next.(model), cmd
}

func TestBrowserScreenGolden(t *testing.T) {
	m := sized(t, 88, 24)
	snaptest.Snap(t, m.render())
}

func TestBrowserFilterAndPreview(t *testing.T) {
	m := sized(t, 88, 24)
	m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	for _, r := range "README" {
		m, _ = update(t, m, key(r))
	}
	if m.filter.Value() != "README" || m.files.FilteredLen() != 0 {
		t.Fatalf("filter = %q, matches = %d; want README and no workspace match", m.filter.Value(), m.files.FilteredLen())
	}

	m, _ = update(t, m, tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl})
	for _, r := range "main" {
		m, _ = update(t, m, key(r))
	}
	if m.files.FilteredLen() == 0 || m.files.SelectedItem() != "cmd/tuiweave/main.go" {
		t.Fatalf("filtered selection = %q, matches = %d; want cmd/tuiweave/main.go", m.files.SelectedItem(), m.files.FilteredLen())
	}
	if !strings.Contains(ansi.Strip(m.preview.View()), "loadConfig") {
		t.Fatalf("preview does not follow filtered selection: %q", ansi.Strip(m.preview.View()))
	}
}

func TestBrowserFocusAndTabComposition(t *testing.T) {
	m := sized(t, 88, 24)
	if m.fm.Index() != 0 || !m.tabs.Focused() {
		t.Fatalf("initial focus index=%d tabs=%v", m.fm.Index(), m.tabs.Focused())
	}
	for i := 1; i <= 3; i++ {
		m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
		if m.fm.Index() != i {
			t.Fatalf("focus step %d = %d", i, m.fm.Index())
		}
	}
	if !m.preview.Focused() || m.files.Focused() {
		t.Fatalf("focus components: preview=%v files=%v", m.preview.Focused(), m.files.Focused())
	}
	m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	if m.fm.Index() != 0 || !m.tabs.Focused() {
		t.Fatalf("focus did not wrap: index=%d tabs=%v", m.fm.Index(), m.tabs.Focused())
	}
}

func TestBrowserWheelRoutesToPreviewWhenBlurred(t *testing.T) {
	m := sized(t, 88, 24)
	if m.preview.Focused() {
		t.Fatal("preview should start blurred for this routing check")
	}

	m, _ = update(t, m, tea.MouseWheelMsg{
		X:      70,
		Y:      10,
		Button: tea.MouseWheelDown,
	})
	if got := m.preview.YOffset(); got != mouse.WheelLines {
		t.Fatalf("preview offset after blurred wheel = %d, want %d", got, mouse.WheelLines)
	}
}

func TestBrowserRecentTabGolden(t *testing.T) {
	m := sized(t, 88, 24)
	m, _ = update(t, m, tea.KeyPressMsg{Code: tea.KeyRight})
	if m.tabs.SelectedID() != "recent" || m.files.SelectedItem() != "README.md" {
		t.Fatalf("recent tab: tab=%q selected=%q", m.tabs.SelectedID(), m.files.SelectedItem())
	}
	snaptest.Snap(t, m.render())
}

func TestBrowserNarrowGolden(t *testing.T) {
	m := sized(t, 48, 14)
	snaptest.Snap(t, m.render())
}

func TestBrowserInspectionGolden(t *testing.T) {
	m := sized(t, 88, 24)
	data, err := inspect.Marshal(m.Inspect())
	if err != nil {
		t.Fatal(err)
	}
	snaptest.Snap(t, string(data))
}

func TestBrowserScenarioGolden(t *testing.T) {
	result := snaptest.RunScenario(sized(t, 88, 24),
		snaptest.ScenarioStep{Name: "focus filter", Msg: tea.KeyPressMsg{Code: tea.KeyTab}},
		snaptest.ScenarioStep{Name: "filter main", Msg: key('m')},
		snaptest.ScenarioStep{Name: "wheel preview while filter focused", Msg: tea.MouseWheelMsg{X: 70, Y: 10, Button: tea.MouseWheelDown}},
		snaptest.ScenarioStep{Name: "focus files", Msg: tea.KeyPressMsg{Code: tea.KeyTab}},
		snaptest.ScenarioStep{Name: "focus preview", Msg: tea.KeyPressMsg{Code: tea.KeyTab}},
		snaptest.ScenarioStep{Name: "scroll preview", Msg: key('f')},
	)
	snaptest.SnapScenario(t, result)
}
