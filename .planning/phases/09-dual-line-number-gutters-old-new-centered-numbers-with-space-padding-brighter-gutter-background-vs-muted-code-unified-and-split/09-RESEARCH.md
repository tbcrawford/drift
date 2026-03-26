# Phase 9: Dual line-number gutters — Research

**Phase goal:** Render **two** line-number columns (old file | new file) with **space-padded centered** numerals, **gutter background** visually distinct from the **code** area (brighter/dimmed contrast), for **both** unified and split renderers.

**Roadmap:** Depends on Phase 8 (theme / palette). Requirements in REQUIREMENTS.md are still TBD for this phase; scope is taken from the phase title and existing `internal/render` architecture.

**Research date:** 2026-03-26

---

## Executive Summary

1. **`edittype.Line` already carries `OldNum` and `NewNum`** (1-indexed; `0` means “not applicable” for that side). The hunk builder populates these — renderers must not re-derive line numbers from scratch.

2. **Unified layout (target):** For each content line, emit **gutter_old | gutter_new | diff_prefix | highlighted_code**, where `diff_prefix` remains `+`, `-`, or space (REND-01 alignment). The two gutters are **fixed-width columns** per hunk (or per file-diff) so columns align vertically. Between the two number columns, a narrow separator (e.g. `│` or ASCII `|`) matches the “old | new” wording in the roadmap.

3. **Split layout (target):** Each panel gets **one** line-number column adjacent to its code: left panel = **old** numbers + old-side content; right panel = **new** numbers + new-side content. The existing center ` │ ` column separator stays between the two panels. **Width budget:** subtract gutter column widths from each panel before `lipgloss` `MaxWidth` / `Width` on code — same pattern as Phase 4 (`internal/render/split.go`).

4. **Centering:** “Centered numbers with space padding” = for column width `W`, format each number (or blank for `0`) so the runes are **horizontally centered** in `W` (Unicode cell width `1` for ASCII digits). Implement a small `centerString(s string, width int) string` using rune count or byte length for ASCII-only line numbers.

5. **Gutter vs code contrast:** Use **Lip Gloss v2** `lipgloss.NewStyle().Background(...).Foreground(...)` on gutter cells only. For **code**, optionally wrap in `Faint(true)` when `!cfg.NoColor` so gutters read “brighter” (higher luminance background) and code reads “muted” relative to gutters — **without** stripping Chroma token colors (apply faint to the composed line or use a documented compromise). When `NoColor`, use spaces/padding only (no ANSI) so columns still align.

6. **Adaptive colors:** Follow `internal/theme` / existing dark–light detection (`HasDarkBackground` path via `theme.DetectDarkBackground`): use `lipgloss.AdaptiveColor` or paired light/dark hex pairs for gutter background so light terminals don’t get illegible gutters.

7. **Public API:** Add a **render-only** toggle (default **on** for new behavior) via `drift.WithLineNumbers(bool)` or `WithLineNumbers()` / `WithoutLineNumbers()` so library users can match `git diff` or scripts. Thread the flag through `render.RenderConfig` in `render.go` when building `rcfg`.

8. **Performance:** Gutter width is O(lines in hunk) to scan max line numbers — negligible vs Chroma highlighting.

---

## Current Code Anchors

| File | Role |
|------|------|
| `internal/render/unified.go` | Emits `---`/`+++`, `@@` headers, `prefix+highlighted` per line — **no** line columns yet |
| `internal/render/split.go` | `pairHunkLines`, `JoinHorizontal`, per-panel `lipgloss.NewStyle().Width(...)` |
| `internal/render/termwidth.go` | Terminal width for split |
| `internal/edittype/edittype.go` | `Line.OldNum`, `Line.NewNum` |
| `render.go` | Builds `RenderConfig`, selects `Unified` vs `Split` |
| `options.go` | Add line-number option on `config` |

---

## Design Notes: Unified Row Composition

- **Order:** `[oldGutter][sep][newGutter][diffPrefix][code]`  
- **Diff prefix:** Keep a **single** character column (`+`/`-`/space) immediately before highlighted code, consistent with REND-01.
- **Blank line numbers:** For inserts, `OldNum == 0`; for deletes, `NewNum == 0`. Render an **empty** centered field (spaces only) in that column so alignment is stable.

---

## Design Notes: Split Panel Composition

- **Left:** `[oldGutter][oldCodeStyled]` then join with separator then **Right:** `[newGutter][newCodeStyled]`.
- Recompute **panel content width** = `(termWidth - separator - leftGutter - rightGutter)` split appropriately, or **inner** content width = previous `panelWidth - gutterWidth` per side.
- **Placeholder rows** (empty opposite side) still get gutter cells — use same `0` rules as unified.

---

## Competitor / Reference UX

- **GitHub PR / VS Code:** Narrow old/new columns, subtle background — align with “brighter gutter vs muted code”.
- **delta:** Line numbers + sidebars — borrow **column discipline** (fixed width, padding), not extra dependencies.

---

## Risks

| Risk | Mitigation |
|------|------------|
| ANSI + Lip Gloss layers interact badly with Chroma | Build gutter strings separately; apply `lipgloss` only to gutter + optional faint wrapper; test with `go test` and one TrueColor snapshot |
| Wide terminals vs narrow | Gutter width derived from **max line number in hunk** (cap e.g. 6 digits if needed) |
| `NoColor` still needs alignment | Space-only padding; no background |

---

## Validation Architecture

**Dimension 8 (Nyquist):** Execution must not go silent >3 tasks without automated feedback.

| Layer | Mechanism |
|-------|-----------|
| **Unit** | `internal/render/*_test.go`: table tests for `centerString`, gutter width computation, and **substring** assertions on rendered output (e.g. contains `@@`, line numbers `1`, `2` for known fixtures) |
| **Integration** | `render_test.go`: `Render(..., WithSplit())` and unified — assert **stable** old/new columns for a **fixed** `DiffResult` fixture |
| **Regression** | `go test ./... -count=1` green after each wave |
| **Manual** | One terminal pass: colored output — gutters readable, code not washed out; repeat with `NO_COLOR` |

**Wave 0:** Existing Go tests — extend `unified_test.go` / `split_test.go` with line-number assertions; no new framework.

**Stopping rule:** If Lip Gloss faint + Chroma interaction is unreadable, **drop faint** and rely on **background-only** gutter distinction (document in SUMMARY).

---

## Out of Scope (this phase)

- Intra-line word highlights (Phase 11)
- Full-line red/green theme styling (Phase 10) — but avoid blocking hooks for later theme overlays

---

## RESEARCH COMPLETE

Proceed to planning: shared gutter helpers (`internal/render/gutter.go` or similar), then `unified.go`, then `split.go`, then `drift` options + tests.
