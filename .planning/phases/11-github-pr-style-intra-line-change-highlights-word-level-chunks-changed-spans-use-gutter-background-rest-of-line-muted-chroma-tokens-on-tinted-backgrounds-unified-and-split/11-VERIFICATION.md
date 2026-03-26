# Phase 11 verification

status: passed (automated); human_verification pending (11-03 visual parity)

## Checks

### Wave 1–2 (11-01 / 11-02) — passed

- [x] Split: paired delete/insert with `WordDiff` + color profile produces segmented ANSI output
- [x] Unified: consecutive `-` / `+` lines use `PairSegments` when `WordDiff` + color profile
- [x] `WithWordDiff` / builder / `RenderConfig.WordDiff` wired from `drift.Render`
- [x] `go test ./...` and `just lint` pass

### Plan 11-03 — terrasort parity + layering (executed; automated checks below)

- [x] **Line colours:** `DiffLineBackgroundColour` matches terrasort `chromaDiffLineRGBA` pipeline (`GenericInserted`/`GenericDeleted`, `blendChromaTowardTerminalBase`, `fallbackDiffChroma` hexes aligned with terrasort `fallbackDiffRGBA`) — `11-03-PLAN.md` T1
- [x] **Word spans / line-number gutters:** `gutterTintStyle` and `gutterStyleForCell` use **gutter neutrals** only (`GutterBackgroundHex`); semantic line backgrounds apply on full code lines via `ApplyDiffLineStyle` — T2
- [x] **Full line:** word-diff unified + split apply full-line wrap after segmentation (`ApplyDiffLineStyle` / `splitApplyDiffLine`) — T3
- [x] Tests extended per T4 (`TestTerrasortParity_*`, `TestGutterNeutralTint_*`, word-diff CSI 48 assertion); `go test ./...` and `just lint` pass

#### Checks (11-03 — pending human sign-off)

- [ ] Terrasort parity: line RGB pipeline matches roles in a true-color terminal (compare drift vs terrasort on the same pair).
- [ ] Gutter-only neutral for intra-line changed words; no standalone semantic red/green on word spans alone.
- [ ] Full-line wrap on word-diff rows (prefix + body) in unified and split.

## human_verification

- [ ] Optional (11-02): run `drift` on a small Go file pair in a true-color terminal and confirm muted vs gutter-tinted word spans.
- [ ] **11-03:** Side-by-side: terrasort diff vs drift on the same pair — **line** and **intra-line** colours should match **roles** (terrasort palette source of truth per `11-CONTEXT.md`). Mark the three 11-03 checks above after visual confirmation.
