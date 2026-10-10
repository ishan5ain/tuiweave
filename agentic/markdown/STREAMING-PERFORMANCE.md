# Streaming Markdown rendering investigation

Investigation for [issue #33](https://github.com/ishan5ain/tuiweave/issues/33).
Production rendering and the public `markdown.Renderer` interface are unchanged.

## Task contract

- Goal: measure cumulative full-document rendering and decide whether follow-up
  work is warranted.
- Reference example: `examples/chat`, including its four-word chunk pattern.
- Components: `markdown.Renderer` and `chat.Assistant`.
- Application-owned state: source/delta sequence and when to request a render;
  benchmarks simulate these without timers, transports, or a Bubble Tea loop.
- Files allowed: Markdown fixtures, external package tests/benchmarks, new
  checkpoint goldens, this report, package README, `DESIGN.md`, and `ROADMAP.md`.
- Behavior to preserve: deterministic output, theme roles, wrapping, partial
  input, error fallback, source replacement, and the existing consumer seam.
- Acceptance tests: fixture reconstruction/UTF-8, repeated partial renders,
  wrapping, streamed versus one-shot assistant output, error fallback, paired
  checkpoint goldens, and all existing tests/goldens.
- Verification commands: package snapshot update/review, `go build ./...`,
  `go vet ./...`, `go test ./...`, and `go test -race ./...`.
- Out of scope: parser/renderer changes, dependency changes, transcript layout,
  scrolling, backend clients, streaming transports, and session orchestration.

## Workloads and measurement boundaries

The checked-in prose and structured Markdown fixtures are repeated to reach
approximately 2, 8, or 32 KiB. Truncation preserves UTF-8 and may leave Markdown
unfinished, as a real stream does. Structured content includes headings,
ordered/unordered lists, emphasis, inline code, links, a quote, and highlighted
Go fences. Both fixtures include wide and accented characters. These are
synthetic workloads, not recordings of production traffic; repetition can
benefit lexer/cache warmup and does not represent every Markdown construct.

Delta patterns are UTF-8-safe chunks of up to 32 or 256 bytes and approximately
four words (`strings.SplitAfter(source, " ")`, matching `examples/chat`).
Widths are 40 and 80 cells. Fixtures, chunk preparation, renderer construction,
width caching, and a full-document warmup are outside the timer. Timed paths
are serial and use the default dark theme through `NewRenderer`:

- `BenchmarkStreaming/Renderer`: concatenate each delta and render the entire
  accumulated source through `Renderer.Render`.
- `BenchmarkStreaming/Assistant`: create an assistant cell per stream, append
  each delta, and call `Assistant.Render`, including `Sprint`, the header, and
  width clipping. Reuse the warmed Markdown renderer across streams.
- `BenchmarkStreaming/AppendOnly`: concatenate identical deltas without
  rendering. This control isolates source copying; it is not a proposed
  alternate renderer.
- `BenchmarkCompleted/Renderer`: repeatedly render the final document, with no
  rendered-output cache. This approximates the cost near the end of a stream
  without hiding it in an average over shorter prefixes.
- `BenchmarkCompleted/AssistantCacheHit`: repeatedly render an unchanged cell.
  Its Markdown output is cached, but header composition and clipping still run.

For streaming benchmarks, `ns/op`, `B/op`, and `allocs/op` refer to **one entire
stream**. `ns/delta` is an average, not a percentile or maximum. `source-B` is
final source length, not cumulative processed bytes. `deltas/op` is the number
of render requests; UTF-8 boundary adjustment may change it from size/32.
Completed controls report one render per operation. Allocation volume measures
cumulative churn, not peak resident memory. No transcript, viewport, terminal
I/O, scheduling delays, or network overhead is included.

## Reproduction

Run benchmarks serially on an otherwise quiet machine, without race or profile
instrumentation. Short single-stream runs are repeated three times so even
expensive cases remain practical; the completed controls use adaptive runs.

```sh
go test ./agentic/markdown -run '^$' -bench '^BenchmarkStreaming$' \
  -benchmem -benchtime=1x -count=3 > bench-streaming.txt
go test ./agentic/markdown -run '^$' -bench '^BenchmarkCompleted$' \
  -benchmem -benchtime=200ms -count=3 > bench-completed.txt
```

For longer samples of one workload, select its complete path and increase
`-benchtime` or `-count`. Keep setup, warmup, delta sequence, source sizes, and
widths identical when comparing revisions. Profiling is a separate run and its
timing is not used in the uninstrumented result tables:

```sh
profile_dir=$(mktemp -d)
go test -c -o "$profile_dir/markdown.test" ./agentic/markdown
for kind in prose structured; do
  "$profile_dir/markdown.test" -test.run '^$' \
    -test.bench "^BenchmarkStreaming$/^Renderer$/$kind/32768/bytes32/w80$" \
    -test.benchtime=1x -test.count=1 -test.benchmem \
    -test.cpuprofile "$profile_dir/$kind.cpu" \
    -test.memprofile "$profile_dir/$kind.mem"
  go tool pprof -top "$profile_dir/markdown.test" "$profile_dir/$kind.cpu"
  go tool pprof -top -cum "$profile_dir/markdown.test" "$profile_dir/$kind.cpu"
  go tool pprof -top -alloc_space "$profile_dir/markdown.test" "$profile_dir/$kind.mem"
  go tool pprof -top -alloc_objects "$profile_dir/markdown.test" "$profile_dir/$kind.mem"
done
```

CPU and memory profiles cover the process, including setup/warmup; their
attribution must be interpreted with that boundary. Inspect callers and
source listings in addition to flat totals, since cumulative percentages
within a call stack overlap and must not be added together.

## Investigation environment

Measured on 2026-10-10 against base revision
`c05ccf548162f861736844f729c18104949c90e5`, with only this investigation's test
and documentation additions. Go `go1.26.2 darwin/arm64`, macOS 26.6.2 (25G83),
Apple M4 Pro, 14 logical CPUs, 48 GiB RAM. Benchmark suffix `-14` records the
default GOMAXPROCS; no `-cpu` override was used. Dependencies remain pinned:
Glamour v2.0.1, Goldmark v1.7.17, Chroma v2.14.0, Lip Gloss v2.0.6, and x/ansi
v0.11.9. Terminal environment: `TERM=xterm-256color`, `COLORTERM=truecolor`,
`NO_COLOR=1`; `CLICOLOR` and `CLICOLOR_FORCE` unset. The renderer receives its
explicit theme stylesheet, not an environment-selected Glamour stylesheet.
No other tests or profiles ran concurrently with the measurements.

The fixture and renderer are reused serially, not shared concurrently. The
renderer width cache is not a rendered-document cache. The assistant revision
cache avoids Markdown work for unchanged source but cannot avoid it after a
nonempty append. Each stream still includes the source copying done by append.

## Uninstrumented results

Raw streaming samples are checked in as
[bench-streaming.txt](testdata/streaming/bench-streaming.txt). The following values are
medians of three complete-stream operations at width 80; MiB means 2^20 bytes.

| Content | Final KiB | 32-byte stream (ms) | 256-byte stream (ms) | 32-byte average delta (ms) | 32-byte MiB/stream | 32-byte million allocations/stream |
|---|---:|---:|---:|---:|---:|---:|
| prose | 2 | 34.14 | 4.73 | 0.53 | 17.4 | 1.96 |
| prose | 8 | 515.22 | 66.15 | 2.00 | 265.1 | 29.70 |
| prose | 32 | 8559.37 | 1124.19 | 8.35 | 4102.6 | 465.32 |
| structured | 2 | 125.07 | 17.35 | 1.95 | 59.0 | 5.86 |
| structured | 8 | 1950.52 | 264.15 | 7.59 | 924.7 | 92.70 |
| structured | 32 | 29206.20 | 3890.19 | 28.49 | 14518.8 | 1468.79 |

At fixed small-delta size, increasing final source size fourfold costs approximately
15–17 times as much cumulative rendering time. That is consistent with repeatedly
rendering an increasingly long document; it does not by itself attribute the work to
Markdown parsing. Larger deltas reduce the number of full renders, while the individual
late-stream renders still process the full document. Width 40 and the four-word pattern
are included in the raw matrix.

For 32 KiB, 32-byte deltas, width 80, the consumer and copying controls give:

| Content | Path | Stream (ms) | Average delta (ms) | MiB/stream |
|---|---|---:|---:|---:|
| prose | Renderer | 8559.374 | 8.3506 | 4102.61 |
| prose | Assistant | 9014.701 | 8.7948 | 4538.22 |
| prose | AppendOnly | 1.842 | 0.0018 | 16.89 |
| structured | Renderer | 29206.198 | 28.4939 | 14518.83 |
| structured | Assistant | 31765.763 | 30.9910 | 16103.57 |
| structured | AppendOnly | 1.558 | 0.0015 | 16.89 |

Source concatenation is a small fraction of measured time and allocation volume.
Assistant header/clipping adds work, but the default full render dominates these
streaming cases. This does not establish which stage inside the renderer is responsible.

Completed-document controls at width 80 (three adaptive runs, median per render):

| Content | Final KiB | Full render (ms) | Full render MiB/op | Assistant cache hit (ms) |
|---|---:|---:|---:|---:|
| prose | 2 | 1.13 | 0.49 | 0.058 |
| prose | 8 | 4.19 | 1.91 | 0.227 |
| prose | 32 | 16.07 | 7.48 | 0.883 |
| structured | 2 | 3.66 | 1.86 | 0.174 |
| structured | 8 | 14.19 | 7.23 | 0.676 |
| structured | 32 | 56.42 | 28.58 | 2.627 |

All completed samples, including width 40 and allocation counts, are in
[bench-completed.txt](testdata/streaming/bench-completed.txt). These are steady-state
whole-document render costs, not observed worst-case or p95 per-delta latency.

## Profile evidence

Separate one-stream profiles use the 32 KiB, 32-byte, width-80 renderer path.
Raw [prose CPU](testdata/streaming/profile-prose-cpu.txt),
[structured CPU](testdata/streaming/profile-structured-cpu.txt),
[prose allocated space](testdata/streaming/profile-prose-alloc-space.txt), and
[structured allocated space](testdata/streaming/profile-structured-alloc-space.txt)
reports are checked in, along with cumulative CPU, full CPU tables, allocated
object estimates, and instrumented run output. Binary profiles remain outside
Git and can be regenerated with the reproduction commands.

| Attribution | Prose | Structured | Interpretation |
|---|---:|---:|---|
| Goldmark `parser.Parse`, cumulative CPU | 0.26% | 0.11% | Markdown parsing is a small sampled cost here |
| Lip Gloss `WrapWriter.Write`, cumulative CPU | 24.31% | 31.28% | Styled wrapping/output writer path |
| `runtime.madvise`, flat CPU | 32.46% | 41.66% | Runtime memory-management system calls |
| `runtime.kevent`, flat CPU | 27.33% | 9.61% | Runtime/GC coordination stacks in these profiles |
| `WrapWriter.Write`, cumulative allocated space | 59.18% | 58.77% | Writer plus its allocating callees |
| `bytes.growSlice`, flat allocated space | 41.23% | 37.99% | Output/buffer growth across call sites |

Structured code highlighting (`CodeBlockElement.Render`/`quick.Highlight`)
accounts for 2.82% cumulative CPU samples. Nested Glamour indent, margin, and
padding writers appear prominently in cumulative CPU stacks. These percentages
are samples, not precise timings; nested cumulative values overlap. Runtime
system-stack samples are not all charged to the triggering renderer frame, so
adding its cumulative percentage to runtime percentages is invalid. System-call
cost and scheduling are particularly machine/OS-specific.

The [source-attributed allocation listing](testdata/streaming/profile-structured-wrapwriter.txt)
points to Lip Gloss v2.0.6 `wrap.go:84`, `w.w.Write([]byte{b})`: roughly 1.20 GiB
flat and 8.21 GiB cumulative sampled allocation for that per-byte write path.
The writer advances an ANSI parser and writes one byte at a time into nested
writers; buffer growth and repeated style-string construction also allocate.
Together, the profiles and pinned source inspection make this output pipeline
and allocation churn the first optimization target. They do not establish that
all runtime time can be eliminated by batching, or that every workload behaves
like these fixtures. Allocation-object profile estimates are sampled and may
differ substantially from benchmark counters, particularly for tiny allocations.

To reproduce the low-percentage parser attribution and the source listing:

```sh
go tool pprof -top -cum -nodefraction=0 -nodecount=10000 \
  "$profile_dir/markdown.test" "$profile_dir/structured.cpu"
go tool pprof -alloc_space -list 'WrapWriter.*Write' \
  "$profile_dir/markdown.test" "$profile_dir/structured.mem"
```

## Decision

Keep the production full-document renderer and public interface unchanged in
this investigation. Shorter fixtures fit comfortably within the example budget:
the 8 KiB structured final render is about 14 ms at width 80. Long, frequent
structured updates warrant separately scoped work: the 32 KiB final render is
about 56 ms, while an application requesting 20 updates/second has only 50 ms
between requests. The assistant and terminal/application work add further cost.
This is an illustrative operating budget, not a claim about a measured production
application or a library-wide latency guarantee.

Opened [issue #49](https://github.com/ishan5ain/tuiweave/issues/49) to investigate
reducing ANSI writer allocation churn, preferably through an upstream fix and
reviewed dependency update. Its proposed success targets on the documented
baseline are <=25 ms median final render (half the illustrative update interval),
<=12.8 seconds for the 1,025-delta structured stream, and at least 50% less
cumulative allocation. It includes output, seam, cache, wrapping, Unicode,
partial-input, fallback, and writer style/link compatibility tests. These are
workload-specific targets to evaluate, not portable CI timing assertions.

The evidence does **not** justify implementing an incremental Markdown parser:
parsing is a very small sampled fraction, and styled output/allocation dominates
the measured path. Re-evaluate remaining costs after a focused writer-path fix
before selecting a broader design. Renderer replacement, public API changes,
transports, session orchestration, transcript layout, and scrolling stay out of
scope.

## Validation

New tests reconstruct every fixture/delta pattern at all benchmark source sizes,
check UTF-8 boundaries, exercise repeated partial renders at both widths, compare
streamed/fresh assistant output, verify wrapping and injected-error fallback,
and snapshot partial emphasis, links, open fences, and the completed response.
New readable and role-aware goldens were generated with
`go test ./agentic/markdown -update` and inspected. Existing Markdown, chat, and
example goldens remain unchanged.

`go build ./...`, `go vet ./...`, `go test ./...`, and `go test -race ./...` pass.
The full streaming matrix has 108 cases, each with three samples; completed
controls have 24 cases, each with three adaptive samples. Full streaming replay
took approximately 15.6 minutes on the investigation machine, so benchmarks are
opt-in rather than part of the normal test run.
