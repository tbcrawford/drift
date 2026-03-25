---
phase: 06-api-hardening-oss-packaging
plan: "01"
subsystem: api
tags: [go, builder, options]

requires: []
provides:
  - Chainable Builder API mirroring all With* options
affects: [06-02, 06-03, 06-05]

tech-stack:
  added: []
  patterns:
    - "Builder holds []Option and delegates to Diff/Render/RenderWithNames"

key-files:
  created:
    - builder.go
    - builder_test.go
  modified:
    - doc.go

key-decisions:
  - "Exported Builder type with pointer receivers; New() returns empty option slice"

patterns-established:
  - "Fluent methods append same Option closures as With* functions"

requirements-completed: ["CORE-05"]

duration: 10min
completed: 2026-03-25
---

# Phase 6: API hardening — Plan 01 Summary

**Introduced `drift.New()` builder with full parity tests against functional options.**

## Performance

- **Tasks:** 1
- **Files modified:** 3

## Accomplishments

- `Builder` chains Algorithm, Context, NoColor, Lang, Theme, Split; terminal Diff/Render/RenderWithNames delegate to package functions.
- `TestBuilderDiffParity` and `TestBuilderRenderParity` lock behavior to the option API.

## Task Commits

1. **06-01-01** — `a3f1ddc` (feat)

## Files Created/Modified

- `builder.go` — Builder type and chain methods
- `builder_test.go` — Parity tests
- `doc.go` — Note on builder alternative

## Self-Check: PASSED

- `builder.go`, `builder_test.go` exist
- Commit grep `06-01` present in message
