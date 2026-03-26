# Phase 11 verification

status: pending (11-03 gap closure)

## Checks

### Wave 1–2 (11-01 / 11-02) — passed

- [x] Split: paired delete/insert with `WordDiff` + color profile produces segmented ANSI output
- [x] Unified: consecutive `-` / `+` lines use `PairSegments` when `WordDiff` + color profile
- [x] `WithWordDiff` / builder / `RenderConfig.WordDiff` wired from `drift.Render`
- [x] `go test ./...` and `just lint` pass

### Plan 11-03 — terrasort parity + layering (pending execution)

- [ ] **Line colours:** `DiffLineBackgroundColour` matches terrasort `chromaDiffLineRGBA` pipeline (`GenericInserted`/`GenericDeleted`, blend, fallbacks) — see `11-03-PLAN.md` T1
- [ ] **Word spans:** intra-line changed spans use **gutter neutrals** only (`GutterBackgroundHex`), not semantic red/green on words alone — T2
- [ ] **Full line:** word-diff unified + split apply full-line `ApplyDiffLineStyle` / `splitApplyDiffLine` after segmentation — T3
- [ ] Tests extended per T4; `go test ./...` and `just lint` pass

## human_verification

- [ ] Optional (11-02): run `drift` on a small Go file pair in a true-color terminal and confirm muted vs gutter-tinted word spans.
- [ ] **11-03:** Side-by-side: terrasort diff vs drift on the same pair — **line** and **intra-line** colours should match **roles** (terrasort palette source of truth per `11-CONTEXT.md`).
