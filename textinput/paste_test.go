package textinput

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/inspect"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func TestPaste(t *testing.T) {
	for _, tc := range []struct{ name, content, want string }{
		{"plain", "hello", "hello"},
		{"unicode", "界e\u0301👩‍💻", "界e\u0301👩‍💻"},
		{"multiline", "a\r\nb\tc", "a b c"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newFocused(12)
			m.SetValue("[]")
			m, _ = m.Update(key("left"))
			var cmd tea.Cmd
			m, cmd = m.Update(tea.PasteMsg{Content: tc.content})
			if cmd != nil {
				t.Fatal("paste produced a command")
			}
			if got := m.Value(); got != "["+tc.want+"]" {
				t.Fatalf("value = %q, want %q", got, "["+tc.want+"]")
			}
			if m.pos != 1+len([]rune(tc.want)) {
				t.Fatalf("cursor = %d", m.pos)
			}
			view := ansi.Strip(m.View())
			if strings.ContainsAny(view, "\r\n\t") || ansi.StringWidth(view) > 12 {
				t.Fatalf("unbounded view %q", view)
			}
			m.Blur()
			before := m.Value()
			m, cmd = m.Update(tea.PasteMsg{Content: "ignored"})
			if m.Value() != before || cmd != nil {
				t.Fatal("blurred paste changed input")
			}
			for _, msg := range []tea.Msg{tea.PasteStartMsg{}, tea.PasteEndMsg{}} {
				m, cmd = m.Update(msg)
				if m.Value() != before || cmd != nil {
					t.Fatal("paste marker changed input")
				}
			}
		})
	}
}

func TestPasteGolden(t *testing.T) {
	for _, width := range []int{18, 5, 3} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			m := newFocused(width)
			m, _ = m.Update(tea.PasteMsg{Content: "界e\u0301\r\n👩‍💻"})
			snaptest.Snap(t, m.View())
			snaptest.SnapCells(t, m.View(), snaptest.WithRoles(tuiweave.Dark()))
		})
	}
}

func TestPasteScenario(t *testing.T) {
	m := New(tuiweave.Dark())
	m.SetSize(12, 1)
	result := snaptest.RunScenario(textinputScenarioModel{input: m},
		snaptest.ScenarioStep{Name: "blurred paste ignored", Msg: tea.PasteMsg{Content: "ignored"}},
		snaptest.ScenarioStep{Name: "focus", Msg: inspect.Invoke(ActionFocus)},
		snaptest.ScenarioStep{Name: "paste starts", Msg: tea.PasteStartMsg{}},
		snaptest.ScenarioStep{Name: "paste unicode and multiline", Msg: tea.PasteMsg{Content: "界e\u0301\r\nx"}},
		snaptest.ScenarioStep{Name: "paste ends", Msg: tea.PasteEndMsg{}},
		snaptest.ScenarioStep{Name: "resize narrow", Msg: tea.WindowSizeMsg{Width: 5, Height: 1}},
	)
	snaptest.SnapScenario(t, result)
}

func TestPasteNoopPreservesCursor(t *testing.T) {
	m := newFocused(12)
	m.SetValue("ab")
	m, _ = m.Update(key("left"))
	for _, msg := range []tea.Msg{tea.PasteMsg{}, tea.PasteMsg{Content: "\x1b[31m\x00\x7f"}, tea.PasteStartMsg{}, tea.PasteEndMsg{}} {
		next, cmd := m.Update(msg)
		if cmd != nil || next.Value() != "ab" || next.pos != 1 {
			t.Fatal("no-op changed input or cursor")
		}
	}
}
