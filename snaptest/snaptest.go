// Package snaptest is gotui's snapshot test harness: it renders component
// views to golden files and compares them on subsequent runs, giving agents
// (and humans) a deterministic, readable way to verify what a TUI renders
// without a terminal.
//
// Two artifacts per snapshot:
//
//   - <name>.golden — the view with ANSI stripped and trailing spaces
//     trimmed. Legible in a git diff; this is the file to read when deciding
//     whether a change is correct.
//   - <name>.styled.golden — the raw view, escape codes included (written by
//     SnapStyled only). Guards styling regressions byte-for-byte.
//   - <name>.scenario.golden — named interaction checkpoints with plain views
//     and command-emission status (written by SnapScenario).
//
// Typical usage:
//
//	func TestStatusbarDefault(t *testing.T) {
//		sb := statusbar.New(gotui.Dark())
//		sb.SetSize(80, 1)
//		snaptest.Snap(t, sb.View())
//	}
//
// Regenerate goldens after an intentional change with:
//
//	go test ./... -run TestStatusbarDefault -update
//
// then read the golden diff in git to confirm the change is what you meant.
package snaptest

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

var update = flag.Bool("update", false, "regenerate snaptest golden files instead of comparing")

// Snap compares the plain-text rendering of view (ANSI stripped, trailing
// spaces trimmed) against testdata/<TestName>.golden, failing the test with a
// line diff on mismatch. With -update it (re)writes the golden instead.
func Snap(t *testing.T, view string) {
	t.Helper()
	compare(t, goldenPath(t, ".golden"), normalize(ansi.Strip(view)))
}

// SnapStyled snapshots both artifacts: the plain golden (as Snap) and the
// raw styled view, escape codes included, at testdata/<TestName>.styled.golden.
func SnapStyled(t *testing.T, view string) {
	t.Helper()
	Snap(t, view)
	compare(t, goldenPath(t, ".styled.golden"), ensureTrailingNewline(view))
}

// normalize trims trailing spaces per line and guarantees a final newline,
// so goldens survive editors that strip whitespace and stay readable in diffs.
func normalize(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " ")
	}
	return ensureTrailingNewline(strings.Join(lines, "\n"))
}

func ensureTrailingNewline(s string) string {
	if !strings.HasSuffix(s, "\n") {
		return s + "\n"
	}
	return s
}

// goldenPath maps the running (sub)test to a file under testdata.
func goldenPath(t *testing.T, suffix string) string {
	name := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		case r == '.', r == '-':
			return r
		default:
			return '_'
		}
	}, t.Name())
	return filepath.Join("testdata", name+suffix)
}

func compare(t *testing.T, path, got string) {
	t.Helper()

	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("snaptest: creating %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("snaptest: writing golden %s: %v", path, err)
		}
		t.Logf("snaptest: updated %s", path)
		return
	}

	wantBytes, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		t.Fatalf("snaptest: golden %s does not exist; run: go test -run %q -update", path, t.Name())
	}
	if err != nil {
		t.Fatalf("snaptest: reading golden %s: %v", path, err)
	}

	want := string(wantBytes)
	if want == got {
		return
	}
	t.Errorf("snaptest: %s does not match rendered output:\n%s", path, diffLines(want, got))
}

// diffLines reports mismatched lines between want and got, capped so a badly
// broken render stays readable.
func diffLines(want, got string) string {
	const maxDiffs = 20

	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")
	n := max(len(wantLines), len(gotLines))

	var b strings.Builder
	diffs := 0
	for i := range n {
		w, g := "<missing line>", "<missing line>"
		if i < len(wantLines) {
			w = wantLines[i]
		}
		if i < len(gotLines) {
			g = gotLines[i]
		}
		if w == g {
			continue
		}
		if diffs == maxDiffs {
			b.WriteString("... more differences elided\n")
			break
		}
		diffs++
		fmt.Fprintf(&b, "line %d:\n  want: %q\n  got:  %q\n", i+1, w, g)
	}
	return b.String()
}
