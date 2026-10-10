# Writer experiment artifact provenance

This manifest maps the raw files to their producing code, dependency graph,
commands, and environment. Raw measurements are retained unchanged; filenames
alone should not be used to infer which graph produced a sample.

## Code snapshots

- **B — baseline:** tuiweave
  `2e7042f94ddd19f94ad9ba03a0b412acdbe17816`, Lip Gloss v2.0.6.
- **C — candidate:** the same benchmark/fixture source as B, with Lip Gloss
  changed to `v2.0.7-0.20260901023042-868e8b5c4fb6` (upstream commit
  `868e8b5c4fb6056c294b9fda2fb9d1090052daee`). The reproducible code and original
  artifacts are archived in tuiweave
  `5d56ab6a6c421e756797b60234e043228f5b76e1`. Measurements were collected in the
  working tree before that archival commit; do not interpret its commit time
  as the measurement time. Benchmark bodies and generated fixture contents
  match B; the only selected module-version change is Lip Gloss.
- **W — isolated writer experiment:** upstream Lip Gloss commit
  `868e8b5c4fb6056c294b9fda2fb9d1090052daee`, plus the archived
  `wrap_batch_test.go.txt` copied as `wrap_batch_test.go` into that module.
  The test-only legacy implementation is the v2.0.6 forwarding algorithm.
  The local experiment was committed as `9d89404`; its full test source is
  preserved here so reproducing it does not depend on that local commit.
- **H — Markdown hash generator:** `render_hashes.go.txt`, originally archived
  in `5d56ab6`. Run the same generator/fixtures from B and C independently.
- **R — direct Wrap review generator:** `wrap_cases.go.txt`, added during the
  independent-review follow-up. Run it from B and C independently; it produces
  ten direct plain/styled `Wrap` cases. The expected bytes in `wrap-cases.json`
  were generated using B and independently matched with C. This is correctness
  evidence, not a new performance sample.

All end-to-end runs use Glamour v2.0.1, Goldmark v1.7.17, Chroma v2.14.0,
x/ansi v0.11.9, and the same selected graph except the Lip Gloss pin.
W uses the standalone Lip Gloss module's graph, including x/ansi v0.11.8.

## Measurement environment

Original performance samples were collected on 2026-10-10 with Go 1.26.2,
macOS 26.6.2 (25G83), darwin/arm64, Apple M4 Pro, 14 logical CPUs, 48 GiB RAM,
and default GOMAXPROCS=14. TERM=xterm-256color, COLORTERM=truecolor, NO_COLOR=1;
CLICOLOR and CLICOLOR_FORCE were unset. Runs were serial, with tests and profiles
separate from uninstrumented measurements. Theme and setup/timing boundaries
are described in [the experiment report](../../WRITER-BATCHING.md).

Absolute elapsed times and allocation totals describe these workloads on this
machine and graph. For another machine, collect its own paired baseline and
candidate runs and record the environment. Reproduction is possible across
machines; the original absolute times are not portable performance thresholds.

## Artifact-to-command mapping

Run Go commands from the indicated module root. Redirect stdout to the named
file. For hash output, sort stdout before comparison. Copy archived `.go.txt`
files to temporary `.go` files before execution; the archived harness is not a
second copy of the living CI test.

| Artifact | Producing code | Command |
|---|---|---|
| `bench-baseline-completed.txt` | B | E1 |
| `bench-candidate-completed.txt` | C | E1 |
| `bench-baseline-streaming.txt` | B | E2 |
| `bench-candidate-streaming.txt` | C | E3 |
| `bench-candidate-assistant.txt` | C | E4 |
| `bench-wrap-writer.txt` | W | E5 |
| `profile-candidate-run.txt` | C | P1 |
| `profile-candidate-cpu.txt` | C, P1 CPU profile | P2 |
| `profile-candidate-cpu-cumulative.txt` | C, P1 CPU profile | P3 |
| `profile-candidate-alloc-space.txt` | C, P1 memory profile | P4 |
| `profile-candidate-alloc-objects.txt` | C, P1 memory profile | P5 |
| `render-baseline-sha256.txt` | B + H | H1 |
| `render-candidate-sha256.txt` | C + H | H1 |
| `render_hashes.go.txt` | H, archived in `5d56ab6` | Generator source, not a result |
| `wrap_batch_test.go.txt` | W, archived in `5d56ab6` | Historical harness source; first-line marker points to living CI coverage |
| `wrap_cases.go.txt` | R | Direct Wrap generator source, not a result |
| `wrap-cases.json` | B + R, independently compared with C + R | R1 |

```sh
# E1
go test ./agentic/markdown -run '^$' -bench '^BenchmarkCompleted$' \
  -benchmem -benchtime=200ms -count=3
# E2
go test ./agentic/markdown -run '^$' \
  -bench '^BenchmarkStreaming/Renderer/structured/32768/bytes32/w80$' \
  -benchmem -benchtime=1x -count=3
# E3
go test ./agentic/markdown -run '^$' \
  -bench '^BenchmarkStreaming$/^Renderer$/(prose|structured)/(2048|8192|32768)/(bytes32|bytes256)/(w40|w80)$' \
  -benchmem -benchtime=1x -count=3
# E4
go test ./agentic/markdown -run '^$' \
  -bench '^BenchmarkStreaming/Assistant/structured/32768/bytes32/w80$' \
  -benchmem -benchtime=1x -count=3
# E5 (inside W's Lip Gloss module)
go test -run '^TestWrapBatch' -bench '^BenchmarkWrapBatch$' \
  -benchmem -benchtime=300ms -count=5

# P1 (inside C; stdout becomes profile-candidate-run.txt)
profile_dir=$(mktemp -d)
go test -c -o "$profile_dir/markdown.test" ./agentic/markdown
"$profile_dir/markdown.test" -test.run '^$' \
  -test.bench '^BenchmarkStreaming/Renderer/structured/32768/bytes32/w80$' \
  -test.benchtime=1x -test.count=1 -test.benchmem \
  -test.cpuprofile "$profile_dir/candidate.cpu" \
  -test.memprofile "$profile_dir/candidate.mem"
# P2
go tool pprof -top -nodecount=25 "$profile_dir/markdown.test" "$profile_dir/candidate.cpu"
# P3
go tool pprof -top -cum -nodefraction=0 -nodecount=10000 \
  "$profile_dir/markdown.test" "$profile_dir/candidate.cpu"
# P4
go tool pprof -top -alloc_space -nodecount=25 \
  "$profile_dir/markdown.test" "$profile_dir/candidate.mem"
# P5
go tool pprof -top -alloc_objects -nodecount=25 \
  "$profile_dir/markdown.test" "$profile_dir/candidate.mem"

# H1 (separately inside B and C, pointing to the same fixture root)
go run /path/to/render_hashes.go /path/to/fixture-root | sort
# R1 (separately inside B and C)
go run /path/to/wrap_cases.go
```

The broader comparisons in the report also reference the earlier #33 artifacts
under `../streaming/`. Their code was based on
`c05ccf548162f861736844f729c18104949c90e5` plus the investigation tests;
`079903b` archives the reproducible investigation. See
[STREAMING-PERFORMANCE.md](../../STREAMING-PERFORMANCE.md) for that dataset's
commands and environment. Those historical samples are explicitly distinguished
from the fresh B/C comparisons above.
