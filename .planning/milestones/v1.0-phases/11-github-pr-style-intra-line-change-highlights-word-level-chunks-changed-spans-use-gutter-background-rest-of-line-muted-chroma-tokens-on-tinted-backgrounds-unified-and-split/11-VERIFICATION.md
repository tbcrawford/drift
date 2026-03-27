# Phase 11 verification

status: passed (automated); human_verification pending (visual confirmation of 11-04 word vs line contrast)

## Checks

### Wave 1–2 (11-01 / 11-02) — passed

- [x] Split: paired delete/insert with `WordDiff` + color profile produces segmented ANSI output
- [x] Unified: consecutive `-` / `+` lines use `PairSegments` when `WordDiff` + color profile
- [x] `WithWordDiff` / builder / `RenderConfig.WordDiff` wired from `drift.Render`
- [x] `go test ./...` and `just lint` pass

### Plan 11-03 — terrasort line pipeline + full-line wrap (executed)

- [x] **Line colours:** `DiffLineBackgroundColour` matches terrasort `chromaDiffLineRGBA` pipeline — `11-03-PLAN.md` T1
- [x] **Line-number gutters:** `gutterStyleForCell` uses **gutter neutrals** only (`GutterBackgroundHex`) — T2
- [x] **Full line:** word-diff unified + split apply full-line wrap after segmentation — T3

### Plan 11-04 — brighter **WordSpan** vs muted full line (gap closure from `11-UAT.md`)

- [x] **`DiffLineMutedBackgroundColour`** + **`WordSpanBackgroundColour`** in `internal/highlight/diffcolors.go`; **`DiffLineStyle`** uses muted full-line plane
- [x] **`wordSpanStyle`** / **`WordSpanBackgroundColour`** replace neutral **`gutterTintStyle`** on changed segments in `wordline.go`
- [x] Tests: **`TestWordSpanBrighterThanMutedLine_githubDark`**, `TestWordSpanStyle_delete_hasBackgroundANSI`
- [x] `go test ./...` and `just lint` pass

#### Checks (pending human sign-off)

- [ ] In a true-color terminal, **changed words** read as **brighter red/green** than the **muted** full-line wash on `-`/`+` rows (unified + split).
- [ ] **Line-number columns** remain neutral gray; semantic colour is on code + word spans as designed.

## human_verification

- [ ] Run `drift` on a small paired change with word diff: confirm **word-level** semantic highlights **pop** above the **full-line** background (not gray word chips).
- [ ] Re-run `/gsd-verify-work 11` or update `11-UAT.md` after visual confirmation.
