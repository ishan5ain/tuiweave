package textinput

import (
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

type runeCluster struct {
	start int
	end   int
	width int
}

func clustersOf(line []rune) []runeCluster {
	if len(line) == 0 {
		return nil
	}
	text := string(line)
	clusters := make([]runeCluster, 0, len(line))
	byteOffset, runeOffset := 0, 0
	for byteOffset < len(text) {
		cluster, width := ansi.FirstGraphemeCluster(text[byteOffset:], ansi.GraphemeWidth)
		runeCount := utf8.RuneCountInString(cluster)
		if runeCount == 0 {
			runeCount = 1
		}
		clusters = append(clusters, runeCluster{
			start: runeOffset,
			end:   runeOffset + runeCount,
			width: width,
		})
		byteOffset += len(cluster)
		runeOffset += runeCount
	}
	return clusters
}

func cellWidth(runes []rune) int {
	return ansi.StringWidth(string(runes))
}

func previousClusterStart(value []rune, pos int) int {
	start := 0
	for _, cluster := range clustersOf(value) {
		if cluster.start >= pos {
			break
		}
		start = cluster.start
	}
	return start
}

func nextClusterEnd(value []rune, pos int) int {
	for _, cluster := range clustersOf(value) {
		if cluster.end > pos {
			return cluster.end
		}
	}
	return len(value)
}

// windowForCursor returns a cluster-aligned visible range. Space before the
// cursor is preferred, matching the usual single-line input behavior; when
// the cursor is near the beginning, remaining cells are filled to the right.
func windowForCursor(value []rune, pos, available int, focused bool) (start, end, focusStart, focusEnd int, focusVisible bool) {
	clusters := clustersOf(value)
	if len(clusters) == 0 {
		return 0, 0, 0, 0, false
	}

	focusIndex := len(clusters)
	focusStart, focusEnd = len(value), len(value)
	for i, cluster := range clusters {
		if pos >= cluster.start && pos < cluster.end {
			focusIndex = i
			focusStart, focusEnd = cluster.start, cluster.end
			break
		}
	}

	textCapacity := available
	if focused {
		textCapacity-- // preserve the input's one-cell cursor gutter
	}
	if textCapacity <= 0 {
		return focusStart, focusStart, focusStart, focusEnd, false
	}

	focusWidth := cellWidth(value[focusStart:focusEnd])
	if focusStart == focusEnd {
		focusWidth = 0
	}
	if focusWidth > textCapacity {
		return focusStart, focusStart, focusStart, focusEnd, false
	}

	start, end = focusStart, focusEnd
	used := focusWidth
	for i := focusIndex - 1; i >= 0; i-- {
		cluster := clusters[i]
		if used+cluster.width > textCapacity {
			break
		}
		start = cluster.start
		used += cluster.width
	}
	for i := focusIndex + 1; i < len(clusters); i++ {
		cluster := clusters[i]
		if used+cluster.width > textCapacity {
			break
		}
		end = cluster.end
		used += cluster.width
	}
	if focusIndex == len(clusters) {
		end = len(value)
	}
	return start, end, focusStart, focusEnd, true
}
