---
phase: 11
slug: github-pr-style-intra-line-change-highlights-word-level-chunks-changed-spans-use-gutter-background-rest-of-line-muted-chroma-tokens-on-tinted-backgrounds-unified-and-split
status: draft
nyquist_compliant: true
wave_0_complete: true
created: "2026-03-26"
---

# Phase 11 — Validation Strategy

> Feedback sampling for terrasort color parity + word-diff layering (`11-RESEARCH.md`, `11-03-PLAN.md`).

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` + testify where used |
| **Config file** | `.golangci.yml` |
| **Quick run command** | `go test ./internal/highlight/... ./internal/render/... ./internal/worddiff/...` |
| **Full suite command** | `go test ./...` |
| **Lint** | `just lint` or `golangci-lint run ./...` |
| **Estimated runtime** | &lt; 30 s full suite |

---

## Sampling Rate

- **After every task commit:** `go test ./internal/highlight/... ./internal/render/...`
- **After plan wave:** `go test ./...` and `just lint`
- **Before verify-work:** Full suite green

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Test Type | Automated Command | Status |
|---------|------|------|-------------|-------------------|--------|
| 11-03-T1 | 11-03 | 1 | unit | `go test ./internal/highlight/...` | ⬜ |
| 11-03-T2 | 11-03 | 1 | unit + grep | `go test ./internal/render/...`; `grep DiffLineBackgroundColour internal/render/gutter.go` must not match `gutterTintStyle` body | ⬜ |
| 11-03-T3 | 11-03 | 1 | unit + grep | `go test ./internal/render/...`; grep `ApplyDiffLineStyle` / `splitApplyDiffLine` in `unified.go` / `wordline.go` | ⬜ |
| 11-03-T4 | 11-03 | 1 | unit | `go test ./internal/highlight/... ./internal/render/...` | ⬜ |
| 11-03-T5 | 11-03 | 1 | doc | `test -f` phase `11-VERIFICATION.md` | ⬜ |

---

## Wave 0 Requirements

Existing Go module + tests cover requirements — no Wave 0 stub layer.

---

## Manual-Only Verifications

| Behavior | Why manual | Test instructions |
|----------|------------|-------------------|
| Visual match drift vs terrasort | ANSI + terminal palette | Same file pair in terrasort CLI vs `drift`; compare line vs word emphasis |
| GitHub PR “feel” | Subjective hierarchy | Open GitHub PR diff in browser; compare full-line + word emphasis to drift output |

---

## Validation Sign-Off

- [x] Research produced `11-RESEARCH.md` with Validation Architecture
- [ ] All 11-03 tasks green
- [ ] `nyquist_compliant: true` reviewed after execution

**Approval:** pending
