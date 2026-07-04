package usagebar

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/snaptest"
)

func TestUsagebarGolden(t *testing.T) {
	u := New(gotui.Dark())
	u.SetSize(60, 1)
	u.SetStats(Stats{
		Model:       "pi-large",
		TokensIn:    12345,
		TokensOut:   987,
		Cost:        0.42,
		ContextUsed: 0.37,
	})
	snaptest.Snap(t, u.View())
	snaptest.SnapCells(t, u.View(), snaptest.WithRoles(gotui.Dark()))
}

func TestContextWarnsByThreshold(t *testing.T) {
	u := New(gotui.Dark())
	u.SetSize(60, 1)

	u.SetStats(Stats{Model: "pi", ContextUsed: 0.85})
	warn := u.View()
	u.SetStats(Stats{Model: "pi", ContextUsed: 0.97})
	danger := u.View()
	if warn == danger {
		t.Error("context styling identical at 85% and 97%")
	}
	if !strings.Contains(ansi.Strip(danger), "ctx 97%") {
		t.Errorf("bar missing ctx figure: %q", ansi.Strip(danger))
	}
}

func TestHumanize(t *testing.T) {
	cases := map[int]string{
		999:       "999",
		12345:     "12.3k",
		1_200_000: "1.2M",
	}
	for n, want := range cases {
		if got := humanize(n); got != want {
			t.Errorf("humanize(%d) = %q, want %q", n, got, want)
		}
	}
}
