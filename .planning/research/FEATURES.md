# Feature Research

**Domain:** Go text diff library + CLI tool
**Researched:** 2026-03-25
**Confidence:** HIGH

---

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete or broken.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Myers diff algorithm | It's the de-facto default — used by GNU diff, git diff, go-diff (67k dependents). Users assume it's there. | LOW | O(ND) time complexity; fastest for small edit distances. Well-understood. |
| Unified diff output (Git-style) | Standard patch format. Every diff tool since 1976 emits this. Consumers (CI, code review, patch) expect it. | LOW | `--- a/file +++ b/file @@ -N,M +N,M @@` format with context lines. |
| Context line control | `git diff -U3` is muscle memory for every developer. Missing it means the library can't be used as a drop-in. | LOW | Default 3 lines; configurable. Hunk merging when hunks overlap. |
| Hunk headers (`@@ ... @@`) | Required to produce valid unified diff output. Downstream tools (patch, git apply) need them. | LOW | Line range info. Optional section name (function context) in `@@ -N,M +N,M @@ funcName` form. |
| Diff two strings in-memory | Core library use case: `diff.Compare(a, b)` without requiring file I/O. | LOW | Both inputs as strings. This is what tests and code-review tools call. |
| Diff two files from disk | Core CLI use case. Any CLI that can't take two file paths is broken for shell scripting. | LOW | Open, read, diff. Simple composition on top of string diff. |
| Read from stdin | Unix composability. `cat file | drift -` or piped git diff must work. | LOW | Standard for CLI tools. |
| ANSI color output | All modern diff tools color additions (green) and deletions (red). A colorless diff is jarring in 2026. | LOW | Red for removed lines, green for added lines. Not negotiable. |
| No-color / plain text mode | CI pipelines, file redirection, and `NO_COLOR` env standard require this. | LOW | `--no-color` flag or `NO_COLOR=1` env variable. |
| Idiomatic Go API | Go library users expect `go get`, godoc, proper package name, examples. A C-style or Java-style API fails immediately. | MEDIUM | Clean exported types, functional options pattern, godoc comments. |
| Error handling via return values | Go convention. `error` as second return; panics reserved for programmer errors only. | LOW | No exceptions; always return `(result, error)`. |

---

### Differentiators (Competitive Advantage)

Features that distinguish from the current Go ecosystem. The gap is **rendering quality** — existing Go diff libraries produce raw output, not beautiful terminal output.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Patience diff algorithm | Produces more readable diffs for code — matches function/class boundaries as anchors, avoids matching empty braces. Git's `--patience`. | MEDIUM | Unique-line anchor selection. Used by `peter-evans/patience` (used by typescript-go). Significantly better output for refactored code. |
| Histogram diff algorithm | Git's `--histogram`, an improvement on Patience with frequency-based anchor selection. Handles duplicate unique lines. Preferred by Linus Torvalds. | HIGH | Implemented in JGit, ported to znkr.io/diff and others. Falls back to Myers when no low-frequency anchors exist. |
| Syntax highlighting via Chroma | No existing Go diff library colors tokens — they only color +/- lines. Drift would be the first to layer Chroma token colors under diff colors. | HIGH | Chroma supports 300+ languages, all bat themes. Diff colors (red/green bg) must compose with syntax colors (token fg). Requires careful ANSI layering. |
| Side-by-side (split) diff output | GitHub-style two-panel view. Existing Go libraries only do unified. `delta` (Rust) does this; no Go library does. | HIGH | Lip Gloss column layout. Left = old, right = new. Line numbers in each panel. Long-line wrapping. Terminal-width-aware. |
| Auto-detect terminal light/dark theme | Tools like `delta` do this. Library users shouldn't have to configure Chroma themes manually. | MEDIUM | Inspect `$COLORFGBG`, `$TERM`, or send OSC 11 background color query. Fall back to dark theme. |
| Intra-line (word-level) diff highlighting | Shows *which words within a changed line* differ, not just "this whole line changed." GitHub, VS Code, delta, lazygit all do this. Users now expect it. | HIGH | Run character/word-level Myers diff on matched +/- line pairs within each hunk. Apply stronger background highlights to changed spans. Requires second-pass diff within lines. |
| Both functional options + builder API | Convenience for simple cases; power for complex cases. No existing Go diff library offers both. | MEDIUM | Simple: `drift.Diff(a, b)`. Advanced: `drift.New().WithAlgorithm(drift.Histogram).WithTheme("github").Render(a, b)`. |
| Language auto-detection from file extension | Users shouldn't have to specify `--lang go` when diffing a `.go` file. Chroma's lexer registry handles the mapping. | LOW | Chroma lexer detection by filename. Fallback to plaintext. Overridable via `--lang` flag. |
| Configurable context size + theme as first-class options | Combined with defaults that "just work." Competing libraries require manual ANSI juggling or don't support themes at all. | LOW | `WithContext(5)`, `WithTheme("dracula")`, `WithAlgorithm(drift.Patience)`. |
| `go install`-able CLI, zero external deps at runtime | Go's toolchain makes this easy but most diff tools (delta: Rust binary, diff-so-fancy: Perl+npm) require system package managers. | LOW | `go install github.com/tylercrawford/drift/cmd/drift@latest` — single static binary, nothing to install. |
| Structured diff AST as library output | Beyond string rendering: export the parsed hunk/chunk structure so callers can build their own renderers. | MEDIUM | `FileDiff`, `Hunk`, `Line` types. Callers can range over hunks and do custom formatting. Enables future HTML/JSON/etc. renderers. |

---

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create scope/complexity problems for v1.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| HTML/web rendering output | "GitHub-like output in a browser" is appealing | Completely different rendering pipeline; CSS + HTML escaping + JS interactivity is a separate product surface. Doubles maintenance burden. Out of scope explicitly in PROJECT.md. | Expose the structured diff AST — callers can render HTML themselves. Add as v2 target. |
| Interactive TUI (vimdiff-style scrollable pager) | "I want to scroll through the diff" | Bubble Tea interactive mode is a fundamentally different code path (event loop, input handling, viewport management). Adds Go module complexity and obscures the core library value. Out of scope in PROJECT.md. | Recommend piping output to `less -R`. Add Bubble Tea viewer as an optional sub-package in v2. |
| Real-time file watching / live diff | "Watch two files and re-diff on save" | Requires OS filesystem event handling (`fsnotify`), infinite loop, terminal refresh. Orthogonal to core diff logic. | Use `watch drift file1 file2` at the shell level instead. |
| JSON / machine-readable diff output | "I want to parse the diff programmatically" | Trivial to add but creates API surface area that must be maintained. Drives users toward treating the library as a data source rather than a renderer. | Expose the structured diff AST (Hunk/Line types) — that is the machine-readable form. |
| Patch application (`patch` command equivalent) | `sergi/go-diff` does patch application; users expect it | Patch application is an orthogonal feature (requires fuzzy matching, offset tracking, conflict detection). Creates serious scope creep with different correctness guarantees. | Direct users to `git apply` or `sergi/go-diff`'s patch API for apply functionality. |
| Directory recursive diff | "diff two folders" | Adds file system traversal, ignore patterns, binary detection, stat comparison — a different product from text diffing. | Use `diff -r` at the shell level. Consider as v2 optional command. |
| Binary file diffing | "diff two binaries" | Binary diff (hex diff, byte-level) has no universal format. Meaningful output requires domain knowledge. Text-only tools break gracefully. | Detect binary inputs and emit `Binary files differ` message (same as git diff). |
| Configurable color palette for +/- lines | "I want purple for additions instead of green" | Fragile: users break their own terminals, creates support burden. Standard red/green is universal and expected. | Allow theme selection which indirectly affects color palette via Chroma themes. Hard-code red/green semantics. |
| Fuzzy / approximate string matching | `sergi/go-diff` includes this; it seems related | Fuzzy matching is for spell-check and search, not code diffing. Conflates two different use cases and adds significant dependency surface. | Out of scope. Refer to `sergi/go-diff`'s `diffmatchpatch` package for fuzzy use cases. |

---

## Feature Dependencies

```
[Unified diff output]
    └──requires──> [Myers diff algorithm]
                       └──requires──> [String comparison core]

[Side-by-side split view]
    └──requires──> [Unified diff output] (hunks feed the layout)
    └──requires──> [Lip Gloss layout] (two-column rendering)
    └──enhances──> [Syntax highlighting] (both panels get highlighting)

[Syntax highlighting]
    └──requires──> [Language auto-detection] (to pick the right lexer)
    └──requires──> [Chroma integration] (lexer + formatter)
    └──requires──> [ANSI color output] (terminal rendering pipeline)

[Intra-line word diff]
    └──requires──> [Unified diff output] (hunks identify changed line pairs)
    └──requires──> [Myers diff algorithm] (run again at character/word level)
    └──enhances──> [Syntax highlighting] (word highlights compose with token colors)

[Auto-detect terminal theme]
    └──enhances──> [Syntax highlighting] (picks light vs dark Chroma theme)

[Patience/Histogram algorithms]
    └──requires──> [Myers diff algorithm] (Patience falls back to Myers; Histogram similar)
    └──enhances──> [Unified diff output] (better hunk boundaries)

[CLI tool]
    └──requires──> [Unified diff output]
    └──requires──> [Side-by-side split view]
    └──requires──> [Language auto-detection]
    └──requires──> [Auto-detect terminal theme]

[Functional + builder API]
    └──requires──> [All core diff + render features] (wraps them)

[Structured diff AST output]
    └──requires──> [Myers diff algorithm]
    └──required-by──> [Unified diff output]
    └──required-by──> [Side-by-side split view]
    └──required-by──> [Intra-line word diff]
```

### Dependency Notes

- **Intra-line word diff requires unified diff output first:** You need hunk-level diffing to identify matched +/- line pairs before you can run a second character-level diff within them.
- **Syntax highlighting requires color output:** Chroma produces ANSI escape sequences; these must not be stripped by any plain-text path.
- **Side-by-side requires both Lip Gloss AND unified diff:** Lip Gloss handles column layout; unified diff hunks provide the content. Both must be solid before split view can work.
- **Patience/Histogram enhance but don't replace Myers:** Both algorithms fall back to Myers internally. Myers is the non-negotiable foundation.
- **Structured diff AST is the foundation everything renders from:** If the AST types are well-designed, all render modes (unified, split, word-level) compose naturally from it.

---

## MVP Definition

### Launch With (v1)

Minimum viable product — what's needed for the core value proposition: "one import, GitHub-quality diff."

- [ ] **Myers diff algorithm** — non-negotiable foundation; everything depends on it
- [ ] **Unified diff output** — the standard format; needed for both library and CLI use
- [ ] **Syntax highlighting via Chroma** — the primary differentiator vs existing Go libraries; must ship at launch to justify the project
- [ ] **Side-by-side split view** — the other primary differentiator; GitHub-quality means split + unified
- [ ] **Auto-detect terminal theme (light/dark)** — mandatory for Chroma to "just work" without config
- [ ] **Language auto-detection from file extension** — must be automatic; requiring `--lang` is user friction
- [ ] **Patience diff algorithm** — second most-used algorithm after Myers; low enough complexity to include at launch
- [ ] **Functional options + simple `Diff()` function** — dual API from day one; harder to add builder pattern later without breaking changes
- [ ] **CLI: two file paths, stdin, `--no-color`, `--lang`, `--theme`** — core CLI surface; must be complete at v1
- [ ] **Structured diff AST types (`FileDiff`, `Hunk`, `Line`)** — must be right at v1; changing these is a breaking change

### Add After Validation (v1.x)

- [ ] **Histogram diff algorithm** — valuable but higher complexity; validate Patience adoption first
- [ ] **Intra-line word-level diff highlighting** — users will request this; add once core rendering pipeline is proven stable
- [ ] **`--context N` flag** — configurable context lines; table stakes for CLI but can ship with hardcoded 3 initially
- [ ] **`--algorithm` flag for CLI** — algorithm selection; add after all three algorithms ship

### Future Consideration (v2+)

- [ ] **Interactive Bubble Tea viewer** — optional sub-package; adds significant complexity; defer until core is adopted
- [ ] **HTML render target** — explicitly out of scope in PROJECT.md; revisit only if library gains adoption in web tooling contexts
- [ ] **Directory recursive diff** — useful CLI feature but orthogonal to core library value; defer
- [ ] **Custom Chroma theme loading from file** — nice for power users; defer until basic theme selection is validated

---

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Myers diff algorithm | HIGH | LOW | P1 |
| Unified diff output | HIGH | LOW | P1 |
| ANSI color output (red/green) | HIGH | LOW | P1 |
| Structured diff AST types | HIGH | MEDIUM | P1 |
| Chroma syntax highlighting | HIGH | HIGH | P1 |
| Side-by-side split view | HIGH | HIGH | P1 |
| Functional options API + `Diff()` shorthand | HIGH | MEDIUM | P1 |
| Language auto-detection | HIGH | LOW | P1 |
| Auto-detect terminal theme | HIGH | MEDIUM | P1 |
| Patience diff algorithm | MEDIUM | MEDIUM | P1 |
| CLI: two-file, stdin, flags | HIGH | LOW | P1 |
| No-color / `NO_COLOR` support | HIGH | LOW | P1 |
| Configurable context lines | MEDIUM | LOW | P2 |
| Histogram diff algorithm | MEDIUM | HIGH | P2 |
| Intra-line word-level diff | HIGH | HIGH | P2 |
| `--algorithm` CLI flag | MEDIUM | LOW | P2 |
| Interactive TUI viewer | LOW | HIGH | P3 |
| HTML render target | LOW | HIGH | P3 |
| Recursive directory diff | LOW | MEDIUM | P3 |

**Priority key:**
- P1: Must have for launch
- P2: Should have; add after core validates
- P3: Nice to have; future consideration

---

## Competitor Feature Analysis

| Feature | `sergi/go-diff` (67k dependents) | `pmezard/go-difflib` (597k dependents, abandoned) | `sourcegraph/go-diff` | `delta` (Rust, 29k ⭐) | `diff-so-fancy` (18k ⭐) | **drift (planned)** |
|---------|----------------------------------|--------------------------------------------------|----------------------|----------------------|-------------------------|----------------------|
| Myers algorithm | ✅ (character-level) | ✅ (line-level) | ❌ (parser only) | ✅ | ✅ | ✅ |
| Patience algorithm | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ |
| Histogram algorithm | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ (v1.x) |
| Unified diff output | ❌ (not really) | ✅ | ✅ (parse only) | ✅ | ✅ | ✅ |
| Side-by-side split view | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ |
| Syntax highlighting | ❌ | ❌ | ❌ | ✅ (bat themes) | ❌ | ✅ (Chroma) |
| Intra-line word diff | ✅ (character) | ❌ | ❌ | ✅ (Levenshtein) | ✅ (basic) | ✅ (v1.x) |
| Auto light/dark theme | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ |
| Language auto-detection | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ |
| `go install`-able binary | N/A | N/A | ✅ (cmd) | ❌ (Rust/cargo) | ❌ (npm/perl) | ✅ |
| Idiomatic Go library API | ⚠️ (old style) | ⚠️ (abandoned) | ✅ | N/A | N/A | ✅ (functional options) |
| Structured diff AST | ⚠️ (partial) | ❌ | ✅ (FileDiff) | N/A | N/A | ✅ |
| Active maintenance | ⚠️ (slow) | ❌ (abandoned) | ⚠️ (infrequent) | ✅ | ✅ | ✅ (new) |

**Key insight:** The Go ecosystem has strong algorithm libraries (sergi/go-diff) and strong format parsers (sourcegraph/go-diff) but **zero libraries that combine diff computation + rich terminal rendering** in a single idiomatic package. That gap is exactly what drift fills.

---

## Sources

- `github.com/sergi/go-diff` — Diff, match and patch library (67k dependents) https://github.com/sergi/go-diff
- `github.com/pmezard/go-difflib` — Abandoned unified diff library (597k dependents) https://github.com/pmezard/go-difflib
- `github.com/sourcegraph/go-diff` — Unified diff parser/printer https://github.com/sourcegraph/go-diff
- `github.com/peter-evans/patience` — Go patience diff (used by microsoft/typescript-go) https://github.com/peter-evans/patience
- `github.com/hexops/gotextdiff` — Archived gopls-based unified diff https://github.com/hexops/gotextdiff
- `znkr.io/diff` — High-performance Myers+heuristics diff (v1.0.0 Mar 2026) https://git.durrantlab.pitt.edu/znkr/diff
- `github.com/dandavison/delta` — Rust syntax-highlighting pager (29.7k ⭐) https://github.com/dandavison/delta
- `github.com/so-fancy/diff-so-fancy` — Perl diff beautifier (18k ⭐) https://github.com/so-fancy/diff-so-fancy
- `github.com/banga/git-split-diffs` — TypeScript side-by-side diffs https://github.com/banga/git-split-diffs
- `github.com/arran4/golang-diff` — New Go side-by-side + char diff CLI (Feb 2026) https://github.com/arran4/golang-diff
- WizlyTools diff algorithm comparison (March 2026) https://wizlytools.com/blog/text-diff-algorithms-explained
- Histogram diff deep-dive (Jan 2026) https://raygard.github.io/2025/01/29/a-histogram-diff-implementation/
- Git diff-options documentation https://git-scm.com/docs/diff-options
- Diff algorithm spelunking (Dec 2025) https://dacharycarey.com/2025/12/29/diff-algorithm-spelunking/
- microsoft/typescript-go patience PR (May 2025) https://github.com/microsoft/typescript-go/pull/891

---
*Feature research for: Go text diff library + CLI (drift)*
*Researched: 2026-03-25*
