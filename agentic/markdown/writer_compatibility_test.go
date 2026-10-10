package markdown_test

import (
	"bytes"
	"io"
	"testing"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Keep the writer boundary used by Glamour covered when updating Lip Gloss.
// A whole-buffer write must emit the same bytes as arbitrarily split writes,
// including splits inside control sequences or UTF-8 and nested writer chains.
func TestWrapWriterChunkCompatibility(t *testing.T) {
	cases := []struct{ name, source string }{
		{"empty", ""},
		{"unicode", "café 界 e\u0301"},
		{"blank-lines", "\n\n"},
		{"style-across-newline", "\x1b[1mfirst\nsecond\x1b[0m\n"},
		{"style-reset", "\x1b[1mfirst\x1b[0m\nplain"},
		{"hyperlink", ansi.SetHyperlink("https://example.com", "id=fixture") + "one\ntwo" + ansi.ResetHyperlink()},
		{"style-and-link", "\x1b[1m" + ansi.SetHyperlink("https://example.com", "") + "café 界\nnext\n" + ansi.ResetHyperlink() + "\x1b[0m"},
		{"incomplete-style", "\x1b[1m"},
		{"incomplete-csi", "\x1b["},
		{"incomplete-link", "\x1b]8;;https://example.com"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for depth := 1; depth <= 3; depth++ {
				want := renderWriterChunks(t, depth, []byte(tc.source))
				for split := 0; split <= len(tc.source); split++ {
					got := renderWriterChunks(t, depth, []byte(tc.source[:split]), []byte(tc.source[split:]))
					if got != want {
						t.Fatalf("depth=%d split=%d\ngot %q\nwant %q", depth, split, got, want)
					}
				}
				var singleBytes [][]byte
				for i := range len(tc.source) {
					singleBytes = append(singleBytes, []byte(tc.source[i:i+1]))
				}
				if got := renderWriterChunks(t, depth, singleBytes...); got != want {
					t.Fatalf("depth=%d single-byte chunks differ", depth)
				}
			}
		})
	}
}

func renderWriterChunks(t *testing.T, depth int, chunks ...[]byte) string {
	t.Helper()
	var buf bytes.Buffer
	var dst io.Writer = &buf
	var writers []*lipgloss.WrapWriter
	for range depth {
		w := lipgloss.NewWrapWriter(dst)
		writers = append(writers, w)
		dst = w
	}
	for _, chunk := range chunks {
		if n, err := dst.Write(chunk); n != len(chunk) || err != nil {
			t.Fatalf("write: n=%d err=%v", n, err)
		}
	}
	for i := len(writers) - 1; i >= 0; i-- {
		if err := writers[i].Close(); err != nil {
			t.Fatal(err)
		}
	}
	before := buf.String()
	if n, err := dst.Write([]byte("closed")); n != 6 || err != nil || before != buf.String() {
		t.Fatalf("write after close: n=%d err=%v", n, err)
	}
	return before
}

type writerCallCounter struct {
	calls int
	buf   bytes.Buffer
}

func (w *writerCallCounter) Write(p []byte) (int, error) {
	w.calls++
	return w.buf.Write(p)
}

func TestWrapWriterBatchesPlainRuns(t *testing.T) {
	var dst writerCallCounter
	w := lipgloss.NewWrapWriter(&dst)
	if _, err := w.Write([]byte("one\ntwo\nthree")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if got := dst.buf.String(); got != "one\ntwo\nthree" {
		t.Fatalf("output=%q", got)
	}
	if dst.calls > 5 {
		t.Fatalf("downstream writes=%d, want at most one per run/newline", dst.calls)
	}
}
