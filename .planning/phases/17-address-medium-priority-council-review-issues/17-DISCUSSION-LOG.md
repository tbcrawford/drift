# Phase 17: Address Medium-Priority Council Review Issues - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-31
**Phase:** 17-address-medium-priority-council-review-issues
**Areas discussed:** CLI module strategy, Phase 17 scope reshape, Golden file test approach

---

## CLI Module Strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Single module (Recommended) | Keep one go.mod at root. Library and CLI share the same module. Simpler distribution, fewer moving parts. | |
| Separate CLI module | Add cmd/drift/go.mod. CLI gets its own dependency graph. Better for future CLI-only deps (e.g., heavy TUI framework). | ✓ |

**User's choice:** Separate CLI module — `cmd/drift/go.mod`

**Notes:** User chose CLI isolation over simplicity, anticipating future CLI-only dependencies.

---

### Follow-up: go.work workspace

| Option | Description | Selected |
|--------|-------------|----------|
| Only cmd/drift/go.mod (Recommended) | CLI module only. Library stays at root. No workspace. | |
| go.work workspace for local dev | Use Go workspace (go.work) to link root module and cmd/drift module during development. | ✓ |

**User's choice:** `go.work` workspace for local development

**Notes:** Workspace avoids needing a `replace` directive or published tag during development. Follows pattern used by multi-module Go repos.

---

## Phase 17 Scope Reshape

| Option | Description | Selected |
|--------|-------------|----------|
| Plan 1 = library-to-root migration (Recommended) | Reshape Plan 1 to be the migration. Plans 2-5 stay as-is. 5 plans total. | ✓ |
| New Plan 0 as prerequisite | Add a Plan 0 for migration. Keep existing 5 plans. 6 plans total. | |
| Migration = separate phase | Library migration becomes Phase 18. Phase 17 stays as 4 code quality fixes. | |

**User's choice:** Plan 1 = library-to-root migration (replace the "document double-path" approach)

**Notes:** Keeps phase count at 5, avoids plan numbering churn. Existing Plans 2-5 remain conceptually valid.

---

### Follow-up: Execution order

| Option | Description | Selected |
|--------|-------------|----------|
| Ordered sequence, tests pass after each (Recommended) | Plans run 1→2→3→4→5. All 223 tests pass after each plan before next begins. | ✓ |
| Plan 1 first, then 2-5 in any order | Plan 1 is prerequisite; 2-5 can run in any order after. | |

**User's choice:** Ordered sequence with test gate after each plan

**Notes:** Provides clear progress checkpoints and ensures no plan masks another plan's failures.

---

## Golden File Test Approach

### Fixture format

| Option | Description | Selected |
|--------|-------------|----------|
| NoColor mode fixtures (Recommended) | Tests call WithNoColor() so ANSI escapes are stripped. Plain-text fixtures. CI-portable. | ✓ |
| Full ANSI fixtures | Capture actual ANSI escape sequences. Fragile, terminal-specific. | |
| Both: plain + ANSI fixtures | Two fixture sets. More coverage, more maintenance. | |

**User's choice:** NoColor mode fixtures — plain-text, CI-portable

---

### Fixture location

| Option | Description | Selected |
|--------|-------------|----------|
| testdata/golden/ at root (Recommended) | Adjacent to library files after migration. Standard Go convention. | ✓ |
| Nested in internal/render/testdata/ | Co-located with render package. | |
| Per-package testdata dirs | One fixture dir per test file. | |

**User's choice:** `testdata/golden/` at repo root

---

### Coverage scope

| Option | Description | Selected |
|--------|-------------|----------|
| Public API surface (Unified + Split) (Recommended) | Test what library users actually call. Representative inputs. | |
| Internal render functions | Test internals directly. More granular. | |
| Both public + internal | Full coverage: public API golden tests + internal render golden tests. | ✓ |

**User's choice:** Both public API and internal render functions

**Notes:** User wanted full regression coverage across both layers despite higher maintenance. The `goldie/v2` `go test -update` flag mitigates maintenance burden.

---

## Agent's Discretion

- Exact fixture scenario count and edge cases — left to executor's judgment
- Whether `go.work` is committed or gitignored — executor should follow Go community convention for multi-module repos

## Deferred Ideas

None — discussion stayed within phase scope.
