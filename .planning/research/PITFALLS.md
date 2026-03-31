# Pitfalls Research

**Domain:** Go text diff library + CLI (Myers/Patience/Histogram algorithms, Chroma highlighting, Lip Gloss layout)
**Researched:** 2026-03-25
**Confidence:** HIGH (all critical pitfalls verified against official issues, open-source post-mortems, and library documentation)

---

## Critical Pitfalls

### Pitfall 1: Myers Backtrack Off-by-One (trace array save timing)

**What goes wrong:**
The Myers diff implementation produces correct diffs on short inputs but generates shifted/wrong insertion and deletion positions on longer inputs. The algorithm "works" in unit tests but silently produces incorrect output in production.

**Why it happens:**
Myers' algorithm stores a `trace` array of V-states for backtracking. The classic mistake is saving the V-state at the **top** of the `d` loop iteration rather than at the **end** — one step too early. Short test cases (≤5 lines) typically never trigger the condition that exposes this because all states are traversed in the first few iterations.

**How to avoid:**
- Study the original Eugene Myers 1986 paper before implementing — the algorithm is only ~40 lines of logic but the data structure semantics are non-obvious.
- Validate against Git's own diff output on large, real-world files (500+ lines) as part of the implementation test suite.
- Use table-driven tests with known-correct edit distances from the Myers paper.
- Consider using an established Go implementation (`github.com/sergi/go-diff` or `github.com/pmezard/go-difflib`) as a correctness oracle to cross-validate output during testing.

**Warning signs:**
- Tests pass on 5–10 line inputs but fail (or visually look wrong) on 100+ line files.
- Edit positions in the output are systematically offset by 1.
- `diff -u` from the system shell produces different output than your library for the same inputs.

**Phase to address:** Algorithm implementation phase (Phase 1 / core engine)

---

### Pitfall 2: Histogram Fallback to Myers Not Implemented

**What goes wrong:**
The Histogram algorithm silently produces wrong or degenerate output when no lines appear fewer than the occurrence threshold (typically 65 in jgit, sometimes 512). The algorithm requires a Myers fallback for this case — without it, high-repetition files (minified code, CSV data, log files with repeated lines) will either panic, produce empty diffs, or return garbage.

**Why it happens:**
Implementors focus on the "happy path" of the histogram algorithm and skip the fallback. The jgit source — the canonical reference — has confusing comments that obscure this requirement. The fallback only triggers on pathological inputs that don't appear in typical unit tests.

**How to avoid:**
- Read the jgit `HistogramDiff.java` source and the analysis at https://raygard.github.io/2025/01/28/how-histogram-diff-works/ before implementing.
- Explicitly test the fallback: create an input where every line appears 66+ times and verify the output is correct.
- Add an internal counter that logs/tracks fallback invocations during testing.

**Warning signs:**
- Histogram diff produces empty or trivially wrong output on files with many repeated identical lines.
- Minified JSON, CSV, or log files cause incorrect diffs.
- No test cases exercise files with >65 occurrences of any single line.

**Phase to address:** Algorithm implementation phase (Phase 1 / core engine)

---

### Pitfall 3: Patience Diff Requires a Myers Fallback Between Unique-Line Anchors

**What goes wrong:**
Patience diff only computes LCS on *unique* lines. When there are no unique matching lines in a region (or very few), the algorithm falls back to nothing and produces an empty/incorrect diff for that region. The output looks correct overall but misses changes in repetitive sections.

**Why it happens:**
The Patience algorithm is designed as a "first pass" that identifies unique-line anchors, then recursively processes the regions between anchors — those inter-anchor regions need a fallback algorithm (typically Myers). Implementors often forget the recursive inter-anchor step or skip the fallback.

**How to avoid:**
- Test Patience diff against files with no unique lines (e.g., a file consisting of only `{` and `}` characters).
- Implement the inter-anchor recursive step explicitly and verify it with tests.
- Use `github.com/peter-evans/patience` as a reference implementation.

**Warning signs:**
- Patience diff produces shorter output than Myers on the same input (missing hunks, not just better-organized ones).
- Regions of highly repeated code (boilerplate, closing braces) are not diffed at all.

**Phase to address:** Algorithm implementation phase (Phase 1 / core engine)

---

### Pitfall 4: Lip Gloss Width/Border Size Mismatch in Side-by-Side Layout

**What goes wrong:**
When constructing side-by-side (split) diff panels with Lip Gloss, the rendered width of bordered panels does not match the width set on the style. The actual rendered width is `setWidth + borderWidth` (typically +2), causing the two panels to overflow the terminal width, wrap incorrectly, or be misaligned.

**Why it happens:**
In Lip Gloss v1.x, `Style.Width()` set the inner content width, not the outer frame width — borders, padding, and margin were additive on top. This was a known bug fixed in Lip Gloss v2.0.0. The fix (released 2025-02-10) changed `Width()` to encompass the entire styled block including borders. However, code written for v1 that uses `GetHorizontalFrameSize()` to compensate may now double-subtract in v2.

**How to avoid:**
- Pin to Lip Gloss v2+ (`charm.land/lipgloss/v2` or `github.com/charmbracelet/lipgloss` v2.x) and use the corrected behavior.
- Always compute panel content width as: `contentWidth = (termWidth / 2) - style.GetHorizontalFrameSize()`, then set `style.Width(contentWidth)`.
- Add an integration test that renders two side-by-side panels and asserts `lipgloss.Width(leftPanel) + lipgloss.Width(rightPanel) == termWidth`.
- Never add border widths manually when using v2+ — the library accounts for them.

**Warning signs:**
- Side-by-side panels render slightly wider than the terminal, causing line wrap artifacts.
- `lipgloss.Width(renderedPanel) != expectedWidth` in tests.
- Off-by-two rendering when borders are enabled.

**Phase to address:** Rendering/layout phase (side-by-side view implementation)

---

### Pitfall 5: ANSI Escape Codes Break String Width Measurement

**What goes wrong:**
When lines have already been Chroma-highlighted (wrapped in ANSI escape codes for colors), passing them to `lipgloss.Width()` or `runewidth.StringWidth()` returns an inflated width because the escape sequences are counted as printable characters. This causes padding calculations to be wrong, truncation to cut visible content, and side-by-side alignment to collapse.

**Why it happens:**
ANSI escape sequences (`\x1b[31m`, etc.) are zero-width control codes, not printable characters, but naive string length functions treat them as bytes/runes with width. Lip Gloss has its own ANSI-aware width function, but it is not automatically used when constructing layouts from pre-rendered strings.

**How to avoid:**
- Always use `lipgloss.Width()` (not `len()`, `utf8.RuneCountInString()`, or `runewidth.StringWidth()`) to measure strings that may contain ANSI codes.
- Apply Chroma highlighting *after* computing layout widths, not before; or strip ANSI codes before width measurement.
- Use `github.com/charmbracelet/x/ansi` for ANSI-aware string operations when building custom layout logic.
- Test rendering with pre-highlighted strings to catch miscalculated padding.

**Warning signs:**
- Side-by-side panels render correctly for plain text but misalign on highlighted code.
- String padding is shorter than expected when color is enabled.
- `lipgloss.Width(chromaHighlightedLine) > lipgloss.Width(plainLine)` for the same logical content.

**Phase to address:** Rendering/layout phase (Chroma + Lip Gloss integration)

---

### Pitfall 6: Unicode Wide Character (East Asian / Emoji) Column Misalignment

**What goes wrong:**
Diff output for files containing Chinese, Japanese, Korean, or emoji characters renders with misaligned columns in side-by-side view. Characters visually take 2 terminal columns but are measured as 1, causing the right panel to shift left progressively as wide characters accumulate.

**Why it happens:**
`mattn/go-runewidth` (used internally by Lip Gloss) measures character widths using East Asian Width Unicode tables. The behavior depends on the `EastAsian` condition setting — when `LANG=zh_CN.UTF-8` or similar locales are set, box-drawing characters used for borders are returned as width 2 instead of 1. Additionally, `runewidth.StringWidth()` does not correctly handle multi-codepoint sequences like emoji with zero-width joiners (ZWJ sequences like `🏳️‍🌈` may return 6 instead of 2).

**How to avoid:**
- Explicitly set `runewidth.DefaultCondition.EastAsian = false` (or set `RUNEWIDTH_EASTASIAN=0` env var) at startup to get consistent half-width box-drawing characters.
- For emoji-heavy content, consider using `github.com/rivo/uniseg` for grapheme cluster-aware width measurement instead of `go-runewidth`.
- Test with files containing CJK characters, emoji, and box-drawing characters.
- Accept that perfect alignment for all Unicode inputs in all terminals is impossible — document the limitation.

**Warning signs:**
- Side-by-side diff is misaligned when the source files contain non-ASCII characters.
- Box-drawing border characters are rendered double-wide in some locales.
- Tests pass on CI (ASCII test data) but users report alignment issues with their actual files.

**Phase to address:** Rendering/layout phase (Unicode handling)

---

### Pitfall 7: Terminal Width Detection Fails When Output Is Piped

**What goes wrong:**
`os.Stdout.Fd()` returns a valid file descriptor when running interactively, but when the user pipes `drift file1 file2 | less` or redirects to a file, the terminal width query fails or returns 0, causing the side-by-side layout to either panic, use a hardcoded width, or render at 0 columns.

**Why it happens:**
`term.GetSize(int(os.Stdout.Fd()))` only works when stdout is a TTY. When stdout is a pipe, it returns an error. Many implementations ignore the error return and use the zero-value width. Side-by-side mode is meaningless when piped (no interactive terminal), but the code must still handle the case gracefully.

**How to avoid:**
- Check `isatty.IsTerminal(os.Stdout.Fd())` before attempting terminal width detection.
- Always provide a fallback width (e.g., 80 columns) when the terminal size cannot be determined.
- For piped output, automatically fall back to unified diff mode (not side-by-side) — this is the same behavior as `git diff`.
- Never use a width of 0 or a negative width in layout calculations.

**Warning signs:**
- `drift file1 file2 | less` panics or produces garbled output.
- Tests using `os.Pipe()` to capture output fail with layout panics.
- Side-by-side mode produces zero-width or 1-pixel panels when piped.

**Phase to address:** CLI implementation phase

---

### Pitfall 8: Color Support Detection — Writing ANSI to Non-TTY Outputs

**What goes wrong:**
The library writes ANSI escape codes to stdout even when the output is a file, pipe, or CI system that does not support them. Users see literal `\x1b[31m` garbage instead of colored text. Alternatively, the library disables color when running in CI systems like GitHub Actions (which do support ANSI via `FORCE_COLOR`).

**Why it happens:**
Color support detection based solely on `isatty` is too conservative — it disables colors in CI environments that support ANSI output. Detection based on `$TERM` alone is insufficient on Windows (conhost.exe has ANSI disabled by default even on modern Windows 10). The `NO_COLOR` standard (no-color.org) is also commonly ignored.

**How to avoid:**
- Use `github.com/charmbracelet/colorprofile` (the Charmbracelet ecosystem's canonical color detection, used internally by Lip Gloss v2) — it handles `NO_COLOR`, `FORCE_COLOR`, `COLORTERM`, `$TERM`, isatty, and Windows VT processing in one call.
- Respect the `NO_COLOR` env var (disable all color output when set).
- Respect `FORCE_COLOR` (enable color even in non-TTY contexts).
- Do not write ANSI to stderr separately from stdout — detect each independently.
- On Windows, call `termenv.EnableVirtualTerminalProcessing()` at startup.

**Warning signs:**
- `drift file1 file2 > output.txt` produces a file with ANSI escape codes.
- Color output broken in GitHub Actions CI.
- Users on Windows CMD.exe see garbled `[31m` text.

**Phase to address:** CLI implementation phase (color detection)

---

### Pitfall 9: Terminal Dark/Light Theme Detection Hangs or Returns Wrong Answer

**What goes wrong:**
Auto-detecting light vs. dark terminal background (to select the appropriate Chroma theme) either hangs for 1–2 seconds while waiting for a terminal response, returns an incorrect result, or fails entirely in non-xterm terminals. OSC 11 queries (`\e]11;?\a`) hang in some terminals that don't respond to the query.

**Why it happens:**
OSC 11/10 terminal color queries require reading a response from `/dev/tty` with a timeout. Some terminals (PuTTY, older Windows Terminal, VS Code's integrated terminal, tmux without `set -g allow-passthrough on`) don't support OSC queries at all and never respond. Without a read timeout, the query blocks indefinitely. `$COLORFGBG` is only set by rxvt-derived terminals and is absent in most modern terminals.

**How to avoid:**
- Use `github.com/muesli/termenv` (`output.HasDarkBackground()`) which queries OSC 11 with a short timeout and falls back gracefully.
- Always set a read timeout (50–200ms) on the `/dev/tty` read.
- Provide `--theme` / `--light` / `--dark` CLI flags so users can override auto-detection.
- Default to a dark theme when detection fails (most developer terminals are dark).
- Test with tmux — it requires special passthrough configuration for OSC queries to work.

**Warning signs:**
- CLI hangs for 1–2 seconds on startup before printing output.
- Auto-detected theme is wrong when tmux is active.
- Tests that run under a CI environment (no TTY) hang or timeout.

**Phase to address:** CLI implementation phase (theme auto-detection)

---

### Pitfall 10: Chroma Lexer v1 vs v2 Import Path Confusion

**What goes wrong:**
Code imports `github.com/alecthomas/chroma` (v1) while the current release is `github.com/alecthomas/chroma/v2`. The v1 API has breaking differences: the `lexers` package path is different, style lookup API differs, and some formatters (especially the ANSI terminal formatter) changed their interface. Mixing v1 and v2 import paths in the same module causes compile errors.

**Why it happens:**
Many online tutorials and Stack Overflow answers still show the v1 import path. The v1 package remains importable at its original path. Both compile without error until a type mismatch is caught at interface boundaries.

**How to avoid:**
- Always import `github.com/alecthomas/chroma/v2` — not the bare path.
- Use `lexers.Get("go")` from `github.com/alecthomas/chroma/v2/lexers` for language lookup.
- When a lexer returns `nil` (language not found), use `lexers.Fallback` (plain text) — never pass a nil lexer to the tokenizer.
- Test with an unknown file extension to exercise the nil-lexer fallback path.

**Warning signs:**
- `cannot use ... (type chroma.Lexer) as type chroma.Lexer` compile errors (two different versions of the same interface in the build graph).
- `nil pointer dereference` at lexer tokenization when an unsupported language is passed.
- Missing styles that were added in chroma v2.14+ if pinned to an old version.

**Phase to address:** Chroma integration phase

---

### Pitfall 11: Chroma ANSI Formatter Outputs 256-Color Codes by Default

**What goes wrong:**
The Chroma terminal formatter outputs 256-color ANSI codes (`\e[38;5;Xm`) by default. On terminals that only support 16 colors (older SSH sessions, Linux console, some CI environments), these codes either render as wrong colors or display as garbled text. Conversely, on true-color terminals, 256-color output produces noticeably degraded color fidelity compared to true-color output.

**Why it happens:**
Chroma has three terminal formatters: `terminal` (8-color ANSI), `terminal256` (256-color), and `terminal16m` (true-color/24-bit). The `quick.Highlight` convenience function uses `terminal` (8-color) by default — but implementors often manually select `terminal256` without checking what the terminal actually supports.

**How to avoid:**
- Detect terminal color depth with `github.com/charmbracelet/colorprofile` and select the appropriate Chroma formatter: `terminal16m` for TrueColor, `terminal256` for ANSI256, `terminal` for basic 16-color, `noop` for no-color.
- Never hardcode `terminal256` — always query the color profile at runtime.
- Test with `COLORTERM=` (unset) and `TERM=xterm` to exercise the 16-color path.

**Warning signs:**
- Color output looks dull or wrong on older terminal emulators.
- `TERM=xterm-256color` is being used as the only color detection check.
- No tests exercise different color depth levels.

**Phase to address:** Chroma integration phase

---

### Pitfall 12: Go API Design — Exported Struct Fields Create Breaking Change Traps

**What goes wrong:**
Exporting concrete struct fields (e.g., `type Options struct { Context int; Algorithm string }`) forces a major version bump for any field addition if users are constructing the struct with positional initialization (`drift.Options{3, "myers"}`). Adding a field breaks all call sites that initialize the struct without field names.

**Why it happens:**
Go's struct literal positional initialization is technically valid but brittle. Library authors export structs without considering how users will initialize them. When a new field is added, all positional initializations fail to compile — this is a breaking change even though Go semver only counts it as "minor" by convention.

**How to avoid:**
- Use functional options pattern (`type Option func(*config)`) for configuring diff calls — this is indefinitely extensible without breaking changes.
- If exporting a struct (e.g., a `Result` type), document that users must use named field initialization.
- Avoid embedding structs in your public API — adding a method to an embedded type creates an ambiguous selector, a silent breaking change.
- Minimize exported surface area: prefer `internal/` for implementation details.

**Warning signs:**
- Users initializing `drift.Options{3, "myers"}` without field names.
- Any exported struct that will need new fields after v1.0.
- Embedded structs in exported types.

**Phase to address:** API design phase (before v1.0.0 stabilization)

---

### Pitfall 13: Go Module Major Version in Import Path Not Updated

**What goes wrong:**
When the library releases v2.0.0 with breaking changes, the `go.mod` module path is not updated to `github.com/tbcrawford/drift/v2`. Users who `go get github.com/tbcrawford/drift@v2.0.0` get a confusing error — the module system treats the old and new paths as different modules, and tooling like `go get` will not automatically resolve to the v2 path.

**Why it happens:**
Go module major versioning requires updating the import path for v2+. This is a Go-specific convention that catches many library authors off-guard — the Go module system treats `github.com/foo/bar` and `github.com/foo/bar/v2` as completely separate modules. Forgetting this means breaking changes ship under the same import path, violating the Go module contract.

**How to avoid:**
- Commit to the current pre-v1.0 period to stabilize the API — avoid releasing v1.0.0 until the public API is solid.
- When breaking changes are needed post-v1.0, bump the `module` line in `go.mod` to `github.com/tbcrawford/drift/v2` and update all internal imports.
- Use `go.mod`'s `module` directive correctly from day one: `module github.com/tbcrawford/drift`.
- Pin to `v0.x` as long as the API is still evolving — the Go community understands `v0` means "no stability guarantee."

**Warning signs:**
- `go get github.com/tbcrawford/drift@v2.0.0` fails with "unknown revision" errors.
- Internal imports still use the old path after a major version bump.
- The CHANGELOG doesn't distinguish breaking vs. non-breaking changes.

**Phase to address:** API design phase + ongoing versioning discipline

---

### Pitfall 14: Large File Performance — O(N²) Memory in Diff Algorithm

**What goes wrong:**
Diffing large files (>10,000 lines) causes excessive memory allocation or runs noticeably slow because the Myers algorithm's worst-case memory is O(N²) for the trace array (one V-state snapshot per edit distance `d`). On files with many differences, `d` can approach `N`, making the trace array enormous.

**Why it happens:**
The basic Myers algorithm stores the entire V-state history (`trace[d]`) for backtracking. For a 10,000-line file with 5,000 edits, this means 5,000 × 5,000 = 25M integers. Most implementations don't implement the "linear space refinement" from Myers' original paper (Section 4), which divides-and-conquers to reduce memory to O(N).

**How to avoid:**
- Profile with `go test -benchmem` using 10,000-line synthetic test files before declaring performance acceptable.
- Consider implementing the divide-and-conquer linear-space refinement for the Myers algorithm (or using it only for files above a line-count threshold).
- Use `sync.Pool` to reuse the `V` array (the current diagonal endpoints) across calls — this is the most impactful single optimization for throughput.
- For histogram/patience, pre-allocate the hash table with `make(map[string]int, len(lines))` to avoid rehashing.
- Add a `maxDiffs` / `maxLines` safety limit with a fallback to coarser diff output (as Go's `x/tools/internal/diff` does).

**Warning signs:**
- `go tool pprof` shows diff algorithm taking >50% of CPU on realistic inputs.
- Heap allocations reported by `-benchmem` scale super-linearly with input size.
- Diffing a 10,000-line file takes >500ms.

**Phase to address:** Algorithm implementation phase (performance validation)

---

### Pitfall 15: Testing Diff Output With String Equality on ANSI-Colored Output

**What goes wrong:**
Tests assert exact string equality on rendered diff output. Any change to Chroma themes, ANSI formatter behavior, or Lip Gloss styling causes every test to fail even when the structural diff is correct. This creates a "test churn" problem — fixing a color or style detail requires updating hundreds of golden strings.

**Why it happens:**
The natural first instinct for testing output is `assert.Equal(t, expected, got)` with hardcoded expected strings. For a rendering library, this embeds rendering decisions into tests, making them brittle to any cosmetic change.

**How to avoid:**
- Separate structural tests (correct hunks, correct line numbers, correct change detection) from rendering tests.
- For structural tests: test the `[]Hunk` / `[]Edit` data structures, not the rendered string.
- For rendering tests: use golden files in `testdata/` (Go's idiomatic pattern) with a `-update` flag to regenerate them: `go test -update ./...`.
- For ANSI output tests: strip ANSI escape codes before comparing (`github.com/charmbracelet/x/ansi` provides `Strip()`), then compare plain text structure.
- Use `github.com/sebdah/goldie` or `github.com/matttproud/goldentest` for golden file management.

**Warning signs:**
- Test file contains hundreds of lines of hardcoded ANSI escape sequences.
- Every Chroma theme change requires a `go generate` or manual string replacement sweep.
- No tests that verify diff algorithm correctness independently of rendering.

**Phase to address:** Testing infrastructure (established in Phase 1, maintained throughout)

---

### Pitfall 16: Single `go.mod` — `go install` Path Must Point to `cmd/drift/main.go`

**What goes wrong:**
Users run `go install github.com/tbcrawford/drift@latest` and get an error (`go: github.com/tbcrawford/drift@latest: module github.com/tbcrawford/drift: no Go files in root`) because the root package is a library (`package drift`) and there's no `main` package at the root.

**Why it happens:**
`go install` on a module root requires `package main` at the root directory. A library package at the root with a CLI in `cmd/drift/` is the correct layout — but the install path for the CLI must include the subdirectory: `go install github.com/tbcrawford/drift/cmd/drift@latest`.

**How to avoid:**
- Document the correct install command prominently in the README: `go install github.com/tbcrawford/drift/cmd/drift@latest`.
- Verify the install path works from a clean `GOPATH` as part of the release checklist.
- Never put `package main` at the module root for a library+CLI module.
- The import path for library users (`go get github.com/tbcrawford/drift`) and the install path for CLI users (`go install github.com/tbcrawford/drift/cmd/drift@latest`) are intentionally different — both are correct.

**Warning signs:**
- `go install github.com/tbcrawford/drift@latest` produces "no Go files" or "no main package" errors.
- README shows the wrong install path.
- CI release checks only verify `go build` not `go install`.

**Phase to address:** Module structure setup (Phase 1) + README documentation

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Hardcode `terminal256` formatter | Simpler color setup | Wrong colors on 16-color terminals; CI test failures | Never — query profile at runtime |
| String-equality tests on rendered ANSI output | Fast to write | Every style tweak breaks tests; maintenance nightmare | Never for golden tests; only for structural unit tests |
| Export concrete struct for options | Simple API | Any new field is a breaking change | Only if struct is tagged as "fill by name only" and documented |
| Skip histogram Myers fallback | Faster to implement | Silent wrong diffs on repetitive files | Never — the fallback is a correctness requirement |
| Use `len(s)` for terminal width | No dependencies | Misalignment for any non-ASCII content | Never in layout code |
| Detect color with only `isatty` | Simple one-liner | Breaks in CI, GitHub Actions, Windows CMD | Only acceptable as a `--no-color` fast path |
| Single golden string for full render | One test covers render + structure | Any layout change breaks all tests | Only for smoke/integration tests, not unit tests |
| `strings.Split(text, "\n")` for line splitting | Simple | Loses trailing newline information; `\r\n` files produce `\r` suffix on each line | Never — use a proper line-splitting function that handles `\r\n` and trailing newlines |

---

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Chroma v2 | Importing `github.com/alecthomas/chroma` (v1 path) | Import `github.com/alecthomas/chroma/v2` and `github.com/alecthomas/chroma/v2/lexers` |
| Chroma lexer | Passing `nil` lexer to tokenizer when language unknown | Always check for nil: `if lexer == nil { lexer = lexers.Fallback }` |
| Chroma formatter | Hardcoding `terminal256` formatter | Select formatter based on `colorprofile.Detect()` result |
| Lip Gloss v2 borders | Computing `contentWidth = termWidth - 2` | Use `style.GetHorizontalFrameSize()` which accounts for all frame elements |
| Lip Gloss + Chroma | Measuring ANSI-highlighted string width | Use `lipgloss.Width()` not `len()` or `runewidth.StringWidth()` |
| termenv / colorprofile | Calling `output.HasDarkBackground()` without timeout handling | Wrap in a goroutine with a 200ms timeout; use a default on timeout |
| `term.GetSize()` | Not checking error return when stdout is a pipe | Check `isatty.IsTerminal()` first; provide fallback width (80) |
| `strings.Split(s, "\n")` | Used for line splitting | Use a scanner that preserves whether final newline exists; handle `\r\n` explicitly |

---

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Myers O(N²) trace allocation | High memory, slow for large files | Implement linear-space variant or add maxDiffs limit | Files > 5,000 lines with many changes |
| `strings.Split()` on large file content | 70% of CPU in allocation (`runtime.makeslice`) | Use `bufio.Scanner` or index-based line splitting | Files > 1MB |
| Chroma re-tokenizing same content per render | Repeated lexer work | Cache tokenized token streams; only re-render when needed | Any file > 500 lines |
| Re-allocating V array per diff call | High GC pressure in library usage | Use `sync.Pool` for the reusable V array | High-frequency library callers (LSP, editors) |
| Concatenating string hunks with `+` | O(N²) string concatenation | Use `strings.Builder` or `bytes.Buffer` | Files with >1000 diff hunks |
| Computing `lipgloss.Width()` in a per-line loop on pre-rendered output | Slow render for large diffs | Pre-compute panel widths once; don't re-measure each line | Side-by-side with >500 lines |

---

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| No `--no-color` flag | Users in pipes or minimal terminals see ANSI garbage | Always honor `NO_COLOR` env and provide `--no-color` flag |
| Side-by-side only, no unified fallback | Unusable when piped to `grep`, `less`, `wc` | Auto-detect TTY and fall back to unified mode when not interactive |
| No `--context N` flag | Cannot control hunk context size like `git diff -U3` | Expose context lines as a configurable option (default: 3) |
| Language detection only by file extension | Fails for files without extensions (Makefile, Dockerfile) | Use `lexers.Match(filename)` first, then `lexers.Analyse(content)` for content-based detection |
| Crashing on binary files | Unintuitive failure for users who diff mixed repos | Detect binary content (NUL byte scan); print "Binary files differ" and exit cleanly |
| No progress indication for large files | Appears frozen; users kill it | Consider a per-file timer: if diff takes >2s, print "Diffing large file..." to stderr |

---

## "Looks Done But Isn't" Checklist

- [ ] **Myers algorithm:** Tested with 500+ line files and compared against `diff -u` output — not just 10-line unit tests.
- [ ] **Histogram fallback:** Tested with a file where every line repeats >65 times (e.g., 1000 lines of `}`).
- [ ] **Side-by-side layout:** Tested with terminal width of 80, 120, and 40 columns (narrow terminal edge case).
- [ ] **ANSI strip before width:** Confirmed `lipgloss.Width(chromaLine) == lipgloss.Width(plainLine)` for same logical content.
- [ ] **Piped output:** Verified `drift file1 file2 | cat` produces clean unified diff without ANSI codes.
- [ ] **`go install` path:** Verified `go install github.com/tbcrawford/drift/cmd/drift@latest` works from a clean environment.
- [ ] **Nil lexer handling:** Tested with an unknown file extension (`.xyz`) — no panic, graceful plain-text fallback.
- [ ] **Dark/light detection timeout:** Verified the CLI does not hang on startup when not in an xterm-compatible terminal.
- [ ] **`\r\n` line endings:** Tested with Windows-style line endings — no `\r` suffix visible in diff output.
- [ ] **Binary file:** Tested diffing a binary file — clean error message, no panic.
- [ ] **Color depth levels:** Tested with `COLORTERM=` (unset), `COLORTERM=truecolor`, and `NO_COLOR=1`.

---

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Myers off-by-one discovered post-release | HIGH | Identify scope with cross-validation against `diff -u`; patch algorithm; re-run all golden tests; release patch version |
| Lip Gloss v1→v2 width breakage after upgrade | MEDIUM | Add `GetHorizontalFrameSize()` to layout calculations; update golden test files |
| Breaking public API change post-v1.0 | HIGH | Bump module path to `/v2`; maintain v1 with security patches; provide migration guide |
| Chroma v1/v2 import confusion in dependency | MEDIUM | `go mod why` to find the conflict; force v2 with `replace` directive; audit all import paths |
| ANSI output in piped context discovered by users | LOW | Add `isatty` + `NO_COLOR` check to color detection; patch release |
| Histogram fallback missing, causes wrong output | HIGH | Implement Myers fallback; requires full algorithm regression test suite |

---

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Myers off-by-one | Phase 1: Core diff engine | Cross-validate against `diff -u` on 500+ line files in CI |
| Histogram fallback missing | Phase 1: Core diff engine | Test with all-repeated-lines input |
| Patience inter-anchor fallback | Phase 1: Core diff engine | Test with no-unique-lines input |
| Lip Gloss width/border mismatch | Phase 2: Rendering (split view) | Assert `lipgloss.Width(panel) == expectedWidth` in layout tests |
| ANSI codes inflate string width | Phase 2: Rendering (Chroma + Lip Gloss) | Test layout with pre-highlighted strings |
| Unicode wide char misalignment | Phase 2: Rendering (Unicode) | Test with CJK and emoji content |
| Terminal width detection on pipes | Phase 3: CLI implementation | Test with `drift file1 file2 \| cat` in CI |
| Color in non-TTY / wrong depth | Phase 3: CLI implementation | Test with `NO_COLOR=1`, piped output, and ANSI-only terminal |
| Dark/light detection hang | Phase 3: CLI implementation | Add 200ms timeout; test in non-xterm context |
| Chroma v1/v2 confusion | Phase 2: Chroma integration | Pin to v2 in go.mod from day one; compile check |
| Chroma formatter color depth | Phase 2: Chroma integration | Test with TrueColor/ANSI256/ANSI16 profiles |
| Exported struct breaks API | Phase 1: API design | Use functional options; review before v1.0 tag |
| Module major version path | Phase 4: Release/versioning | Verify `go install /cmd/drift@latest` in release CI |
| Large file O(N²) memory | Phase 1: Core diff engine | Benchmark with 10,000-line file; add to CI benchmarks |
| ANSI in test assertions | Phase 1: Testing infrastructure | Use golden files + `ansi.Strip()` for structure tests |
| `go install` wrong path | Phase 3: CLI / module setup | Smoke test install path in release CI |

---

## Sources

- Myers 1986 paper off-by-one post-mortem: https://dev.to/tommy_worklab/i-implemented-myers-diff-in-130-lines-then-lost-half-a-day-to-an-off-by-one-bug-545d (2026-03-07)
- Histogram diff algorithm deep-dive and working implementation: https://raygard.github.io/2025/01/28/how-histogram-diff-works/ and https://raygard.github.io/2025/01/29/a-histogram-diff-implementation/ (2025-01-29)
- `x/tools/internal/diff` maxDiffs / common prefix issues: https://github.com/golang/go/issues/64023 and https://github.com/golang/go/issues/71648
- Lip Gloss border width bug (#449): https://github.com/charmbracelet/lipgloss/issues/449 (fixed in v2.0.0, 2025-02-10)
- Lip Gloss ANSI wrap regression (#58 in charmbracelet/x): https://github.com/charmbracelet/x/issues/58
- Lip Gloss Width rendering bug with BubbleTea (#1225): https://github.com/charmbracelet/bubbletea/issues/1225
- `go-runewidth` East Asian / box-drawing width issues: https://github.com/mattn/go-runewidth/issues/49, #64
- `charmbracelet/colorprofile` — canonical Charm color detection: https://github.com/charmbracelet/colorprofile
- `muesli/termenv` dark background detection: https://pkg.go.dev/github.com/muesli/termenv
- ANSI support detection Windows broken: https://github.com/zkat/supports-color/issues/22
- Terminal dark/light detection via OSC 11: https://unix.stackexchange.com/questions/245378/common-environment-variable-to-set-dark-or-light-terminal-background
- Go library API design and backwards compatibility: https://abhinavg.net/2022/12/06/designing-go-libraries/ — Abhinav Gupta (2022)
- Embedding structs breaks semver: https://benma.github.io/2020/05/05/golang-embeding-structs-breaks-modularity.html
- Go module layout (official): https://tip.golang.org/doc/modules/layout
- Go project layout best practices: https://alnah.io/post/go-project-layout/ (2026-01-14)
- Golden file testing in Go: https://getpid.dev/blog/golden-tests/ and https://matttproud.com/blog/posts/golden-file-testing.html
- Chroma v2 official: https://github.com/alecthomas/chroma (latest: v2.23.1, 2026-01-23)
- Patience diff Go implementation: https://github.com/peter-evans/patience

---
*Pitfalls research for: Go text diff library + CLI (drift)*
*Researched: 2026-03-25*
