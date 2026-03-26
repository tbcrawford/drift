# Phase 11: GitHub PR-style intra-line highlights — Context

**Gathered:** 2026-03-26 (initial), **revised:** 2026-03-26 (user session — line/word color parity)
**Status:** Ready for replan / implementation follow-up within Phase 11 scope
**Revision note:** Existing wave-1 context preserved; decisions below **supersede** earlier ambiguity on *which* RGB applies to full line vs intra-line spans.

<domain>
## Phase Boundary

Word-level alignment between paired old/new lines; muted unchanged regions; changed spans highlighted distinctly; Chroma applied per segment; **unified and split** views. Discussion here clarifies **color derivation and layering** so output matches the **GitHub PR diff** reference without expanding scope to unrelated features.

</domain>

<decisions>
## Implementation Decisions

### Visual reference (locked)
- **D-REF-01:** Target behavior is **GitHub PR unified-diff style**: the **entire** added/removed line reads as one semantic row (full-line background), with **intra-line** emphasis on the exact changed chunks. Split view should preserve the same **semantic layering** (left = old / delete, right = new / insert).

### Color derivation vs terrasort (locked)
- **D-COLOR-01:** **terrasort** (`internal/highlight`, especially `DeriveUXTheme`, `diffcolors`, gutter RGB helpers) is the **canonical palette** for *how* RGB/hex are chosen (theme background blend, fallbacks, gutter neutrals). Drift should **re-implement or directly align** that derivation here — **not** invent a parallel scheme that drifts from terrasort’s outputs.
- **D-COLOR-02:** **terrasort’s** weakness called out by the user is **which spans get marked changed** (segmentation / pairing), **not** the colors themselves. Drift keeps **`internal/worddiff`** (Myers-on-tokens) as the **source of truth for changed ranges** unless research proves otherwise.
- **D-COLOR-03:** **Intra-line changed spans** must use the **same color role** as in terrasort’s UX (e.g. gutter-column tint vs line tint — whichever terrasort uses for “changed word” **must** match in drift after porting). If drift currently applies the wrong hex/layer (e.g. semantic red/green on words when reference uses neutral gutter gray), **fix** to match terrasort + GitHub PR layering.
- **D-COLOR-04:** **Supersedes neutral-only intra-line** where it conflicts with product intent: **changed word spans** use **semantic red/green** (`WordSpanBackgroundColour`) **brighter** than the **muted** full-line plane (`DiffLineMutedBackgroundColour` / `DiffLineStyle`). Line-number gutters stay neutral (`GutterBackgroundHex`). (Owner / `11-UAT.md` gap closure, plan **11-04**.)

### Layering (locked intent)
- **D-LAYER-01:** **Full line** (including prefix where applicable) carries the **primary** add/remove semantic background.
- **D-LAYER-02:** **Changed words** receive a **second-layer** highlight that matches the **reference stack** (terrasort colors + GitHub PR-like readability). Unchanged portions of the line stay on the full-line plane (muted syntax as already designed).

### Pipeline boundaries
- **D-PIPE-01:** `internal/highlight` owns **Chroma theme → RGB** for diff chrome; `internal/render` applies Lip Gloss / ANSI; `internal/worddiff` owns **segment ranges only**.

### Claude's Discretion
- Exact function names and whether to **duplicate** terrasort helpers vs **extract shared** package — choose for minimal drift diff and testability.
- Goldens / snapshot tests for ANSI output once colors are stable.

</decisions>

<specifics>
## Specific Ideas

- **Reference repos:** terrasort `/internal/highlight` for **color math**; GitHub PR web UI for **expected layering** (full line + word emphasis).
- User expectation: word-level highlighting **works** (ranges correct) but **colors are wrong** — prioritize **color port + layering** over changing Myers segmentation first.

</specifics>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### External (terrasort — color source)
- `/Users/tylercrawford/dev/github/terrasort/internal/highlight/uxtheme.go` — `DeriveUXTheme`, `DeriveUXThemeFromColor`, gutter vs line fields.
- `/Users/tylercrawford/dev/github/terrasort/internal/highlight/diffcolors.go` — `chromaDiffLineRGBA`, `driftStyleGutterRGBA`, fallbacks.

### Drift (implementation targets)
- `internal/highlight/diffcolors.go`, `internal/highlight/diff_line.go` — line backgrounds.
- `internal/render/gutter.go`, `internal/render/wordline.go`, `internal/render/unified.go`, `internal/render/split.go` — application order (word diff + full line).

### Product reference
- GitHub pull request **Files changed** tab — unified diff presentation (not necessarily identical ANSI, but **same information hierarchy**: full line colored, words emphasized).

</canonical_refs>

<code_context>
## Existing Code Insights

- `internal/worddiff/PairSegments` — segment boundaries (keep unless D-COLOR-02 overridden by research).
- Phase 10 `ApplyDiffLineStyle` / Phase 11 word-diff path — **ordering** of full-line wrap vs per-span tint must satisfy D-LAYER-*.
- Terrasort applies `AddBg`/`RemoveBg` + separate gutter/intra colors in `internal/diff` — compare layering when porting.

</code_context>

<deferred>
## Deferred Ideas

- Shared Go module between drift and terrasort for **one** color table (only if user opens a milestone for it).
- Changing word-diff **algorithm** to match terrasort’s (if ever desired) — explicitly **out of scope** unless user revises D-COLOR-02.

</deferred>
