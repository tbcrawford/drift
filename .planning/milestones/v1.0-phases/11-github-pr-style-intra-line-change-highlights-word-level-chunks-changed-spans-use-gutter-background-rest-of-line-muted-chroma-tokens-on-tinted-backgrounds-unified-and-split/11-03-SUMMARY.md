---
phase: 11-github-pr-style-intra-line-change-highlights-word-level-chunks-changed-spans-use-gutter-background-rest-of-line-muted-chroma-tokens-on-tinted-backgrounds-unified-and-split
plan: 11-03
subsystem: testing
tags: [chroma, lipgloss, diff, terrasort-parity]

requires: []
provides:
  - Terrasort-aligned full-line diff RGB from Chroma GenericInserted/GenericDeleted
  - Neutral-only line-number gutters; semantic colour on full code line
affects: []

tech-stack:
  added: []
  patterns:
    - Match terrasort chromaDiffLineRGBA + fallback hexes in drift diffcolors

key-files:
  created: []
  modified:
    - internal/highlight/diffcolors.go
    - internal/highlight/diff_line.go
    - internal/highlight/diffcolors_test.go
    - internal/render/gutter.go
    - internal/render/gutter_test.go
    - internal/render/unified_test.go
    - .planning/phases/.../11-VERIFICATION.md

key-decisions:
  - Line-number gutter cells use GutterBackgroundHex only; semantic backgrounds apply via ApplyDiffLineStyle on the rendered line

requirements-completed: []

duration: ""
completed: "2026-03-26"
---

# Plan 11-03 — Terrasort colour parity + layering

**Full-line diff colours now follow terrasort’s Chroma pipeline (GenericInserted/GenericDeleted, terminal-base blend, same fallbacks); gutters stay neutral; tests cover github-dark bias, gutter tint, and word-diff full-line CSI 48.**

## Performance

- **Tasks:** 5 (T1–T5)
- **Files modified:** see frontmatter

## Accomplishments

- Ported `DiffLineBackgroundColour` to terrasort-equivalent `diffEntryChromaColour` + `blendChromaTowardTerminalBase` + shared fallbacks.
- `gutterStyleForCell` no longer applies semantic red/green to line-number columns; `gutterTintStyle` unchanged (neutral only).
- Confirmed unified/split word-diff paths already wrap with full-line diff style after segmentation; extended `TestUnified_WordDiffPairedDeleteInsert` for `\x1b[48;`.
- Updated `11-VERIFICATION.md` with 11-03 automated vs human checks.

## Self-Check: PASSED

- `go test ./...` and `just lint` succeed.
