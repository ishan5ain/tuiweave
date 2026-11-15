package textarea

import (
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

// Tabs occupy four cells, matching the component's explicit render expansion.
const tabWidth = 4

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
		if cluster == "\t" {
			width = tabWidth
		}
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

func previousClusterStart(line []rune, pos int) int {
	start := 0
	for _, cluster := range clustersOf(line) {
		if cluster.start >= pos {
			break
		}
		start = cluster.start
	}
	return start
}

func clusterStartAt(line []rune, pos int) int {
	for _, cluster := range clustersOf(line) {
		if pos < cluster.end {
			return cluster.start
		}
	}
	return len(line)
}

func nextClusterEnd(line []rune, pos int) int {
	for _, cluster := range clustersOf(line) {
		if cluster.end > pos {
			return cluster.end
		}
	}
	return len(line)
}

func runeOffsetAtCell(text []rune, target, rowWidth int) int {
	if target <= 0 {
		return 0
	}
	used := 0
	for _, cluster := range clustersOf(text) {
		width := min(cluster.width, rowWidth)
		if used+width > target {
			return cluster.start
		}
		used += width
		if used == target {
			return cluster.end
		}
	}
	return len(text)
}

// cellWidth matches clipped clusters in visualRows, including a tab wider
// than the available row. Logical offsets still count each tab as one rune.
func cellWidth(text []rune, rowWidth int) int {
	width := 0
	for _, cluster := range clustersOf(text) {
		width += min(cluster.width, rowWidth)
	}
	return width
}
