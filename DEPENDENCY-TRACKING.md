# Dependency Tracking Report

Generated: 2026-07-17

Tracking upstream API changes in tuiweave's core dependencies that may require
deliberate upgrades to tuiweave.

This is a dated operational snapshot, not the canonical compatibility policy.
[COMPATIBILITY.md](COMPATIBILITY.md) is authoritative; DESIGN.md §6 records the
architectural rationale. The upstream items below are selected for their
relevance to tuiweave and are not an exhaustive roadmap.

---

## Current Pins

| Dependency | Version | Minimum Go |
|---|---|---|
| `charm.land/bubbletea/v2` | `v2.0.8` (tagged) | 1.25.0 |
| `charm.land/lipgloss/v2` | `v2.0.5` (tagged) | 1.25.0 |
| `github.com/charmbracelet/ultraviolet` | `v0.0.0-20260703014108-f5a850f9c2b7` (pseudo) | 1.25.0 |
| `charm.land/glamour/v2` | `v2.0.1` (tagged) | **1.25.8** ← compatibility floor |
| `github.com/charmbracelet/x/ansi` | `v0.11.7` (tagged) | 1.24.2 |

The Go 1.25.8 floor is set by `glamour/v2` v2.0.1. Bubble Tea, Lip Gloss, and
Ultraviolet declare 1.25.0; `x/ansi` declares 1.24.2. Lowering the floor
requires an upstream or module-structure change (see DESIGN.md §6).

---

## Stability policy

> `charm.land/bubbletea/v2` (v2.0.8), `charm.land/lipgloss/v2` (v2.0.5) —
> **beta**; pinned, upgraded deliberately.
>
> `github.com/charmbracelet/ultraviolet` — **v0**, pseudo-versioned; expect
> churn; confined to `layout` (geometry), `overlay` (compositing), and
> `snaptest` (cell-grid parsing). Nothing else imports it.

tuiweave treats Bubble Tea and Lip Gloss as beta dependencies even though they
publish v2 tags. Ultraviolet is explicitly v0 and pseudo-versioned. `x/ansi`
is a tagged utility dependency, but its parsing and visible-width behavior is
observable throughout tuiweave. All upgrades are deliberate, not automatic;
see [COMPATIBILITY.md](COMPATIBILITY.md) for the normative policy.

---

## 1. Bubble Tea v2 — `charm.land/bubbletea/v2`

**Status:** Beta and deliberately pinned (per DESIGN.md).

### Selected open PRs to track

| PR | State | What | Why tuiweave cares |
|---|---|---|---|
| [#1693](https://github.com/charmbracelet/bubbletea/pull/1693) | **OPEN** | WASM support (`GOARCH=wasm`) | If merged, tuiweave's `layout` package (imports ultraviolet) may need build-tag guards for WASM targets. |
| [#1722](https://github.com/charmbracelet/bubbletea/pull/1722) | **OPEN** | `WithHardTabs(false)` / `WithBackspace(false)` ProgramOptions | The cursed renderer's hard-tab optimization can corrupt space-padded column alignment in `View()` output. If tuiweave apps hit alignment bugs on macOS (where `TAB0` is default), this option is the fix. |
| [#1725](https://github.com/charmbracelet/bubbletea/pull/1725) | **OPEN** (conflicts) | Scroll-optimized flush in cursedRenderer | Performance improvement for scroll-heavy views. |
| [#1718](https://github.com/charmbracelet/bubbletea/pull/1718) | **OPEN** | Fall back to 80×24 when initial GetSize returns 0 | Fixes blank first frame under tmux. |
| [#1714](https://github.com/charmbracelet/bubbletea/pull/1714) | **OPEN** | Downsample `insertAbove` content to color profile | Fixes `tea.Println` bypassing `NO_COLOR` ([#1709](https://github.com/charmbracelet/bubbletea/issues/1709)). |

### Merged (post-v2.0.0)

| Change | PR | Notes |
|---|---|---|
| Extended keyboard enhancements (Kitty protocol) | [#1626](https://github.com/charmbracelet/bubbletea/pull/1626) | New `KeyboardEnhancements` field on `tea.View`, new `KeyboardEnhancementsMsg`. tuiweave could eventually leverage `KeyReleaseMsg` / `IsRepeat`. |

### Key open issues

| Issue | Summary |
|---|---|
| [#1590](https://github.com/charmbracelet/bubbletea/issues/1590) | Garbage chars printed when program exits too early. Affects Ghostty, WezTerm, Windows Terminal. Workaround exists, no fix merged. |
| [#1564](https://github.com/charmbracelet/bubbletea/issues/1564) | Renderer mangles output on some terminals (CRLF/newline). Related to ultraviolet PR [#133](https://github.com/charmbracelet/ultraviolet/pull/133). |
| [#1541](https://github.com/charmbracelet/bubbletea/issues/1541) | Tab characters not preserved in v2. |
| [#1571](https://github.com/charmbracelet/bubbletea/issues/1571) | Scrollback lost on `tea.Quit`. |

---

## 2. Lipgloss v2 — `charm.land/lipgloss/v2`

**Status:** Beta and deliberately pinned (per DESIGN.md).

### Milestones to watch

| Milestone | Status | What's tracked |
|---|---|---|
| **Border title** ([#4](https://github.com/charmbracelet/lipgloss/milestone/4)) | 1/6 closed | Four competing PRs (#97, #284, #316, #384) with **no clear frontrunner**. Approaches range from simple `BorderTitle` methods to a full `Borderer` interface with `Region` abstraction. tuiweave's `frame.Panel` already implements its own title rendering — don't depend on any particular shape yet. |
| **Tree upgrades** ([#5](https://github.com/charmbracelet/lipgloss/milestone/5)) | 1/2 closed | PR [#461](https://github.com/charmbracelet/lipgloss/pull/461) would add child mutation after creation. Low relevance unless tuiweave adds a tree view. |

### Selected open PRs to track

| PR | State | What | Why tuiweave cares |
|---|---|---|---|
| [#695](https://github.com/charmbracelet/lipgloss/pull/695) | **OPEN** | Multi-character corners on borders | If merged, tuiweave's `frame.Panel` could define custom borders with multi-character corners like `──`. |
| [#711](https://github.com/charmbracelet/lipgloss/pull/711) | **OPEN** | Terminate margin background on each line | Fixes background bleed in styled margins. Relevant to `frame.Panel` background rendering. |
| [#710](https://github.com/charmbracelet/lipgloss/pull/710) | **OPEN** | Don't overshoot width with wide whitespace | Fixes width calculation with wide chars. Relevant to `line` package. |
| [#708](https://github.com/charmbracelet/lipgloss/pull/708) | **OPEN** | Don't mangle embedded ANSI when styling spaces | Important for any component that embeds ANSI in styled content. |
| [#707](https://github.com/charmbracelet/lipgloss/pull/707) | **OPEN** | Preserve word separation in inline multiline render | Rendering correctness for inline mode. |
| [#700](https://github.com/charmbracelet/lipgloss/pull/700) | **OPEN** | Fix border size getters without sides | Affects `GetHorizontalFrameSize()` — relevant to `frame.PanelContentRect`. |
| [#697](https://github.com/charmbracelet/lipgloss/pull/697) | **OPEN** | Table `ContentWidth` property | Draft. tuiweave has its own table component. Low relevance. |
| [#680](https://github.com/charmbracelet/lipgloss/pull/680) | **OPEN** | Collapse newlines to spaces in inline mode | Rendering correctness for inline mode. |

### Community proposals (no official commitment)

| Proposal | What | Relevance to tuiweave |
|---|---|---|
| [#643](https://github.com/charmbracelet/lipgloss/issues/643) | Unified theming system (`charm-themes`) | tuiweave already has its own `Theme` system with semantic roles. If upstream adopts this, tuiweave would need a mapping layer. |
| [#644](https://github.com/charmbracelet/lipgloss/issues/644) | Flexbox/grid layout engine (`charm-layout`) | Overlaps significantly with tuiweave's `layout` package. Low priority — no official movement. |

### Known upstream wart

`Style.Width()` includes borders/padding, and `GetHorizontalFrameSize()` also
includes padding. tuiweave's `frame.PanelContentRect` already works around this
correctly. If upstream adds `ContentWidth()` or `OverallWidth()`, tuiweave
could simplify its panel math.

---

## 3. Ultraviolet — `github.com/charmbracelet/ultraviolet`

**Status:** v0, pseudo-versioned (per DESIGN.md). **Zero tagged releases.**
Breaking changes can land without notice. Currently pinned at commit
`f5a850f9c2b7`. A newer pseudo-version at `4bee1914c0cf` exists but that alone
is not a reason to upgrade.

### Where tuiweave touches ultraviolet

Per AGENTS.md rule #2, exactly three packages:

| Package | Import | What it uses |
|---|---|---|
| `layout/layout.go` | `uvlayout` | Constraint solver and provisional public constraint identity |
| `overlay/overlay.go` | `uv` | Cell-level overlay composition |
| `snaptest/cells.go` | `uv` | Style-run golden testing |

### Selected open PRs to track

All of these are **open** — none are merged. If and when they merge, they
would benefit tuiweave upon deliberate upgrade.

| PR | What | Why tuiweave cares |
|---|---|---|
| [#143](https://github.com/charmbracelet/ultraviolet/pull/143) | Repaint unchanged rows inside hard-scrolled regions | Rendering correctness fix. |
| [#142](https://github.com/charmbracelet/ultraviolet/pull/142) | Rotate lines instead of copying cells when scrolling | Performance optimization (31×–95× faster scroll). |
| [#140](https://github.com/charmbracelet/ultraviolet/pull/140) | Don't split wide cell when deleting preceding narrow cell | CJK/emoji rendering fix. |
| [#138](https://github.com/charmbracelet/ultraviolet/pull/138) | Cancel reads on any console handle, not just stdin | Input reliability fix. |
| [#136](https://github.com/charmbracelet/ultraviolet/pull/136) | Legacy Linux console F1-F5 keys | Keyboard input fix. |
| [#135](https://github.com/charmbracelet/ultraviolet/pull/135) | Clamp stale line areas to buffer bounds | Panic prevention on resize. |
| [#134](https://github.com/charmbracelet/ultraviolet/pull/134) | Guard bounds check in `printString` | Panic prevention. |
| [#133](https://github.com/charmbracelet/ultraviolet/pull/133) | Emit CRLF for downward cursor moves in inline mode | Fixes rendering corruption on some terminals. |
| [#131](https://github.com/charmbracelet/ultraviolet/pull/131) | Keep columns aligned on terminals without mode 2027 | Fixes rendering on Apple Terminal.app. |
| [#130](https://github.com/charmbracelet/ultraviolet/pull/130) | `DrawOver` and `HardScroll` for incremental scroll rendering | New public APIs. tuiweave doesn't use `TerminalRenderer` directly. |
| [#128](https://github.com/charmbracelet/ultraviolet/pull/128) | Flush trailing empty cells in renderLine | Rendering correctness. |
| [#122](https://github.com/charmbracelet/ultraviolet/pull/122) | Shift+non-English letter parsed as BaseCode | Keyboard input fix for international layouts. |
| [#120](https://github.com/charmbracelet/ultraviolet/pull/120) | Replace Windows input busy-poll with WaitForSingleObject | Windows CPU usage fix. |
| [#119](https://github.com/charmbracelet/ultraviolet/pull/119) | Rotate line slices in Insert/DeleteLineArea fast path | Performance optimization. |
| [#116](https://github.com/charmbracelet/ultraviolet/pull/116) | ParseOsc correctly reports ClipboardEvent selection | Clipboard handling fix. |
| [#100](https://github.com/charmbracelet/ultraviolet/pull/100) | `DisableCaps()` for individual terminal cap control | New API for terminal capability control. |
| [#76](https://github.com/charmbracelet/ultraviolet/pull/76) | Don't draw cell when border Side content is empty | Border rendering fix. |
| [#75](https://github.com/charmbracelet/ultraviolet/pull/75) | Preclude stale screen content in full-screen mode | Screen corruption fix. |

### Key concern: layout package coupling

The `layout` package imports `uvlayout` (ultraviolet's layout solver). If
ultraviolet changes its layout API, tuiweave's `layout` package breaks. The
`uvlayout` API appears stable (no recent changes), but there is no stability
contract.

---

## 4. ANSI utilities — `github.com/charmbracelet/x/ansi`

**Status:** Tagged direct dependency, currently pinned at `v0.11.7`.

tuiweave uses `x/ansi` throughout rendering code and tests for ANSI-safe
measurement, truncation, wrapping, stripping, and cell-width behavior. Changes
can affect exact-width component contracts and golden output even when the API
remains source-compatible. Review wide-character, combining-grapheme, embedded
ANSI, and narrow-width coverage on every upgrade.

---

## Upgrade Validation Procedure

When upgrading any dependency, run the full validation suite:

```bash
# 1. Targeted tests for ultraviolet-touching packages
go test ./layout/... ./overlay/... ./snaptest/...

# 2. Full build and vet
go build ./...
go vet ./...

# 3. Full test suite, then repeat with the race detector
go test ./...
go test -race ./...

# 4. Module consistency
go mod tidy -diff
go mod verify

# 5. Golden file review
# After intentional visual changes:
go test ./... -update
git diff  # review golden diffs carefully
```

Do not skip steps. Regenerate goldens only for intentional visual changes —
never to silence a failure you don't understand. If an upgrade intentionally
requires `go mod tidy`, review the resulting `go.mod` and `go.sum` changes,
then run `go mod tidy -diff` as the clean-state validation.

---

## Summary: Action Items

| Priority | Item | Dependency | Action |
|---|---|---|---|
| 🔴 High | Monitor ultraviolet pseudo-version bumps | Ultraviolet | Run full validation procedure before bumping. No API stability contract. |
| 🟡 Medium | Track PR #1722 (WithHardTabs) | Bubble Tea | Investigate only if tuiweave reproduces a column-alignment failure; if merged, document the option as a targeted workaround. |
| 🟡 Medium | Track ANSI width behavior | x/ansi | Review wide, combining, embedded-ANSI, and narrow-width coverage before bumping. |
| 🟢 Low | Border title milestone | Lipgloss | If upstream ships a native API, evaluate it only when it materially simplifies `frame.Panel`. |
| 🟢 Low | Keyboard enhancements | Bubble Tea | Consider `KeyReleaseMsg` or `IsRepeat` awareness only for a concrete interaction requirement. |
| 🟢 Low | WASM support (PR #1693) | Bubble Tea | If WASM becomes a supported target, verify `layout` and the full module under the relevant `GOOS`/`GOARCH`. |
| 🟢 Low | Unified theming proposal (#643) | Lipgloss | Monitor only. tuiweave's own Theme system is mature. |
| 🟢 Low | Flexbox/grid proposal (#644) | Lipgloss | Monitor only. Overlaps with tuiweave's layout package. |
| 🟢 Low | DevTools/testing proposals (#1654/#1655) | Bubble Tea | tuiweave's `snaptest` already covers this territory. |
| 🟢 Low | Multi-character corners (#695) | Lipgloss | Evaluate after merge and a deliberate dependency upgrade. |

---

## Methodology

This report was generated by:

1. Querying GitHub via `gh` CLI for open/merged PRs, issues, milestones, and
   releases on each repository. Every PR's merge status was verified with
   `gh pr view <number> --json state,mergedAt` — not inferred from issue
   comments or subagent summaries.
2. Reading DESIGN.md §6 for the project's own dependency policy and stability
   classification.
3. Cross-referencing findings against tuiweave's actual imports and usage
   patterns in `go.mod` and source files.

Next review recommended: monthly cadence, or when bumping any dependency
version.
