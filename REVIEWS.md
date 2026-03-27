---
reviewers: [self (claude-sonnet-4.6)]
reviewed_at: 2026-03-27T19:35:37Z
note: No external AI CLIs available; single-AI self-review of full repository
---

# Repository Review — drift

## Self-Review (claude-sonnet-4.6)

### Summary

`drift` is a well-structured, idiomatic Go library with a clean public API, thoughtful separation of concerns, and real production polish. The diff algorithm layer (Myers, Patience, Histogram), rendering pipeline (unified + split), and CLI are all coherent. The codebase is ready for a v1.0 release with a small number of targeted fixes.

One off-by-one in the histogram algorithm needs correction before v1.0. A handful of inconsistencies and documentation gaps are lower-risk but worth addressing for a production OSS library.

---

### Strengths

- **Layered architecture**: `internal/algo`, `internal/hunk`, `internal/render`, `internal/highlight` each have clear single responsibilities. The `Differ` interface in `internal/algo/algo.go` is minimal and correct.
- **Idiomatic public API**: `drift.Diff` + `drift.Render` for one-shot use; `drift.Builder` for composable configuration. The `With*` option functions are consistent and discoverable.
- **Testing coverage**: 219 tests including property-based tests (`pgregory.net/rapid`), golden file tests (`goldie`), and CLI integration tests (`testscript`). The test strategy matches the stack recommendations.
- **Color pipeline**: `colorprofile` → `FormatterForProfile` → Chroma ANSI stream → Lip Gloss is clean. No-color / non-TTY fallbacks are handled correctly.
- **CLI ergonomics**: Two-file, stdin, and git working-tree modes are all supported. The `exitCodeErr` pattern cleanly separates diff-found (exit 1) from error (exit 2).

---

### Concerns

#### Bugs

**BUG-01 · LOW · `internal/algo/histogram/histogram.go`**

`lowcount` is initialized to `histogramMaxOccurrences + 1` (65) instead of `histogramMaxOccurrences` (64). This means lines appearing exactly 65 times pass the `cnt > lowcount` check and are selected as histogram anchors, when they should trigger a Myers fallback. The constant is named `histogramMaxOccurrences` — the intent is clear; the initialization is off by one.

```go
// current (wrong):
lowcount := histogramMaxOccurrences + 1

// correct:
lowcount := histogramMaxOccurrences
```

---

#### Inconsistencies

**INCONSISTENCY-01 · LOW · Duplicated `applyOffset` helper**

`applyOffset` is defined identically in both `internal/algo/patience/patience.go` and `internal/algo/histogram/histogram.go`. It should be extracted to `internal/algo/algo.go` (the shared package file) as an unexported helper.

---

**INCONSISTENCY-02 · LOW · `Builder` missing `ThemeResolved` method**

`drift/builder.go` exposes chainable `With*` methods for every `drift.Option` except `WithThemeResolved`. A `Builder.ThemeResolved(fn func(string)) *Builder` method is missing, breaking the symmetry of the builder pattern.

---

**INCONSISTENCY-03 · LOW · `Render` / `RenderWithNames` share ~80 lines of duplicated setup**

Both functions in `drift/render.go` independently construct `cfg`, `profile`, `lexer`, `style`, `formatter`, `writer`, `termWidth`, and `rcfg`. This ~80-line block is identical except for the lexer detection call. A private `buildRenderDeps(opts []Option, filename, content string) (...)` helper would eliminate the duplication.

---

**INCONSISTENCY-04 · LOW · Hardcoded `isDark=true` fallback in `internal/render/unified.go`**

Line 97 calls `highlight.SelectTheme("", true)` with `isDark` hardcoded to `true`. This fallback only triggers when `cfg.Style == nil`, which cannot happen via the public API (the default is always set), but the hardcoded value is inconsistent and would mislead on light-background terminals if ever reached.

---

**INCONSISTENCY-05 · MEDIUM · `go.mod` declares `go 1.25.0`; documentation says Go 1.21+**

`go.mod` sets `go 1.25.0` as the minimum, but `CLAUDE.md` documents "Go 1.21+" and the stack research explicitly targets Go 1.21+ for generics and modern stdlib. This prevents library importers on Go 1.21–1.24 from using `drift`, which contradicts the documented minimum. The `go` directive should be set to the actual minimum tested version.

---

**INCONSISTENCY-06 · LOW · `headBlobMissing` parses English-only git error strings**

`cmd/drift/gitworktree.go`'s `headBlobMissing` detects new/untracked files by matching the literal string `"exists on disk, but not in"` in git's stderr. This breaks on non-English git installations. The git `--no-optional-locks` or porcelain status approach would be locale-independent.

---

**INCONSISTENCY-07 · LOW · Redundant `false` assignment in `internal/worddiff/worddiff.go`**

`pairWithTokenizer` sets `oldChanged[i] = false` / `newChanged[i] = false` in the `Equal` branch. `make([]bool, n)` already initializes to `false`; these assignments are no-ops. Minor, but misleads readers into thinking `false` is a meaningful write.

---

#### Documentation Gaps

**DOC-01 · LOW · `drift.Diff` always returns `nil` error**

The `Diff` function signature is `func Diff(a, b string, opts ...Option) (DiffResult, error)`. In the current implementation the error is always `nil` — none of the algorithm implementations can fail. This is not documented. Users check `if err != nil` unnecessarily. The godoc should note: "err is currently always nil; reserved for future algorithm implementations."

---

**DOC-02 · LOW · `edittype.Span` / `Line.Spans` is always nil in v1.0**

`internal/edittype/edittype.go` marks `Span` as "Reserved for v1.x" in a comment, but `drift/types.go` re-exports `Line` (which contains `Spans []Span`) without surfacing this reservation. Callers inspecting `Line.Spans` will always get nil. The re-export should include the note, or `Spans` should be omitted from the public API until implemented.

---

**DOC-03 · LOW · `drift.Render` never detects lexer from content**

`Render` (the no-names variant) calls `DetectLexer("", "")` — passing empty strings for both filename and content. This means it always falls back to plain text. `RenderWithNames` correctly passes filenames for extension-based detection. This behavioral difference is not documented and may surprise users who expect `Render` to detect syntax from content.

---

**DOC-04 · LOW · `ReplaceAnsiBackground` appears to be dead code**

`internal/highlight/linebackground.go` exports `ReplaceAnsiBackground`. The word-diff renderer (`internal/render/wordline.go`) uses `highlightLineWithSegments` (segment-based approach) and does not call `ReplaceAnsiBackground`. If this function is genuinely unused, it should be removed or have its retention reason documented.

---

**DOC-05 · LOW · `SelectTheme` godoc comment misnames the dark theme**

`internal/highlight/highlight.go`'s `SelectTheme` doc comment says "monokai for dark, github for light". The implementation returns `"github-dark"` for dark terminals, not `"monokai"`. The comment should be updated to match.

---

**DOC-06 · LOW · `edittype.go` package doc references wrong import path**

The package doc in `internal/edittype/edittype.go` says: "Consumers should use the re-exported types in the root drift package: drift.Op, drift.Edit...". The root module path is `github.com/tylercrawford/drift`; the public library package is at `github.com/tylercrawford/drift/drift`. The doc should reference the full import path.

---

### Risk Assessment

| ID | Severity | Release-Blocking? |
|----|----------|-------------------|
| BUG-01 | LOW | Yes — histogram behavior is subtly wrong for high-frequency lines |
| INCONSISTENCY-05 | MEDIUM | Yes — `go 1.25.0` in `go.mod` contradicts documented minimum and limits adopters |
| INCONSISTENCY-01 | LOW | No |
| INCONSISTENCY-02 | LOW | No |
| INCONSISTENCY-03 | LOW | No |
| INCONSISTENCY-04 | LOW | No |
| INCONSISTENCY-06 | LOW | No |
| INCONSISTENCY-07 | LOW | No |
| DOC-01 | LOW | No |
| DOC-02 | LOW | No |
| DOC-03 | LOW | No |
| DOC-04 | LOW | No |
| DOC-05 | LOW | No |
| DOC-06 | LOW | No |

**2 items are recommended to fix before tagging v1.0**: BUG-01 and INCONSISTENCY-05.

The remaining 12 items are quality improvements suitable for a follow-on patch or the first v1.x release.
