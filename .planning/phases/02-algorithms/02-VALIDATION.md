---
phase: 2
slug: algorithms
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-25
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — standard Go test toolchain |
| **Quick run command** | `go test ./internal/algo/...` |
| **Full suite command** | `go test ./... && just test` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/algo/...`
- **After every plan wave:** Run `go test ./... && just test`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 02-02-01 | 02 | 1 | ALGO-01 | unit | `go test ./internal/algo/patience/...` | ❌ W0 | ⬜ pending |
| 02-02-02 | 02 | 1 | ALGO-01 | unit | `go test ./internal/algo/patience/...` | ❌ W0 | ⬜ pending |
| 02-03-01 | 03 | 1 | ALGO-02 | unit | `go test ./internal/algo/histogram/...` | ❌ W0 | ⬜ pending |
| 02-03-02 | 03 | 1 | ALGO-02, ALGO-03 | unit | `go test ./internal/algo/histogram/...` | ❌ W0 | ⬜ pending |
| 02-04-01 | 04 | 2 | ALGO-04 | integration | `go test ./...` | ❌ W0 | ⬜ pending |
| 02-04-02 | 04 | 2 | ALGO-01, ALGO-02, ALGO-03, ALGO-04 | property | `go test -run TestRoundTrip ./...` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/algo/patience/patience_test.go` — stubs for ALGO-01
- [ ] `internal/algo/histogram/histogram_test.go` — stubs for ALGO-02, ALGO-03
- [ ] Integration of `WithAlgorithm()` round-trip test stubs in existing property test file

*Existing infrastructure covers Myers; new packages need test stubs.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Patience produces visually superior diff on refactored code | ALGO-01 | Output quality is subjective | Run `go run . --algorithm patience` on a known refactored Go file; verify moved blocks are grouped |
| Histogram produces better hunk boundaries on repetitive files | ALGO-02 | Requires visual inspection | Diff two versions of a file with repeated struct definitions; verify histogram groups them better than Myers |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
