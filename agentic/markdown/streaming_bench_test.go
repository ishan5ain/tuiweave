package markdown_test

import (
	"fmt"
	"testing"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/agentic/chat"
	"github.com/ishan5ain/tuiweave/agentic/markdown"
)

var streamingResult string

// One operation is an entire append-and-render stream, not one delta.
// Renderer construction and width/lexer warmup are outside the timer. Source
// concatenation is timed, matching Assistant.Append; AppendOnly isolates it.
func BenchmarkStreaming(b *testing.B) {
	for _, path := range []string{"Renderer", "Assistant", "AppendOnly"} {
		for _, kind := range []string{"prose", "structured"} {
			for _, size := range []int{2048, 8192, 32768} {
				for _, pattern := range []string{"bytes32", "words4", "bytes256"} {
					for _, width := range []int{40, 80} {
						name := fmt.Sprintf("%s/%s/%d/%s/w%d", path, kind, size, pattern, width)
						b.Run(name, func(b *testing.B) {
							full := fixture(b, kind, size)
							chunks := deltas(full, pattern)
							theme := tuiweave.Dark()
							r := markdown.NewRenderer(theme)
							if _, err := r.Render(full, width); err != nil {
								b.Fatal(err)
							}
							b.ReportAllocs()
							for b.Loop() {
								switch path {
								case "Assistant":
									a := chat.NewAssistant(theme, r)
									for _, chunk := range chunks {
										a.Append(chunk)
										streamingResult = a.Render(width)
									}
								default:
									var source string
									for _, chunk := range chunks {
										source += chunk
										if path == "Renderer" {
											out, err := r.Render(source, width)
											if err != nil {
												b.Fatal(err)
											}
											streamingResult = out
										} else {
											streamingResult = source
										}
									}
								}
							}
							b.ReportMetric(float64(len(full)), "source-B")
							b.ReportMetric(float64(len(chunks)), "deltas/op")
							b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*len(chunks)), "ns/delta")
						})
					}
				}
			}
		}
	}
}

// Completed provides late-stream costs without averaging early short prefixes.
// AssistantCacheHit still includes header composition and width clipping; only
// Markdown output is cached. Neither control includes transcript or viewport.
func BenchmarkCompleted(b *testing.B) {
	for _, path := range []string{"Renderer", "AssistantCacheHit"} {
		for _, kind := range []string{"prose", "structured"} {
			for _, size := range []int{2048, 8192, 32768} {
				for _, width := range []int{40, 80} {
					b.Run(fmt.Sprintf("%s/%s/%d/w%d", path, kind, size, width), func(b *testing.B) {
						full := fixture(b, kind, size)
						r := markdown.NewRenderer(tuiweave.Dark())
						if _, err := r.Render(full, width); err != nil {
							b.Fatal(err)
						}
						a := chat.NewAssistant(tuiweave.Dark(), r)
						a.SetSource(full)
						a.Render(width)
						b.ReportAllocs()
						for b.Loop() {
							if path == "Renderer" {
								out, err := r.Render(full, width)
								if err != nil {
									b.Fatal(err)
								}
								streamingResult = out
							} else {
								streamingResult = a.Render(width)
							}
						}
						b.ReportMetric(float64(len(full)), "source-B")
					})
				}
			}
		}
	}
}
