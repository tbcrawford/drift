---
phase: 05-cli
plan: "02"
subsystem: cli
tags: [input, stdin, flags]

requires:
  - plan: 05-01
    provides: cobra root and flags
provides:
  - resolveInputs for positional paths, '-', and --from/--to
affects: [05-03]

tech-stack:
  added: []
  patterns: [mutually exclusive positionals vs from/to]

key-files:
  created: [cmd/drift/input.go, cmd/drift/input_test.go]
  modified: [cmd/drift/main.go]

key-decisions:
  - "RunE validates inputs then returns stub until 05-03 render wiring"

patterns-established: []

requirements-completed: [CLI-01, CLI-02, CLI-03]

duration: —
completed: 2026-03-25
---

# Phase 5 Plan 05-02 Summary

Implemented `resolveInputs` with `os.ReadFile` / `io.ReadAll`, `drift - -` single stdin read, `--from`/`--to` with names `a/string`/`b/string`, and errors containing `invalid`/`usage` for misuse. `RunE` calls resolver then returns `not implemented: complete plan 05-03`.

## Task Commits

1. **05-02-01** — `d38538e` feat(cli): resolve stdin, files, and --from/--to (05-02)

## Self-Check: PASSED

- `go test ./cmd/drift/... -run TestResolve`
- `grep` for `os.ReadFile` / `io.ReadAll` in `input.go`
