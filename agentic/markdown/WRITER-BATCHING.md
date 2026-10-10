# ANSI writer batching experiment

Follow-up to [issue #49](https://github.com/ishan5ain/tuiweave/issues/49) and the
[streaming investigation](STREAMING-PERFORMANCE.md), landed in
[PR #50](https://github.com/ishan5ain/tuiweave/pull/50).

## Task contract

- Goal: isolate writer batching, measure its end-to-end effect, validate output,
  and prepare a compatible dependency update for review.
- Reference example: `examples/chat` and the #33 streaming fixtures.
- Components: Lip Gloss `WrapWriter`, `Wrap`, `markdown.Renderer`, and
  `chat.Assistant`; source/delta timing remains application-owned.
- Files allowed: isolated dependency checkout and development worktree,
  benchmark/test evidence, this report, dependency pins, compatibility docs,
  README, ROADMAP, architecture note, and release notes.
- Behavior to preserve: exact styled bytes, newline style/link restoration,
  arbitrary write boundaries, nested writers, Unicode, partial Markdown, cache
  invalidation, and the renderer seam.
- Acceptance tests: writer compatibility oracle, write-call count, existing
  package tests/goldens, build, vet, module verification/tidy, and race tests.
- Out of scope: parser/renderer replacement, public API changes, application
  cadence policy, transcript layout, and session orchestration.

## Upstream change selection

The planned batching prototype already exists upstream:

- [c7d9f357](https://github.com/charmbracelet/lipgloss/commit/c7d9f3577af5c4a8961caea03b4188dcad5f5e6b)
  forwards runs between newlines instead of allocating/writing one byte at a time.
- [868e8b5c](https://github.com/charmbracelet/lipgloss/commit/868e8b5c4fb6056c294b9fda2fb9d1090052daee)
  also pre-sizes the output buffer in `Wrap`.

The experiment uses exact commit `868e8b5c4fb6056c294b9fda2fb9d1090052daee`,
resolved as `charm.land/lipgloss/v2 v2.0.7-0.20260901023042-868e8b5c4fb6`.
Compared with v2.0.6, only `wrap.go` and upstream Dependabot configuration
change. The dependency's module requirements are identical and its Go minimum
remains 1.25.0; tuiweave's Go minimum remains 1.26.0. Current upstream main has
other changes and a higher Go minimum, so this experiment deliberately pins the
specific writer commits. There is no application-local fork or committed
`replace` directive.

## Isolated writer checks

An isolated upstream checkout contains the
[writer test/benchmark harness](testdata/writer-batching/wrap_batch_test.go.txt).
Its test-only legacy writer reproduces v2.0.6's byte-forwarding implementation
as an oracle. The new writer matches its exact output for plain text, bold
styling, hyperlink opening/closing, style changes, Unicode, combining text,
empty lines, incomplete control sequences, every two-chunk byte split, and
chains of one to three nested writers. Writes after close remain harmless.
A separate check confirms five downstream writes for three plain runs separated
by two newlines, rather than one write per input byte.

The microbenchmark writes a fixed styled Unicode response through three writer
layers into `io.Discard`, isolating forwarding and style/newline work. The same
process compares legacy and batched implementations over five adaptive runs.
It is not a Markdown/transcript benchmark and should not be used to predict
end-to-end speedup. Its standalone module uses the pinned upstream dependencies
(x/ansi v0.11.8); end-to-end tuiweave measurements retain x/ansi v0.11.9.

To reproduce the isolated checks without a fork:

```sh
gh repo clone charmbracelet/lipgloss lipgloss-writer-lab
cd lipgloss-writer-lab
git checkout 868e8b5c4fb6056c294b9fda2fb9d1090052daee
# Copy wrap_batch_test.go.txt from tuiweave's testdata as wrap_batch_test.go.
go test ./...
go test -race ./...
go test -run '^TestWrapBatch' -bench '^BenchmarkWrapBatch$' \
  -benchmem -benchtime=300ms -count=5
```

Raw samples are in [bench-wrap-writer.txt](testdata/writer-batching/bench-wrap-writer.txt).

| Writer | Median µs/op | B/op | Allocations/op |
|---|---:|---:|---:|
| legacy | 380.7 | 37419 | 34013 |
| batched | 183.7 | 3733 | 610 |

## End-to-end methodology

Measurements use the unchanged #33 fixtures, size/chunk helpers, dark theme,
and renderer/assistant seams on Go 1.26.2, macOS 26.6.2 (25G83), Apple M4 Pro,
14 logical CPUs/default GOMAXPROCS, and 48 GiB RAM, on 2026-10-10. Terminal
settings match the investigation: TERM=xterm-256color, COLORTERM=truecolor,
NO_COLOR=1, CLICOLOR/CLICOLOR_FORCE unset. Renderer construction and warmup are
outside timing; source concatenation is included. No tests or profiles ran
concurrently with a benchmark measurement.

The baseline is merged revision `2e7042f94ddd19f94ad9ba03a0b412acdbe17816`,
using Lip Gloss v2.0.6. The candidate changes only the selected Lip Gloss pin
and checksum, with no changes to other selected module versions or the public
renderer. A local development replacement was used for initial checks, then
removed; recorded candidate runs use the downloaded, checksum-verified exact
pseudo-version. Both dependency manifests require the same modules/versions.

Fresh baseline completed/cache controls and the worst structured streaming case
were run immediately before the candidate. The candidate renderer matrix covers
both content types, 2/8/32 KiB, 32/256-byte deltas, and widths 40/80 with three
complete-stream samples per case. A three-sample 32 KiB small-delta assistant
case checks the consumer. All 24 completed/cache controls use three adaptive
200 ms runs. Comparisons outside the freshly repeated baseline case use the
three-sample #33 dataset from earlier on the same machine, and are labeled as
such; they are useful regression checks rather than perfectly paired samples.

Run the following serially in separate clean baseline/candidate worktrees:

```sh
# Fresh baseline and candidate completed controls:
go test ./agentic/markdown -run '^$' -bench '^BenchmarkCompleted$' \
  -benchmem -benchtime=200ms -count=3
# Fresh baseline worst case:
go test ./agentic/markdown -run '^$' \
  -bench '^BenchmarkStreaming/Renderer/structured/32768/bytes32/w80$' \
  -benchmem -benchtime=1x -count=3
# Candidate renderer matrix:
go test ./agentic/markdown -run '^$' \
  -bench '^BenchmarkStreaming$/^Renderer$/(prose|structured)/(2048|8192|32768)/(bytes32|bytes256)/(w40|w80)$' \
  -benchmem -benchtime=1x -count=3
# Candidate consumer case:
go test ./agentic/markdown -run '^$' \
  -bench '^BenchmarkStreaming/Assistant/structured/32768/bytes32/w80$' \
  -benchmem -benchtime=1x -count=3
```

Streaming ns/op, B/op, and allocations/op describe an entire response;
ns/delta is an average and completed controls describe one full render.
Neither is a measured p95 or delivery latency. Allocated bytes measure cumulative
churn rather than resident memory. Source/delta cadence remains application-owned.

## Completed-document results

Medians of three freshly repeated baseline/candidate runs at width 80:

| Content | KiB | Baseline ms/render | Candidate ms/render | Time reduction | Baseline MiB/render | Candidate MiB/render |
|---|---:|---:|---:|---:|---:|---:|
| prose | 2 | 1.05 | 0.83 | 20.3% | 0.48 | 0.42 |
| prose | 8 | 3.95 | 3.28 | 17.0% | 1.91 | 1.65 |
| prose | 32 | 15.57 | 12.52 | 19.6% | 7.47 | 6.40 |
| structured | 2 | 3.60 | 2.81 | 21.9% | 1.86 | 1.60 |
| structured | 8 | 13.70 | 10.99 | 19.8% | 7.22 | 6.26 |
| structured | 32 | 53.95 | 43.88 | 18.7% | 28.54 | 24.73 |

At 32 KiB structured/width 80, allocations per full render fall from 2,869,431 to
630,593 (78.0% fewer). Allocated bytes fall much less because output buffers, copying,
and other rendering work remain.

Raw [baseline completed controls](testdata/writer-batching/bench-baseline-completed.txt)
and [candidate completed
controls](testdata/writer-batching/bench-candidate-completed.txt) include both widths,
cache hits, and all allocation counts.

## Streaming results

Fresh 32 KiB structured/32-byte/width-80 comparison (1,025 render requests per operation):

| Metric | Baseline | Candidate | Reduction |
|---|---:|---:|---:|
| Seconds/stream | 27.54 | 22.88 | 16.9% |
| Average ms/delta | 26.86 | 22.32 | 16.9% |
| GiB allocated/stream | 14.18 | 12.41 | 12.5% |
| Million allocations/stream | 1468.79 | 322.33 | 78.1% |

The following width-80 comparisons use the earlier #33 baseline for the broader
regression matrix (three samples per case). Fresh baseline samples above are preferable
for the headline improvement.

| Content | KiB | Delta bytes | Earlier baseline seconds/stream | Candidate seconds/stream |
|---|---:|---:|---:|---:|
| prose | 2 | 32 | 0.034 | 0.028 |
| prose | 2 | 256 | 0.005 | 0.004 |
| prose | 8 | 32 | 0.515 | 0.404 |
| prose | 8 | 256 | 0.066 | 0.051 |
| prose | 32 | 32 | 8.559 | 6.319 |
| prose | 32 | 256 | 1.124 | 0.794 |
| structured | 2 | 32 | 0.125 | 0.091 |
| structured | 2 | 256 | 0.017 | 0.013 |
| structured | 8 | 32 | 1.951 | 1.435 |
| structured | 8 | 256 | 0.264 | 0.194 |
| structured | 32 | 32 | 29.206 | 22.880 |
| structured | 32 | 256 | 3.890 | 2.952 |

Raw [fresh baseline stream](testdata/writer-batching/bench-baseline-streaming.txt) and
[candidate matrix](testdata/writer-batching/bench-candidate-streaming.txt) retain every
sample, both widths, and allocation metrics. Across the candidate matrix, median time
and allocation counts improve compared with the earlier baseline; no sampled rendering
case regresses. This is limited to the tested fixtures and machine.

The candidate assistant consumer takes 23.85 seconds per stream,
23.27 ms per delta on average, and 13.96 GiB
cumulative allocation. Its three samples are in
[bench-candidate-assistant.txt](testdata/writer-batching/bench-candidate-assistant.txt).
The earlier #33 assistant baseline was 31.77 seconds; that comparison is not a
fresh paired consumer measurement. Header and clipping overhead remain outside
the upstream writer change.

## Remaining profile evidence

A separate instrumented candidate run profiles the 32 KiB structured response,
32-byte deltas, width 80. Timing from that run is excluded from the benchmark
comparison. Raw [CPU](testdata/writer-batching/profile-candidate-cpu.txt),
[cumulative CPU](testdata/writer-batching/profile-candidate-cpu-cumulative.txt),
[allocated space](testdata/writer-batching/profile-candidate-alloc-space.txt),
[allocated objects](testdata/writer-batching/profile-candidate-alloc-objects.txt),
and [run output](testdata/writer-batching/profile-candidate-run.txt) are retained.

Direct sampled allocation in `WrapWriter.Write` is about 2.5 MiB, compared with
roughly 1.20 GiB in the earlier profile. The direct per-byte allocation has
largely disappeared, as expected from the upstream source change. Its cumulative
allocation remains substantial because downstream output and style processing
still allocate.

The remaining flat allocation sites include `bytes.growSlice` (38.23%) and
`bytes.Buffer.String` (11.93%), together about half the sampled allocated space.
Style construction/foreground formatting also remains prominent. CPU samples
include nested Glamour `IndentWriter.Write` (24.70% cumulative) and
`PaddingWriter.Write` (17.70% cumulative), Lip Gloss `WrapWriter.Write` (26.64%
cumulative), and runtime memory-management/coordination calls. Goldmark parsing
is only about 0.24% cumulative CPU, while highlighting is about 3.6%.
Cumulative percentages overlap and must not be added; runtime/system-call costs
and sampling are machine-specific. Object profile estimates do not exactly match
benchmark allocation counts, especially for tiny allocations.

This supports a next focused experiment in nested writer forwarding, padding
style generation, and output-buffer copying. Preserve callback behavior and
style/link boundaries when investigating those paths. It does not support
implementing an incremental Markdown parser as the next step.

To reproduce the candidate profile separately from tests/measurements:

```sh
profile_dir=$(mktemp -d)
go test -c -o "$profile_dir/markdown.test" ./agentic/markdown
"$profile_dir/markdown.test" -test.run '^$' \
  -test.bench '^BenchmarkStreaming/Renderer/structured/32768/bytes32/w80$' \
  -test.benchtime=1x -test.count=1 -test.benchmem \
  -test.cpuprofile "$profile_dir/candidate.cpu" \
  -test.memprofile "$profile_dir/candidate.mem"
go tool pprof -top "$profile_dir/markdown.test" "$profile_dir/candidate.cpu"
go tool pprof -top -cum -nodefraction=0 -nodecount=10000 \
  "$profile_dir/markdown.test" "$profile_dir/candidate.cpu"
go tool pprof -top -alloc_space "$profile_dir/markdown.test" "$profile_dir/candidate.mem"
go tool pprof -top -alloc_objects "$profile_dir/markdown.test" "$profile_dir/candidate.mem"
```

Profiles cover setup/warmup as well as the timed stream; they are sampled evidence
rather than precise accounting for just the timed operation.

## Decision and remaining targets

Adopt the exact upstream writer commit as a measured, compatible first
improvement. It requires no renderer or application API change, retains the
Go floor and dependency requirements, and improves every sampled renderer case
in this matrix. The pin is an unreleased upstream commit, not a tagged release;
upgrade to a later tag only after an equivalent compatibility review.

The broader [issue #49](https://github.com/ishan5ain/tuiweave/issues/49) remains
open. Against its illustrative workload-specific targets:

| Target | Measured candidate | Status |
|---|---|---|
| <=25 ms median final structured render | 43.88 ms | Remaining |
| <=12.8 seconds cumulative small-delta structured stream | 22.88 seconds | Remaining |
| >=50% reduction in cumulative allocated bytes | 12.5% versus fresh baseline | Remaining |

Allocation **count** falls about 78%; that does not meet a target about allocated
**bytes**. The example 20 Hz cadence and half-interval rendering budget are
engineering assumptions that still need validation in a real consumer; these
measurements do not establish delivery latency, a p95, or a universal supported
response size. Keep the benchmark fixtures and profile evidence as regression
baselines, and review the actual consumer's cadence/response distribution before
turning these targets into release requirements.

## Validation

The isolated upstream checkout passes its full test and race suites, including
legacy byte-for-byte checks at all split positions. tuiweave additionally keeps
writer chunk-invariance, single-byte input, nested writers, post-close behavior,
and plain-run write-call regression tests in
`agentic/markdown/writer_compatibility_test.go` so they run in normal CI.
A separate baseline/candidate binary comparison hashes raw styled Markdown output
for both fixtures, 2/8/32 KiB documents, partial emphasis/link/fence checkpoints,
and widths 40/80 (22 renders). All hashes match exactly; the checked-in
[baseline hashes](testdata/writer-batching/render-baseline-sha256.txt) and
[candidate hashes](testdata/writer-batching/render-candidate-sha256.txt) retain
the results. Reproduce by copying
[render_hashes.go.txt](testdata/writer-batching/render_hashes.go.txt) to a temporary
`.go` file, running it from each clean version's module root with the same fixture
root as its argument, and sorting/comparing stdout. This complements the writer
oracle and role-aware goldens with a raw-byte check for the measured corpus.

Existing Markdown/chat/example and other component goldens remain unchanged;
no snapshot update was required because output is preserved.

Module verification and tidy, formatting, build, vet, normal tests, and race
tests validate the pinned graph. Full cross-platform CI results are recorded in
the dependency PR. No development replacement or isolated test fork is committed
as a dependency.
