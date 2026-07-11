// Package grapheme contains small internal helpers for preserving Unicode
// graphemes while ANSI strings are decomposed into terminal cells.
package grapheme

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Protect replaces ASCII-leading, single-cell graphemes that contain
// combining marks with private-use sentinels. Some cell decoders fast-path the
// ASCII base rune and then let following padding overwrite the separate
// width-zero mark. The returned replacements map restores the original text
// after cell placement.
func Protect(view string) (string, map[string]string) {
	views, replacements := ProtectMany(view)
	return views[0], replacements
}

// ProtectMany protects multiple ANSI strings while sharing sentinel
// allocation, so replacements from layered base and overlay views cannot
// collide.
func ProtectMany(views ...string) ([]string, map[string]string) {
	replacements := map[string]string{}
	source := strings.Join(views, "")
	protected := make([]string, len(views))
	nextMarker := rune('\ue000')
	for i, view := range views {
		protected[i] = protectOne(view, source, replacements, &nextMarker)
	}
	return protected, replacements
}

func protectOne(view, source string, replacements map[string]string, nextMarker *rune) string {
	var b strings.Builder
	b.Grow(len(view))
	parser := ansi.GetParser()
	defer ansi.PutParser(parser)

	state := byte(0)
	remaining := view
	for len(remaining) > 0 {
		seq, width, n, nextState := ansi.DecodeSequence(remaining, state, parser)
		if n <= 0 {
			b.WriteByte(remaining[0])
			remaining = remaining[1:]
			state = 0
			continue
		}

		if width == 1 && n == 1 && remaining[0] < 0x80 {
			cluster, _ := ansi.FirstGraphemeCluster(remaining, ansi.GraphemeWidth)
			if len(cluster) > 1 && ansi.StringWidth(cluster) == 1 {
				marker := nextSentinel(source, replacements, nextMarker)
				replacements[marker] = cluster
				b.WriteString(marker)
				remaining = remaining[len(cluster):]
				state = 0
				continue
			}
		}

		b.WriteString(seq)
		remaining = remaining[n:]
		state = nextState
	}
	return b.String()
}

func nextSentinel(source string, replacements map[string]string, nextMarker *rune) string {
	for {
		marker := string(*nextMarker)
		*nextMarker = *nextMarker + 1
		if !strings.Contains(source, marker) {
			if _, exists := replacements[marker]; !exists {
				return marker
			}
		}
	}
}
