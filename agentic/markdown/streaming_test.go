package markdown_test

import (
	"embed"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/agentic/chat"
	"github.com/ishan5ain/tuiweave/agentic/markdown"
	"github.com/ishan5ain/tuiweave/snaptest"
)

//go:embed testdata/streaming/*.md
var streamingFixtures embed.FS

func fixture(tb testing.TB, kind string, size int) string {
	tb.Helper()
	data, err := streamingFixtures.ReadFile("testdata/streaming/" + kind + ".md")
	if err != nil {
		tb.Fatal(err)
	}
	// Keep a blank paragraph separator when repeating fixtures. Adding it here
	// avoids a whitespace-only final line in the checked-in Markdown files.
	unit := string(data) + "\n"
	if size == 0 {
		return unit
	}
	source := strings.Repeat(unit, (size+len(unit)-1)/len(unit))
	// Sizes are approximate byte counts; never end within a UTF-8 rune.
	for size < len(source) && !utf8.RuneStart(source[size]) {
		size--
	}
	return source[:size]
}

func deltas(source, pattern string) []string {
	if pattern == "words4" {
		words := strings.SplitAfter(source, " ") // same pattern as examples/chat
		var chunks []string
		for i := 0; i < len(words); i += 4 {
			chunks = append(chunks, strings.Join(words[i:min(i+4, len(words))], ""))
		}
		return chunks
	}
	size := 32
	if pattern == "bytes256" {
		size = 256
	}
	var chunks []string
	for len(source) > 0 {
		n := min(size, len(source))
		for n < len(source) && !utf8.RuneStart(source[n]) {
			n--
		}
		chunks = append(chunks, source[:n])
		source = source[n:]
	}
	return chunks
}

func TestStreamingFixtures(t *testing.T) {
	for _, kind := range []string{"prose", "structured"} {
		for _, pattern := range []string{"bytes32", "words4", "bytes256"} {
			t.Run(kind+"/"+pattern, func(t *testing.T) {
				full := fixture(t, kind, 0)
				chunks := deltas(full, pattern)
				if strings.Join(chunks, "") != full {
					t.Fatal("deltas do not reconstruct the fixture")
				}
				for _, chunk := range chunks {
					if !utf8.ValidString(chunk) {
						t.Fatal("delta splits a UTF-8 rune")
					}
				}
				for _, width := range []int{40, 80} {
					r := markdown.NewRenderer(tuiweave.Dark())
					a := chat.NewAssistant(tuiweave.Dark(), r)
					var source, got string
					for _, chunk := range chunks {
						source += chunk
						a.Append(chunk)
						out, err := r.Render(source, width)
						if err != nil {
							t.Fatalf("partial source: %v", err)
						}
						again, err := r.Render(source, width)
						if err != nil || out != again {
							t.Fatal("partial render is not deterministic")
						}
						for _, row := range strings.Split(out, "\n") {
							if ansi.StringWidth(row) > width {
								t.Fatalf("render exceeds width %d: %q", width, row)
							}
						}
						got = a.Render(width)
					}
					fresh := chat.NewAssistant(tuiweave.Dark(), markdown.NewRenderer(tuiweave.Dark()))
					fresh.SetSource(full)
					if a.Source() != full || got != fresh.Render(width) {
						t.Fatal("streamed assistant differs from one-shot assistant")
					}
				}
			})
		}
	}
}

func TestStreamingCheckpointsGolden(t *testing.T) {
	source := fixture(t, "structured", 0)
	checkpoints := []struct{ name, end string }{
		{"partial-emphasis", "**request"},
		{"partial-link", "[runbook](https://example.com/"},
		{"open-fence", "    return pending < 12\n"},
		{"complete", ""},
	}
	for _, checkpoint := range checkpoints {
		t.Run(checkpoint.name, func(t *testing.T) {
			prefix := source
			if checkpoint.end != "" {
				i := strings.Index(source, checkpoint.end)
				if i < 0 {
					t.Fatalf("missing checkpoint %q", checkpoint.end)
				}
				prefix = source[:i+len(checkpoint.end)]
			}
			r := markdown.NewRenderer(tuiweave.Dark())
			out, err := r.Render(prefix, 40)
			if err != nil {
				t.Fatal(err)
			}
			snaptest.Snap(t, out)
			snaptest.SnapCells(t, out, snaptest.WithRoles(tuiweave.Dark()))
		})
	}
}

type failingRenderer struct{}

func (failingRenderer) Render(string, int) (string, error) {
	return "", fmt.Errorf("fixture renderer failed")
}

func TestStreamingFallback(t *testing.T) {
	for _, kind := range []string{"prose", "structured"} {
		var source string
		for _, delta := range deltas(fixture(t, kind, 0), "bytes32") {
			source += delta
			if got := markdown.Sprint(failingRenderer{}, source, 40); got != source {
				t.Fatal("fallback did not preserve partial source")
			}
		}
	}
}

func TestStreamingFixtureSizes(t *testing.T) {
	for _, kind := range []string{"prose", "structured"} {
		for _, size := range []int{2048, 8192, 32768} {
			source := fixture(t, kind, size)
			if !utf8.ValidString(source) || len(source) > size || len(source) < size-3 {
				t.Fatalf("%s/%d: invalid fixture length or UTF-8", kind, size)
			}
			for _, width := range []int{40, 80} {
				r := markdown.NewRenderer(tuiweave.Dark())
				out, err := r.Render(source, width)
				if err != nil {
					t.Fatal(err)
				}
				again, err := r.Render(source, width)
				if err != nil || again != out {
					t.Fatal("sized fixture render is not deterministic")
				}
				for _, row := range strings.Split(out, "\n") {
					if ansi.StringWidth(row) > width {
						t.Fatalf("sized fixture exceeds width %d", width)
					}
				}
			}
			for _, pattern := range []string{"bytes32", "words4", "bytes256"} {
				chunks := deltas(source, pattern)
				if strings.Join(chunks, "") != source {
					t.Fatal("sized chunks do not reconstruct source")
				}
				for _, chunk := range chunks {
					if chunk == "" || !utf8.ValidString(chunk) {
						t.Fatal("empty or invalid UTF-8 delta")
					}
				}
			}
		}
	}
}
