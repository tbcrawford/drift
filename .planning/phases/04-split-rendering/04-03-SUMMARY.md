---
phase: 04-split-rendering
plan: "03"
subsystem: api
tags: [WithSplit, Render, options]

requires:
  - phase: 04-01
    provides: render.Split, RenderConfig.TermWidth
  - phase: 04-02
    provides: render.TerminalWidth
provides:
  - drift.WithSplit() Option
  - Render / RenderWithNames route to Split when split enabled
affects: [Phase 5 CLI]

tech-stack:
  added: []
  patterns: [term width resolved from destination writer before each render]

key-files:
  created: []
  modified: [options.go, render.go, render_test.go]

key-decisions:
  - "Split uses same colorprofile.NewWriter path as unified; TermWidth from raw w before wrap"

patterns-established: []

requirements-completed: [REND-02]

duration: 10min
completed: 2026-03-25
---

# Phase 4 Plan 04-03 Summary

**Public API:** `WithSplit()` sets `config.split`; `Render` and `RenderWithNames` call `TerminalWidth(w)`, set `rcfg.TermWidth`, and dispatch to `render.Split` or `render.Unified`. Integration tests use `WithNoColor` and assert `|` separator and panel content.

## Task Commits

1. **04-03-01** — `ea7b480` feat(drift): WithSplit
2. **04-03-02** — `2d82958` feat(drift): wire Split in Render
3. **04-03-03** — `87a1a69` test(drift): WithSplit integration

## Self-Check: PASSED

- `go test ./... -count=1` and `-race` pass
