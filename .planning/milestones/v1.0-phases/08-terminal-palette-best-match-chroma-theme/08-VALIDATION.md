---
phase: 8
slug: terminal-palette-best-match-chroma-theme
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-25
---

# Phase 8 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` + stdlib |
| **Config file** | none |
| **Quick run command** | `go test ./internal/highlight/...` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~30–60 seconds |

---

## Sampling Rate

- **After every task commit:** `go test ./internal/highlight/...` (wave 1) or `go test ./cmd/drift/...` (wave 2)
- **After every plan wave:** `go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite green

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | Status |
|---------|------|------|-------------|-----------|-------------------|--------|
| 08-01-01 | 01 | 1 | REND-04 | unit | `go test ./internal/highlight/... -run BestMatch` | ⬜ |
| 08-01-02 | 01 | 1 | REND-04 | unit | `go test ./internal/highlight/... -run OSC4` | ⬜ |
| 08-02-01 | 02 | 2 | REND-04 | unit + integration | `go test ./cmd/drift/...` | ⬜ |
| 08-02-02 | 02 | 2 | REND-08 | unit | `go test ./... -count=1` | ⬜ |

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements — no Wave 0 stubs.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|---------------------|
| OSC 4 theme matches terminal emulator | REND-04 | Needs real TTY + palette | Run `drift --show-theme` on macOS/Linux in iTerm vs Terminal; compare stderr theme name |

---

## Validation Sign-Off

- [x] All tasks have automated verify or manual table
- [ ] `nyquist_compliant: true` after execution review

**Approval:** pending
