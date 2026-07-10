package snaptest

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNormalize(t *testing.T) {
	in := "header   \nbody\t\nfooter"
	want := "header\nbody\t\nfooter\n"
	if got := normalize(in); got != want {
		t.Errorf("normalize(%q) = %q, want %q", in, got, want)
	}
}

func TestGoldenPathSanitizesSubtests(t *testing.T) {
	t.Run("with spaces/and slashes", func(t *testing.T) {
		got := goldenPath(t, ".golden")
		if strings.ContainsAny(got[len("testdata/"):], "/ ") {
			t.Errorf("goldenPath produced unsanitized name: %q", got)
		}
	})
}

func TestDiffLinesReportsMismatches(t *testing.T) {
	out := diffLines("a\nb\nc\n", "a\nX\nc\nextra\n")
	for _, wantFragment := range []string{"line 2:", `want: "b"`, `got:  "X"`, "line 4:", "<missing line>"} {
		if !strings.Contains(out, wantFragment) {
			t.Errorf("diff output missing %q in:\n%s", wantFragment, out)
		}
	}
	if strings.Contains(out, "line 1:") || strings.Contains(out, "line 3:") {
		t.Errorf("diff output reports matching lines:\n%s", out)
	}
}

// TestSnapGolden exercises the real golden workflow end to end; its files in
// testdata/ are generated with -update and checked in.
func TestSnapGolden(t *testing.T) {
	view := "┌ demo ─┐   \n│ hello │\n└───────┘"
	Snap(t, view)
}

func TestSnapStyledGolden(t *testing.T) {
	// Hand-written ANSI (bold red "hi") keeps this deterministic across
	// environments, independent of any color-profile detection.
	view := "\x1b[1;31mhi\x1b[0m there"
	SnapStyled(t, view)
}

type scenarioKeyModel struct {
	text     string
	executed *bool
}

func (m scenarioKeyModel) Init() tea.Cmd { return nil }

func (m scenarioKeyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	m.text = key.String()
	if key.String() == "enter" {
		return m, func() tea.Msg {
			*m.executed = true
			return nil
		}
	}
	return m, nil
}

func (m scenarioKeyModel) View() tea.View { return tea.NewView(m.text) }

func TestRunScenarioCapturesFramesWithoutRunningCommands(t *testing.T) {
	executed := false
	result := RunScenario(scenarioKeyModel{text: "ready", executed: &executed},
		ScenarioStep{
			Name: "type x",
			Msg:  tea.KeyPressMsg{Code: 'x', Text: "x"},
		},
		ScenarioStep{
			Name: "submit",
			Msg:  tea.KeyPressMsg{Code: tea.KeyEnter},
		},
	)

	if len(result.Frames) != 3 {
		t.Fatalf("captured %d frames, want 3", len(result.Frames))
	}
	if result.Frames[0].Name != "initial" || result.Frames[0].HasCommand {
		t.Fatalf("initial frame = %+v", result.Frames[0])
	}
	if result.Frames[1].View != "x" || result.Frames[1].HasCommand {
		t.Fatalf("type frame = %+v", result.Frames[1])
	}
	if !result.Frames[2].HasCommand {
		t.Fatal("submit frame did not record emitted command")
	}
	if executed {
		t.Fatal("RunScenario executed a command; command execution must be explicit")
	}
	final, ok := result.Model.(scenarioKeyModel)
	if !ok || final.text != "enter" {
		t.Fatalf("final model = %#v, want text enter", result.Model)
	}
}

func TestSnapScenarioGolden(t *testing.T) {
	result := RunScenario(scenarioKeyModel{text: "ready", executed: new(bool)},
		ScenarioStep{Name: "type x", Msg: tea.KeyPressMsg{Code: 'x', Text: "x"}},
		ScenarioStep{Name: "submit", Msg: tea.KeyPressMsg{Code: tea.KeyEnter}},
	)
	SnapScenario(t, result)
}
