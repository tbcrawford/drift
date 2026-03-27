---
status: partial
phase: 11-github-pr-style-intra-line-change-highlights-word-level-chunks-changed-spans-use-gutter-background-rest-of-line-muted-chroma-tokens-on-tinted-backgrounds-unified-and-split
source:
  - 11-01-SUMMARY.md
  - 11-02-SUMMARY.md
  - 11-03-SUMMARY.md
  - 11-VERIFICATION.md
started: "2026-03-26T12:00:00Z"
updated: "2026-03-26T18:30:00Z"
---

## Current Test

[testing paused — product owner rejected assumptions in tests 1–2 / 4; see issues and Gaps]

## Tests

### 1. Unified word-diff — layering in a true-color terminal
expected: (UAT assumed neutral gutter tint on changed words — **rejected by product owner**)
result: issue
reported: "Changed word spans should not use a neutral gray gutter-style tint. They should use red or green appropriately and be highlighted in a brighter red/green than the entire line itself, which should be more muted. I reject verification of this."
severity: major

### 2. Split view — word-diff + full-line wrap
expected: (Same wrong assumption — neutral tints on changed spans)
result: issue
reported: "Superseded by test 1 — product requires brighter semantic word-level red/green vs muted full line, not neutral intra-line tints."
severity: major

### 3. Line-number gutters stay neutral
expected: On delete/insert rows, old/new line number columns use neutral grays; semantic colouring on code line.
result: skipped
reason: "Not evaluated; intra-line/word-span colour model is being redesigned per owner feedback (tests 1–2)."

### 4. Terrasort colour-role parity (if available)
expected: (Assumed terrasort-style neutral intra-line roles)
result: issue
reported: "Owner rejects verification framed around terrasort neutral word spans; target UX is brighter word-level semantic colour vs muted full line."
severity: major

## Summary

total: 4
passed: 0
issues: 3
pending: 0
skipped: 1
blocked: 0

## Gaps

- truth: "Intra-line changed spans use semantic remove (red) / add (green) highlighting; word-level highlights are brighter/more saturated than the full-line background, which remains the muted semantic plane. Unchanged tokens stay muted (e.g. comment-style)."
  status: failed
  reason: "Product owner: rejects neutral gray gutter-style tint on changed words; requires appropriate red/green on changed spans, brighter than the (more muted) full-line background."
  severity: major
  test: 1
  artifacts: []
  missing:
    - "Replace neutral gutterTintStyle for changed spans with brighter semantic delete/insert colours derived from theme"
    - "Keep full-line DiffLineStyle as the muted base; layer brighter word spans on top (unified + split)"
  root_cause: "Phase 11 implementation and 11-03 plan followed terrasort/D-COLOR-03 'gutter neutrals only' for intra-line emphasis; this contradicts the owner's stated GitHub PR–style expectation (brighter word-level red/green vs muted line)."
  debug_session: ""
