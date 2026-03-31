---
phase: 6
slug: api-hardening-oss-packaging
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-25
---

# Phase 6 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` / `go test` |
| **Config file** | none — existing repo |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test ./...` and `go test -bench=. -benchmem -run '^$' ./...` (benchmarks only when bench plan landed) |
| **Estimated runtime** | ~30–120 seconds (includes benchmarks) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test ./...`; after wave containing benchmarks, add `go test -bench=. -run '^$' -count=1` for touched packages
- **Before `/gsd-verify-work`:** Full suite green
- **Max feedback latency:** 120 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 06-01-01 | 01 | 1 | CORE-05 | unit | `go test ./... -run TestBuilder` | ⬜ W1 | ⬜ pending |
| 06-02-01 | 02 | 1 | OSS-02 | manual/script | `go doc -all github.com/tbcrawford/drift` | ⬜ W1 | ⬜ pending |
| 06-03-01 | 03 | 2 | OSS-03 | integration | `go run ./examples/basic/` | ⬜ W2 | ⬜ pending |
| 06-03-02 | 03 | 2 | OSS-03 | integration | `go run ./examples/builder/` | ⬜ W2 | ⬜ pending |
| 06-04-01 | 04 | 2 | OSS-07 | bench | `go test -bench=BenchmarkDrift -count=1 ./...` | ⬜ W2 | ⬜ pending |
| 06-05-01 | 05 | 3 | OSS-06 | manual | `test -f README.md && grep -q 'go install' README.md` | ⬜ W3 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- Existing infrastructure covers all phase requirements — no Wave 0 stubs.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| README readability | OSS-06 | Subjective layout | Open README; confirm five required sections exist with code fences |
| godoc rendering on pkg.go.dev | OSS-02 | External site | After tag/push, spot-check package page |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 120s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
