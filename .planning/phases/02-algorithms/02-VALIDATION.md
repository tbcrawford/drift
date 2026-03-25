---
phase: 2
slug: algorithms
status: draft
nyquist_compliant: true
wave_0_complete: true
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
| 02-01-01 | 02-01 | 1 | ALGO-01 | unit | `go test ./internal/algo/patience/... -run TestBothEmpty` | ❌ W0 | ⬜ pending |
| 02-01-02 | 02-01 | 1 | ALGO-01 | unit | `go test ./internal/algo/patience/... -race` | ❌ W0 | ⬜ pending |
| 02-02-01 | 02-02 | 1 | ALGO-02 | unit | `go test ./internal/algo/histogram/... -run TestBothEmpty` | ❌ W0 | ⬜ pending |
| 02-02-02 | 02-02 | 1 | ALGO-02 | unit | `go test ./internal/algo/histogram/... -race` | ❌ W0 | ⬜ pending |
| 02-03-01 | 02-03 | 2 | ALGO-01, ALGO-02, ALGO-03 | integration | `go build ./...` | ✅ exists | ⬜ pending |
| 02-03-02 | 02-03 | 2 | ALGO-01, ALGO-02, ALGO-03 | integration | `go test -run TestAllAlgorithmsCorrect ./...` | ❌ W0 | ⬜ pending |
| 02-04-01 | 02-04 | 2 | ALGO-04 | property | `go test -run TestProperty_RoundTrip_Patience ./... -race` | ❌ W0 | ⬜ pending |
| 02-04-02 | 02-04 | 2 | ALGO-01, ALGO-02, ALGO-03, ALGO-04 | property | `go test -run TestProperty ./... -race` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `internal/algo/patience/patience_test.go` — minimal `TestBothEmpty` created alongside `patience.go` in task 02-01-01; full suite expanded in 02-01-02
- [x] `internal/algo/histogram/histogram_test.go` — minimal `TestBothEmpty` created alongside `histogram.go` in task 02-02-01; full suite expanded in 02-02-02
- [x] Integration of `WithAlgorithm()` round-trip test stubs in existing property test file — covered by 02-03-02 (`TestAllAlgorithmsCorrect`) and 02-04-01 (property tests)

*Existing infrastructure covers Myers; new packages create test files in their respective implementation tasks.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Patience produces visually superior diff on refactored code | ALGO-01 | Output quality is subjective | Run `go run . --algorithm patience` on a known refactored Go file; verify moved blocks are grouped |
| Histogram produces better hunk boundaries on repetitive files | ALGO-02 | Requires visual inspection | Diff two versions of a file with repeated struct definitions; verify histogram groups them better than Myers |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 10s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
