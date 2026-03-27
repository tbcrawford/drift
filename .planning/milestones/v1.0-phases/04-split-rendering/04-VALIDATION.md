---
phase: 4
slug: split-rendering
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-25
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — stdlib testing |
| **Quick run command** | `go test ./internal/render/... -run TestSplit` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/render/... -run TestSplit`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 4-01-01 | 01 | 1 | REND-02 | unit | `go test ./internal/render/... -run TestSplitRenderer` | ❌ W0 | ⬜ pending |
| 4-01-02 | 01 | 1 | REND-02 | unit | `go test ./internal/render/... -run TestANSIWidth` | ❌ W0 | ⬜ pending |
| 4-02-01 | 02 | 1 | REND-02 | unit | `go test ./internal/render/... -run TestTermWidth` | ❌ W0 | ⬜ pending |
| 4-02-02 | 02 | 1 | REND-02 | unit | `go test ./internal/render/... -run TestPipeWidth` | ❌ W0 | ⬜ pending |
| 4-03-01 | 03 | 2 | REND-02 | integration | `go test ./... -run TestDiffSplitOption` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/render/split_test.go` — stubs for REND-02 split renderer, ANSI width, panel layout
- [ ] `internal/render/termwidth_test.go` — terminal width detection and pipe fallback
- [ ] `drift_test.go` or `diff_test.go` — `WithSplit()` integration tests

*Existing `go test` infrastructure covers all phase requirements — no new framework install needed.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Visual panel alignment at 80/120/200 cols | REND-02 | ANSI visual output cannot be fully verified programmatically | Run `go run ./cmd/drift diff testdata/a.txt testdata/b.txt --split` at each terminal width; confirm panels are equal width with no overflow |
| Syntax highlighting preserved in both panels | REND-02 | Chroma color codes require visual inspection | Diff two Go files; confirm both panels show colored keywords, not raw ANSI escape strings |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
