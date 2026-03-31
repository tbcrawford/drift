# Phase 17: Address Medium-Priority Council Review Issues - Context

**Gathered:** 2026-03-31
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase delivers five sequential code-quality improvements identified in the council review at `.reviews/drift-library/REVIEW.md`. The biggest structural change is moving the library package from `drift/` to the module root — making `import "github.com/tylercrawford/drift"` the canonical import path. The remaining four plans fix API surface issues, add golden regression tests, improve rendering correctness, and clean up a duplicate terminal dependency.

All 223 existing tests must pass after each plan before the next plan begins.

</domain>

<decisions>
## Implementation Decisions

### CLI Module Strategy

- **D-01:** The CLI (`cmd/drift`) gets its own `go.mod` — a separate module from the library. This keeps the library's dependency graph clean for importers and allows the CLI to add heavier TUI deps in the future without affecting `go get github.com/tylercrawford/drift`.
- **D-02:** A `go.work` workspace file at the repo root links the library module and the CLI module for local development. This avoids needing a `replace` directive or a published tag during development.
- **D-03:** `go install github.com/tylercrawford/drift/cmd/drift@latest` must continue to work from either layout (confirmed: works with separate `go.mod` inside `cmd/drift/`).

### Phase 17 Plan Structure

- **D-04:** Five plans, executed in strict order. All 223 tests must pass after each plan before moving to the next.
- **D-05:** Plan 1 is reshaped from "document the double import path" to "perform the full library-to-root migration." This is the largest structural change and a prerequisite for all other plans.
- **D-06:** Plan sequence:
  1. Library-to-root migration (move 13 files, update 10 import sites, set up `cmd/drift/go.mod` + `go.work`)
  2. Remove `Line.Spans` stub (`[]Span` field and public `Span` alias)
  3. Add golden file tests
  4. Fix `pairHunkLines` positional pairing (bottom-align inserts in asymmetric blocks)
  5. Audit and drop duplicate `golang.org/x/term` / `charmbracelet/x/term` dependency
- **D-07:** Plans 2–5 are independent of each other but all depend on Plan 1 completing first (import paths change in Plan 1).

### Golden File Tests

- **D-08:** Fixtures use NoColor mode — ANSI escape sequences are stripped before snapshotting. This makes fixtures plain-text and portable across any CI environment without TTY detection concerns.
- **D-09:** Golden fixture files live at `testdata/golden/` at the repo root (adjacent to the library source files after migration). This follows the Go convention of `testdata/` adjacent to test files.
- **D-10:** Coverage spans both layers:
  - Public API: `Unified()` and `Split()` with representative inputs (empty diff, small, large, asymmetric hunks)
  - Internal render functions: unified renderer, split renderer, hunk builder — to catch subtle formatting regressions
- **D-11:** Use `goldie/v2` for fixture management — provides `go test -update` flag to regenerate fixtures and colored diff output on failure.

### Agent's Discretion

- Exact set of fixture scenarios (how many inputs, what edge cases to cover) — planner/executor can decide based on what gives meaningful regression coverage without excessive maintenance burden.
- Whether `go.work` is committed to the repo or added to `.gitignore` — common practice varies; executor should follow whatever is standard for Go multi-module repos with a CLI consumer.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Council Review (source of all issues)
- `.reviews/drift-library/REVIEW.md` — Full council review; medium-priority issues are the source of all 5 plans in this phase

### Current Codebase Structure
- `drift/` — 13 library source files to migrate to repo root (Plan 1)
- `cmd/drift/` — CLI; gets its own `go.mod` in Plan 1
- `internal/render/split.go` — `pairHunkLines` function to fix in Plan 4
- `internal/terminal/` — dual term dep to audit in Plan 5

### Project Planning
- `.planning/ROADMAP.md` — Phase 17 entry with original plan descriptions
- `.planning/PROJECT.md` — "Out of Scope" entry for "Separate library and CLI modules" must be updated to reflect that CLI separation IS happening in this phase

### Go Module Layout Reference
- `go.mod` at repo root — current single module; Plan 1 splits this into library module (root) + CLI module (`cmd/drift/go.mod`)

No external specs or ADRs beyond the above — requirements fully captured in decisions above.

</canonical_refs>

<code_context>
## Existing Code Insights

### Files to Migrate (Plan 1)
- `drift/*.go` — 13 files (drift.go, options.go, options_test.go, types.go, builder.go, render.go, doc.go, and associated test files) move to repo root
- Package declaration changes from `package drift` (in `drift/`) to `package drift` (at root) — name stays the same

### Import Sites to Update (Plan 1)
- `cmd/drift/main.go` — imports `"github.com/tylercrawford/drift/drift"` → `"github.com/tylercrawford/drift"`
- `cmd/drift/flags.go` — same
- `examples/basic/main.go` — same
- `examples/builder/main.go` — same
- 6 additional test files referencing `drift/drift` — all need updating
- Total: 10 import sites

### Unaffected Packages
- `internal/` packages — only import from `internal/...` paths; unaffected by library relocation
- `internal/edittype/` — type cycle-breaker pattern; stays as-is regardless

### Key Functions
- `internal/render/split.go` — `pairHunkLines` function: currently uses positional pairing; Plan 4 fixes to bottom-align inserts in asymmetric delete/insert blocks
- `drift/types.go` — `Line` struct with `Spans []Span` field (stub); Plan 2 removes this
- `internal/terminal/` — contains duplicate term dependency; Plan 5 audits `golang.org/x/term` vs `charmbracelet/x/term`

### Test Infrastructure
- 223 existing tests — all must pass after each plan
- `goldie/v2` not yet a dependency — Plan 3 adds it
- `testdata/` directories currently exist per-package; Plan 3 adds `testdata/golden/` at root for public API fixtures

</code_context>

<specifics>
## Specific Ideas

- Library import path becomes `"github.com/tylercrawford/drift"` — exactly what a Go developer would expect and what PROJECT.md's core value statement describes
- The `go.work` workspace pattern follows how `golang.org/x/tools` and other multi-module Go repos handle library + CLI separation
- Golden test fixtures in NoColor mode should be human-readable — a developer should be able to open a `.golden` file and understand what the diff output looks like structurally

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 17-address-medium-priority-council-review-issues*
*Context gathered: 2026-03-31*
