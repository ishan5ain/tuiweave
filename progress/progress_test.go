package progress

import (
	"math"
	"testing"

	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/snaptest"
)

func TestProgressGolden(t *testing.T) {
	m := New(tuiweave.Dark())
	m.SetSize(40, 1)
	m.SetLabel("Indexing")
	m.SetPercent(0.72)
	m.SetStatus(StatusInfo)
	snaptest.Snap(t, m.View())
	snaptest.SnapCells(t, m.View(), snaptest.WithRoles(tuiweave.Dark()))
}

func TestProgressStatusAndClamping(t *testing.T) {
	m := New(tuiweave.Dark())
	m.SetSize(20, 1)

	for _, test := range []struct {
		name   string
		value  float64
		want   float64
		status Status
	}{
		{name: "negative", value: -1, want: 0, status: StatusDanger},
		{name: "success", value: 0.5, want: 0.5, status: StatusSuccess},
		{name: "over", value: 2, want: 1, status: StatusWarning},
		{name: "nan", value: math.NaN(), want: 0, status: StatusNormal},
	} {
		t.Run(test.name, func(t *testing.T) {
			m.SetPercent(test.value)
			m.SetStatus(test.status)
			if got := m.Percent(); got != test.want {
				t.Fatalf("percent = %v, want %v", got, test.want)
			}
			if got := m.Status(); got != test.status {
				t.Fatalf("status = %v, want %v", got, test.status)
			}
			if got := lipgloss.Width(m.View()); got != 20 {
				t.Fatalf("rendered width = %d, want 20", got)
			}
		})
	}
}

func TestProgressExactWidthAndNarrowStates(t *testing.T) {
	for width := 1; width <= 48; width++ {
		m := New(tuiweave.Dark())
		m.SetSize(width, 1)
		m.SetLabel("A long task label")
		m.SetPercent(0.37)
		if got := lipgloss.Width(m.View()); got != width {
			t.Fatalf("width %d rendered as %d", width, got)
		}
	}

	m := New(tuiweave.Dark())
	m.SetSize(10, 0)
	if got := m.View(); got != "" {
		t.Fatalf("zero-height view = %q, want empty", got)
	}
	m.SetSize(0, 1)
	if got := m.View(); got != "" {
		t.Fatalf("zero-width view = %q, want empty", got)
	}
}

func TestProgressInspection(t *testing.T) {
	m := New(tuiweave.Dark())
	m.SetSize(24, 1)
	m.SetLabel("Sync")
	m.SetPercent(0.625)
	m.SetStatus(StatusWarning)
	m.SetShowPercent(false)

	node := m.Inspect()
	if node.Kind != "progress" || node.Label != "Sync" || node.Status != "warning" {
		t.Fatalf("unexpected inspection node: %+v", node)
	}
	if node.Bounds.Width != 24 || node.Bounds.Height != 1 {
		t.Fatalf("unexpected bounds: %+v", node.Bounds)
	}
	if node.Attributes["percent"] != "0.625" || node.Attributes["show_percent"] != "false" {
		t.Fatalf("unexpected attributes: %+v", node.Attributes)
	}
}
