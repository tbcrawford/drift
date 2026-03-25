---
phase: 7
slug: support-diffs-from-git-that-is-if-a-single-file-is-provided-and-the-file-is-in-a-git-repo-drift-will-show-the-current-changes
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-25
---

# Phase 7 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` package |
| **Config file** | none — uses existing module |
| **Quick run command** | `go test ./cmd/drift/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~10–30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./cmd/drift/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 07-01-01 | 01 | 1 | Phase 7 git helper | unit | `go test ./cmd/drift/... -run Git` | ⬜ W0 | ⬜ pending |
| 07-02-01 | 02 | 2 | CLI wiring | unit | `go test ./cmd/drift/...` | ⬜ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- Existing `cmd/drift` test layout covers resolver tests; no new framework install.
- Fake-`git` tests use a temp directory + `PATH` prepend (documented in plan 07-01).

*If none: "Existing infrastructure covers all phase requirements."*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Real repo UX | Phase 7 goal | Requires real Git state | Clone or use local repo; modify one file; run `drift ./path`; eyeball vs `git diff HEAD -- path` |

*If none: "All phase behaviors have automated verification."*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
