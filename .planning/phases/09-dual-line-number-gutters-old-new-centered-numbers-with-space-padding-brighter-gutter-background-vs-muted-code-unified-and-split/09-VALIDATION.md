---
phase: 9
slug: dual-line-number-gutters-old-new-centered-numbers-with-space-padding-brighter-gutter-background-vs-muted-code-unified-and-split
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-26
---

# Phase 9 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — stdlib testing |
| **Quick run command** | `go test ./internal/render/... -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~10 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/render/... -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 9-01-01 | 01 | 1 | Phase 9 | unit | `go test ./internal/render/... -run TestCenter -count=1` | ❌ W0 | ⬜ pending |
| 9-01-02 | 01 | 1 | Phase 9 | unit | `go test ./internal/render/... -run TestUnified.*Line -count=1` | ❌ W0 | ⬜ pending |
| 9-02-01 | 02 | 2 | Phase 9 | unit | `go test ./internal/render/... -run TestSplit.*Line -count=1` | ❌ W0 | ⬜ pending |
| 9-02-02 | 02 | 2 | Phase 9 | integration | `go test ./... -run TestRender.*LineNumber -count=1` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/render/unified_test.go` — extended fixtures asserting line-number columns for unified output
- [ ] `internal/render/split_test.go` — fixtures asserting gutters in split output
- [ ] `render_test.go` — `drift.Render` integration with `WithLineNumbers` / default behavior

*Existing `go test` infrastructure covers execution — extend tests only.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Gutter background vs code contrast (TrueColor) | Phase 9 | Subjective luminance | Run `go run ./cmd/drift/...` on two small files; confirm gutters visually distinct from code in unified and `--split` |
| Readability on light terminal | Phase 9 | Adaptive colors need eyeball | Toggle macOS light/dark or light theme; gutters remain legible |

---

## Validation Sign-Off

- [ ] All tasks have `<acceptance_criteria>` with grep or `go test` commands
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
