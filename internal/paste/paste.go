// Package paste normalizes terminal paste for the input components.
package paste

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Mode determines which whitespace is permitted in pasted text.
type Mode uint8

const (
	SingleLine Mode = iota
	Multiline
)

// Normalize strips terminal escapes and controls, and normalizes line endings.
// Multiline input preserves LF and tabs; single-line input replaces them with
// spaces. Unicode formatting characters such as ZWJ are preserved.
func Normalize(content string, mode Mode) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	return strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\t':
			if mode == Multiline {
				return r
			}
			return ' '
		}
		if r < 0x20 || r >= 0x7f && r <= 0x9f {
			return -1
		}
		return r
	}, ansi.Strip(content))
}
