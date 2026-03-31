---
phase: 5
slug: cli
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-03-25
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` + `github.com/rogpeppe/go-internal/testscript` (integration) |
| **Config file** | none — standard `go test` |
| **Quick run command** | `go test ./cmd/drift/...` |
| **Full suite command** | `just test` (repo root: `go test ./...`) |
| **Estimated runtime** | ~30–90 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./cmd/drift/...` (or `go test ./...` if shared packages touched)
- **After every plan wave:** Run `just test`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 120 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 05-01-01 | 01 | 1 | CLI-04, CLI-05, CLI-06 | unit | `go test ./cmd/drift/...` | ✅ | ⬜ pending |
| 05-02-01 | 02 | 2 | CLI-01, CLI-02, CLI-03 | unit | `go test ./cmd/drift/...` | ✅ | ⬜ pending |
| 05-03-01 | 03 | 3 | CLI-07 | unit | `go test ./cmd/drift/...` | ✅ | ⬜ pending |
| 05-04-01 | 04 | 4 | OSS-04 | integration | `go test ./cmd/drift/... -run Script` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] Existing `go test ./...` and `justfile` cover the module; no new framework install required.
- [ ] `testscript` dependency added in plan 05-04 when integration scripts are introduced.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `go install …@latest` on unpublished module | OSS-04 | Requires `GOPROXY` or local replace during dev | After tag/publish: run `go install github.com/tbcrawford/drift/cmd/drift@latest` from a temp env; verify `drift --help` |

*Automated path: `go build` / `go test` with `testscript` using locally built binary.*

---

## Validation Sign-Off

- [ ] All tasks have `<acceptance_criteria>` with `go test` or `testscript` commands
- [ ] Sampling continuity: tests run after each wave
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 120s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
