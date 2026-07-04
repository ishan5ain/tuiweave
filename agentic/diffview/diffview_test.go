package diffview

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/snaptest"
)

const sample = `diff --git a/main.go b/main.go
index 1111111..2222222 100644
--- a/main.go
+++ b/main.go
@@ -1,4 +1,4 @@
 func main() {
-	fmt.Println("old")
+	fmt.Println("new")
 }
`

func TestSprintGolden(t *testing.T) {
	out := Sprint(gotui.Dark(), sample, 40)
	snaptest.Snap(t, out)
	snaptest.SnapCells(t, out, snaptest.WithRoles(gotui.Dark()))
}

func TestModelScrolls(t *testing.T) {
	m := New(gotui.Dark())
	m.SetSize(40, 4)
	m.SetDiff(sample)
	m.Focus()

	first := m.View()
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	if m.View() == first {
		t.Error("diff pane did not scroll on j")
	}
	if got := len(strings.Split(m.View(), "\n")); got != 4 {
		t.Errorf("rendered %d rows, want 4", got)
	}
}

func TestSprintTruncates(t *testing.T) {
	long := "+" + strings.Repeat("x", 100)
	out := Sprint(gotui.Dark(), long, 20)
	for _, line := range strings.Split(out, "\n") {
		stripped := line
		if got := len([]rune(stripAnsi(stripped))); got > 20 {
			t.Errorf("line width %d > 20: %q", got, line)
		}
	}
}

func stripAnsi(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		switch {
		case inEsc:
			if r == 'm' {
				inEsc = false
			}
		case r == '\x1b':
			inEsc = true
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
