# Phase 27: add function context to hunk header where possible — Context

**Gathered:** 2026-04-10
**Status:** Ready for planning (replanning required — existing plans predate this context)

<domain>
## Phase Boundary

This phase adds function context to drift's rendered hunk header — but ONLY when
that context is already provided by git. Specifically: when drift operates as a git
pager receiving `git diff` output, git already computes `@@ -x,y +a,b @@ func_name`
lines via its own `userdiff.c` regex logic. Drift should parse that `code_fragment`
(the text after the closing `@@`) and display it in a styled format alongside the
line number.

**Out of scope:** Computing function context via custom regex patterns for standalone
`drift.Diff()` mode. That would require a `funcctx` package with per-language regex
tables — significant complexity that is not justified when the "free" git passthrough
covers the primary use case (git pager mode, which is how most users encounter drift).

</domain>

<decisions>
## Implementation Decisions

### Core Approach — Git Passthrough Only

- **D-01:** Function context is extracted from the `code_fragment` portion of
  incoming git `@@ ... @@ code_fragment` hunk header lines. This only applies when
  drift is operating as a git pager (receiving raw git diff text on stdin).

- **D-02:** Standalone `drift.Diff()` mode is unchanged. Hunk headers remain
  `@@ -OldStart,OldLines +NewStart,NewLines @@` with no trailing text. No
  `funcctx` package, no language-specific regex patterns.

- **D-03:** No custom regex engine for function detection. Git already did the
  work. Drift just needs to extract and render what git provided.

### Display Format

- **D-04:** Extract the `code_fragment` from git's `@@` line (trimmed whitespace).
  Display in the rendered hunk header as:
  ```
  111: func_name
  ```
  where `111` is the line number (already shown in drift's hunk header) and
  `func_name` is the extracted code_fragment. Exact styling can be tweaked — the
  format is a starting point, not locked.

- **D-05:** The line number used should be the new-file (`+`) line number from the
  `@@` header (same as the existing line-number display behavior in drift's hunk
  header).

### No-Match / No-Context Fallback

- **D-06:** When no `code_fragment` is present (git didn't add one, unsupported
  language, anonymous function), the hunk header shows just
  `@@ -x,y +a,b @@` — no trailing text. This matches git's own behavior.

### Activation

- **D-07:** Function context display activates automatically when operating in git
  pager mode (drift receives raw git diff output). It is not gated on `WithLang()`
  or any explicit option — if git put the `code_fragment` in the `@@` line,
  drift shows it.

- **D-08:** No new `WithFuncContext(bool)` option is needed for phase 27. The
  feature is implicit in the git pager parsing path. (The planner's pre-context
  design that added `WithFuncContext` and `Hunk.FuncName` to the standalone Diff()
  path is dropped.)

### Replanning Note

- **D-09:** The existing plans (27-01, 27-02) were created before this context and
  implement the WRONG approach (standalone regex + `funcctx` package + `Hunk.FuncName`
  field). They must be deleted and replanned. New plans should target:
  - Plan 1: Parse `code_fragment` from git pager `@@` input lines; store/pass it
    through the pager rendering path.
  - Plan 2: Styled rendering of `{line_number}: {code_fragment}` in hunk headers
    for git pager mode (unified + split); integration test with a real git diff fixture.

### Agent's Discretion

- Exact styling of `{line_number}: {func_name}` (color, weight, separator character).
- Whether to use a new struct field or pass code_fragment inline through the pager path.
- Trimming heuristics for code_fragment (e.g., strip trailing `(` for display or keep as-is).

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Delta — How it handles code_fragment (reference only)

- `https://github.com/dandavison/delta/blob/main/src/handlers/hunk_header.rs` — Delta
  parses `code_fragment` with regex `r"@+ ([^@]+)@+(.*\s?)"` (second capture group).
  Delta does NOT compute function context; it just renders what git provides. This is
  the model for drift's approach.

### Drift — Hunk header rendering (files to modify)

- `internal/render/unified.go:134` — Current hunk header format string (bare `@@` line,
  no code_fragment). The `for _, h := range result.Hunks` loop is where `header` is built.
- `internal/render/split.go:76` — Split renderer hunk header (same bare format).
- `internal/edittype/edittype.go:51` — `Hunk` struct — currently has no `FuncName`
  or `CodeFragment` field (DO NOT add one to the standalone path per D-02/D-08).

### Drift — Git pager input path (where to intercept `@@` lines)

- The git pager code path (from Phase 25) is where incoming `@@ ... @@ func_name`
  lines arrive. Downstream agents should locate this path and understand how
  `@@` lines are currently processed before adding code_fragment extraction.

### Existing plans to delete / replace

- `.planning/phases/27-add-function-context-to-hunk-header-where-possible/27-01-PLAN.md`
- `.planning/phases/27-add-function-context-to-hunk-header-where-possible/27-02-PLAN.md`

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `internal/render/unified.go`: The `Unified()` function builds hunk headers at line 134.
  The format string currently ignores any trailing context — this is where display changes
  would go for standalone mode (but scope says NOT to change standalone for this phase).
- `internal/render/split.go`: Same pattern at line 76 for side-by-side headers.
- `internal/edittype/edittype.go`: `Hunk` struct at line 51 — do not add fields (D-08).

### Established Patterns

- Hunk header formatting is a simple `fmt.Sprintf` in both renderers — easy to extend
  with an optional trailing string if a `CodeFragment` field is ever added (future phase).
- The git pager input path (Phase 25) reads raw git diff lines from stdin and must
  already handle `@@` lines — the code_fragment lives in that parsing layer.

### Integration Points

- Git pager stdin parsing → extract code_fragment → pass to hunk header renderer.
- The connection between the git pager `@@` line parsing and the styled hunk header
  output is the primary integration point for this phase.

</code_context>

<specifics>
## Specific Ideas

- Display format starting point: `{line_number}: {code_fragment}` — user wants something
  like `111: func_name`. Delta shows `file:line: code_fragment`; drift's line number is
  already prominent, so just prepending it as a prefix is the right direction.
- Styling can evolve during implementation — the user explicitly said "we can tweak the
  exact style as we go," so don't over-spec it in planning.

</specifics>

<deferred>
## Deferred Ideas

- **Standalone regex-based function context** — Computing function context for
  `drift.Diff()` standalone mode using per-language regex patterns (like git's
  `userdiff.c`). This is the approach the planner originally designed (funcctx package,
  `Hunk.FuncName`, `WithFuncContext` option). It's valid but adds significant complexity.
  Deferred to a future phase when the user wants it.

- **Porting git's userdiff.c patterns** — User considered this then decided against it
  for phase 27. Good candidate for a "funcctx" phase if standalone function context
  becomes a priority.

</deferred>

---

*Phase: 27-add-function-context-to-hunk-header-where-possible*
*Context gathered: 2026-04-10*
