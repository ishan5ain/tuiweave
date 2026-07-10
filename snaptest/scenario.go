package snaptest

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// ScenarioStep is one explicit message in an interaction scenario.
//
// Name is used as the checkpoint heading in a scenario golden. Msg is delivered
// directly to the model; commands returned by Update are recorded but not
// executed. Keeping command execution outside the harness makes timers, I/O,
// and command batching an explicit application-test decision.
type ScenarioStep struct {
	Name string
	Msg  tea.Msg
}

// ScenarioFrame is one rendered checkpoint from a scenario.
type ScenarioFrame struct {
	// Name identifies the initial state or the message that produced the frame.
	Name string
	// View is the styled tea.View.Content captured at the checkpoint.
	View string
	// HasCommand reports whether the corresponding Update returned a command.
	// The initial frame always has HasCommand=false.
	HasCommand bool
}

// ScenarioResult contains the final model and every rendered checkpoint.
type ScenarioResult struct {
	Model  tea.Model
	Frames []ScenarioFrame
}

// RunScenario applies explicit messages to model and captures a frame before
// the first message and after every step. It does not call Init or execute
// commands; callers can model those behaviors with explicit scenario steps.
func RunScenario(model tea.Model, steps ...ScenarioStep) ScenarioResult {
	result := ScenarioResult{
		Model: model,
		Frames: []ScenarioFrame{{
			Name: "initial",
			View: model.View().Content,
		}},
	}

	for i, step := range steps {
		name := step.Name
		if name == "" {
			name = fmt.Sprintf("step-%d", i+1)
		}

		next, cmd := result.Model.Update(step.Msg)
		if next == nil {
			panic(fmt.Sprintf("snaptest: scenario step %q returned a nil model", name))
		}
		result.Model = next
		result.Frames = append(result.Frames, ScenarioFrame{
			Name:       name,
			View:       next.View().Content,
			HasCommand: cmd != nil,
		})
	}

	return result
}

// SnapScenario snapshots the plain-text rendering and command-emission status
// of every frame to testdata/<TestName>.scenario.golden.
//
// Scenario goldens are intentionally separate from view goldens: they describe
// an interaction sequence and are not a replacement for a focused component
// snapshot. Use RunScenario's final Model for state assertions that should stay
// semantic rather than textual.
func SnapScenario(t *testing.T, result ScenarioResult) {
	t.Helper()
	compare(t, goldenPath(t, ".scenario.golden"), formatScenario(result))
}

func formatScenario(result ScenarioResult) string {
	var b strings.Builder
	for i, frame := range result.Frames {
		if i > 0 {
			b.WriteString("\n")
		}

		fmt.Fprintf(&b, "## %s\ncommand: %s\n", frame.Name, yesNo(frame.HasCommand))
		b.WriteString(normalize(ansi.Strip(frame.View)))
	}
	return ensureTrailingNewline(b.String())
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
