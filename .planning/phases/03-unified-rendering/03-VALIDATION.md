---
phase: 3
slug: unified-rendering
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-25
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — stdlib testing |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test -race ./...` |
| **Estimated runtime** | ~10 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test -race ./...`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** ~10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 3-01-01 | 01 | 1 | REND-07 | unit | `go test ./internal/theme/...` | ❌ W0 | ⬜ pending |
| 3-02-01 | 02 | 1 | REND-03, REND-05 | unit | `go test ./internal/highlight/...` | ❌ W0 | ⬜ pending |
| 3-03-01 | 03 | 1 | REND-04, REND-06 | unit | `go test ./internal/highlight/... -run TestDetectLexer` | ❌ W0 | ⬜ pending |
| 3-04-01 | 04 | 2 | REND-01, REND-08 | unit | `go test ./internal/render/...` | ❌ W0 | ⬜ pending |
| 3-05-01 | 05 | 2 | REND-05, REND-09 | integration | `go test . -run TestRender` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/theme/theme_test.go` — stubs for REND-07 dark/light detection
- [ ] `internal/highlight/highlight_test.go` — stubs for REND-03, REND-04, REND-05, REND-06
- [ ] `internal/render/unified_test.go` — stubs for REND-01, REND-08
- [ ] `render_test.go` — integration test stubs for REND-05, REND-09
- [ ] `go get charm.land/lipgloss/v2 github.com/alecthomas/chroma/v2 github.com/charmbracelet/colorprofile` — new dependencies

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Dark terminal auto-detects dark theme (monokai) | REND-07 | Requires real TTY with OSC 11 response | Run `drift` CLI in dark terminal; verify highlighted output uses warm/dark color palette |
| Light terminal auto-detects light theme (github) | REND-07 | Requires real TTY with light background | Run `drift` CLI in light terminal; verify highlighted output uses light/pastel palette |
| 16-color terminal receives degraded ANSI output | REND-05 | Requires real 16-color terminal env | Set `TERM=xterm`, unset `COLORTERM`; run `drift`; verify `\033[3Xm` sequences present, no `38;2;` TrueColor |
| Go tokens visually distinct in terminal output | REND-03 | Subjective visual assessment | Diff two `.go` files; verify `func` keyword, string literals, and identifiers have distinct colors |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
