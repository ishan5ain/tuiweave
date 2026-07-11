package textarea

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

func runeOffsetAtCell(text []rune, target int) int {
	if target <= 0 {
		return 0
	}
	used := 0
	for _, cluster := range clustersOf(text) {
		if used+cluster.width > target {
			return cluster.start
		}
		used += cluster.width
		if used == target {
			return cluster.end
		}
	}
	return len(text)
}
