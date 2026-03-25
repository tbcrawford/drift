---
plan: "03-01"
title: "Safe OSC 11 dark background detection"
status: "completed"
completed_at: "2026-03-25"
---

# Plan 03-01 Summary: Safe OSC 11 Dark Background Detection

## Objective

Implemented `internal/theme/` with a `DetectDarkBackground(profile colorprofile.Profile) bool` function that safely guards the Lip Gloss v2 OSC 11 terminal query behind a TTY/color profile check, preventing the 2-second hang in piped/non-TTY environments.

## Tasks Completed

### Task 03-01-01: Add charm.land/lipgloss/v2 and colorprofile dependencies
- Added `charm.land/lipgloss/v2 v2.0.2` to `go.mod`
- Added `github.com/charmbracelet/colorprofile v0.4.2` to `go.mod`
- `go build ./...` exits 0
- **Commit**: `feat(deps): add charm.land/lipgloss/v2 v2.0.2 and colorprofile v0.4.2`

### Task 03-01-02: Implement internal/theme/theme.go with safe DetectDarkBackground
- Created `internal/theme/theme.go` with `DetectDarkBackground(profile colorprofile.Profile) bool`
- Short-circuits to `true` (dark default) for `colorprofile.NoTTY` and `colorprofile.Ascii`
- Calls `lipgloss.HasDarkBackground(os.Stdin, os.Stdout)` only for color-capable TTY profiles
- `go build ./...` exits 0
- **Commit**: `feat(theme): implement DetectDarkBackground with OSC 11 timeout guard`

### Task 03-01-03: Write unit tests for DetectDarkBackground
- Created `internal/theme/theme_test.go` with three test functions:
  - `TestDetectDarkBackground_NoTTY`: verifies `NoTTY` and `Ascii` profiles return `true` immediately
  - `TestDetectDarkBackground_TrueColor`: verifies TrueColor does not short-circuit (no panic)
  - `TestDetectDarkBackground_ANSI256`: mirrors TrueColor test
- `go test ./internal/theme/...` exits 0
- `go test -race ./internal/theme/...` exits 0
- **Commit**: `test(theme): add unit tests for DetectDarkBackground NoTTY/TrueColor/ANSI256`

## Files Modified

| File | Change |
|------|--------|
| `go.mod` | Added `charm.land/lipgloss/v2 v2.0.2` and `github.com/charmbracelet/colorprofile v0.4.2` |
| `go.sum` | Updated with new dependency checksums |
| `internal/theme/theme.go` | Created — `DetectDarkBackground` implementation |
| `internal/theme/theme_test.go` | Created — unit tests for NoTTY/Ascii short-circuit and TTY paths |

## Key Decisions

- **OSC 11 guard**: `colorprofile.NoTTY` and `colorprofile.Ascii` short-circuit to `true` (dark default) — avoids any terminal escape sequence emission on piped outputs
- **Dark default**: `true` is the safe fallback because most developer terminals are dark-themed, and Chroma dark themes look better on light terminals (minor rendering degradation) than vice versa
- **Import path**: `charm.land/lipgloss/v2` (NOT `github.com/charmbracelet/lipgloss`) — consistent with project STACK.md; v2 changed canonical import path

## Acceptance Criteria Verification

- [x] `internal/theme/theme.go` exports `func DetectDarkBackground(profile colorprofile.Profile) bool`
- [x] `DetectDarkBackground` short-circuits to `true` when profile is `colorprofile.NoTTY` or `colorprofile.Ascii`
- [x] `DetectDarkBackground` calls `lipgloss.HasDarkBackground(os.Stdin, os.Stdout)` for TTY-capable profiles
- [x] `go test ./internal/theme/...` exits 0
- [x] `go build ./...` exits 0
