package paste

import "testing"

func TestNormalize(t *testing.T) {
	for _, tc := range []struct{ name, source, single, multi string }{
		{"plain", "hello", "hello", "hello"},
		{"line endings", "a\r\nb\rc\nd\te", "a b c d e", "a\nb\nc\nd\te"},
		{"whitespace", "  a\n\n b  ", "  a   b  ", "  a\n\n b  "},
		{"unicode", "界e\u0301👩‍💻", "界e\u0301👩‍💻", "界e\u0301👩‍💻"},
		{"controls", "a\x00\x07\x08\x7f\u0085b", "ab", "ab"},
		{"escapes", "\x1b[31mred\x1b[0m\x1b]8;;https://example.com\x07link\x1b]8;;\x07", "redlink", "redlink"},
		{"terminal commands", "a\x1b[2Jb\x1bPignored\x1b\\c", "abc", "abc"},
		{"unfinished escape", "a\x1b[31", "a", "a"},
		{"empty", "", "", ""},
		{"filtered empty", "\x1b[31m\x00\x7f", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, mode := range []Mode{SingleLine, Multiline} {
				want := tc.single
				if mode == Multiline {
					want = tc.multi
				}
				if got := Normalize(tc.source, mode); got != want {
					t.Fatalf("mode %d: got %q, want %q", mode, got, want)
				}
			}
		})
	}
}
