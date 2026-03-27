---
phase: 11-github-pr-style-intra-line-change-highlights-word-level-chunks-changed-spans-use-gutter-background-rest-of-line-muted-chroma-tokens-on-tinted-backgrounds-unified-and-split
plan: 11-04
subsystem: highlight
tags: [word-diff, chroma, lipgloss, uat-gap]

requires: []
provides:
  - Muted full-line diff background vs brighter semantic word-span backgrounds
  - Removal of neutral gutterTintStyle for changed segments
affects: []

tech-stack:
  added: []
  patterns:
    - DiffLineMutedBackgroundColour(0.42 toward terminal base); WordSpanBackgroundColour(0.32 toward pure R/G)

key-files:
  created: []
  modified:
    - internal/highlight/diffcolors.go
    - internal/highlight/diff_line.go
    - internal/highlight/diffcolors_test.go
    - internal/render/wordline.go
    - internal/render/gutter.go
    - internal/render/gutter_test.go
    - .planning/phases/.../11-VERIFICATION.md
    - .planning/phases/.../11-CONTEXT.md

key-decisions:
  - D-COLOR-04 documents brighter word spans vs muted line (11-UAT gap)

requirements-completed: []

duration: ""
completed: "2026-03-26"
---

# Plan 11-04 — Brighter semantic word spans vs muted full-line

**Word-diff changed segments now use `WordSpanBackgroundColour` (blend toward pure red/green); full-line wrap uses `DiffLineMutedBackgroundColour`. Neutral `gutterTintStyle` removed.**

## Self-Check: PASSED

- `go test ./...` and `just lint` pass.
