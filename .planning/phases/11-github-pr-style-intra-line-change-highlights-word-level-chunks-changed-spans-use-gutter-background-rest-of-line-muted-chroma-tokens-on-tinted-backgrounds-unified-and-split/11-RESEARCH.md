# Phase 11 — Technical research

**Question:** What must drift implement so **line** and **word-span** colors match **terrasort**’s `DeriveUXTheme` / `diffcolors` roles while keeping **`internal/worddiff`** segmentation, and how does that interact with **full-line** Lip Gloss wrapping?

**Sources read (2026-03-26):**

- `11-CONTEXT.md` (D-COLOR-01–03, D-LAYER-01/02)
- `internal/highlight/diffcolors.go`, `internal/highlight/diff_line.go`
- `internal/render/gutter.go`, `internal/render/wordline.go`, `internal/render/unified.go`
- `/Users/tylercrawford/dev/github/terrasort/internal/highlight/diffcolors.go`
- `/Users/tylercrawford/dev/github/terrasort/internal/highlight/uxtheme.go` (AddBg/RemoveBg wiring)

---

## 1. Terrasort line colour pipeline (canonical per CONTEXT)

`DeriveUXTheme` sets:

- `addBg := chromaDiffLineRGBA(style, isDark, false)`
- `removeBg := chromaDiffLineRGBA(style, isDark, true)`

`chromaDiffLineRGBA`:

1. Resolve `GenericInserted` or `GenericDeleted` from the active `*chroma.Style`.
2. If `e.Background.IsSet()` → use that RGB.
3. Else if `e.Colour.IsSet()` → `blendChromaTowardTerminalBase` (mix **0.78** toward `(18,18,22)` dark or `(255,255,255)` light).
4. Else unset → `fallbackDiffRGBA` (`#3a2228` / `#243520` dark del/ins; `#ffeaea` / `#e6f7e6` light).

**Note:** This is **not** a blend from `chroma.Background` toward pure red/green.

---

## 2. Drift line colour pipeline (current)

`DiffLineBackgroundColour` in drift (`internal/highlight/diffcolors.go`):

1. Reads `style.Get(chroma.Background).Background`.
2. Blends toward **pure** `(255,0,0)` or `(0,255,0)` with α **0.15** (dark) or **0.12** (light).
3. Fallbacks match terrasort hex fallbacks.

**Gap:** Drift and terrasort **disagree on the algorithm** for non-fallback paths. For the same theme + `isDark`, RGB outputs will **not** match terrasort `AddBg`/`RemoveBg` unless the theme’s `Background` blend happens to coincide with gd/gi-derived colours (unlikely).

---

## 3. Terrasort gutter / intra-line neutrals

`driftStyleGutterRGBA` / drift `GutterBackgroundHex`:

| Mode   | Old column | New column |
|--------|------------|------------|
| Dark   | `#585858` (ANSI 240) | `#444444` (ANSI 238) |
| Light  | `#e4e4e4` (ANSI 254) | `#eeeeee` (ANSI 255) |

Terrasort sets `IntraAddBg` / `IntraRemoveBg` equal to gutter add/remove backgrounds — **intra-line emphasis uses the same neutral column greys**, not semantic red/green on the span alone.

---

## 4. Drift render layering (current code audit)

| Concern | Status |
|---------|--------|
| **Word-diff changed spans** | `gutterTintStyle` → **only** `GutterBackgroundHex` (neutral). Matches terrasort **intra = gutter** intent. |
| **Full line after segmentation** | `splitHighlightPair` returns `splitApplyDiffLine(...)` after `highlightSegmented` when word diff applies. |
| **Unified paired lines** | After `unifiedHighlightPair`, `ApplyDiffLineStyle` wraps `-`+body and `+`+body when `LineDiffStyle && !NoColor`. |
| **Gutter **cells** on +/- rows** | `gutterStyleForCell` uses `DiffLineBackgroundColour` for delete/insert column — semantic line colour on **line-number** cells, neutrals on context rows. |

**Conclusion:** Layering order for **word diff** is largely correct; the remaining **parity gap** is **`DiffLineBackgroundColour` vs terrasort `chromaDiffLineRGBA`** for the **semantic** plane (line + semantic gutter cells).

---

## 5. Risks and trade-offs

| Risk | Mitigation |
|------|------------|
| Some Chroma themes define **muted** gd/gi backgrounds (gray/yellow). Terrasort still uses them — parity means accepting that **unless** product later overrides CONTEXT. | Document in tests; optional future flag if users want “strong” semantic blend. |
| **Double background** stacking (full-line red/green + gray word span): terminal-dependent; GitHub PR uses similar hierarchy. | Manual verify in true-color terminal (`11-VERIFICATION.md`). |
| **ANSI** from Chroma between tokens: `ApplyDiffLineStyle` already downgrades `\x1b[0m` → `\x1b[39m`. | Keep when touching render. |

---

## 6. Recommendations for execution (aligns with `11-03-PLAN.md`)

1. **Replace** drift `DiffLineBackgroundColour` body with logic **equivalent** to terrasort `chromaDiffLineRGBA` + shared helpers (`diffEntryChromaColour`, `blendChromaTowardTerminalBase`, same fallbacks). Update godoc to say “terrasort chromaDiffLineRGBA parity”.
2. **Keep** `gutterTintStyle` as **GutterBackgroundHex** only.
3. **Extend** `internal/highlight` tests: for `styles.Get("monokai")` + `isDark`, insert/delete colours show **green/red bias** as today **after** port (may differ from pre-port drift blend).
4. **Optional:** property test comparing drift `DiffLineBackgroundColour` string to terrasort `chromaDiffLineRGBA` RGB for a fixed list of theme names (same `isDark`).

---

## Validation Architecture (Nyquist)

| Dimension | Phase 11 handling |
|-----------|-------------------|
| **1–3** | `go test` on `internal/highlight`, `internal/render`, `internal/worddiff` |
| **4** | Human: side-by-side terrasort vs drift in true-color TTY |
| **5–6** | `just lint`, `golangci-lint` |
| **7** | N/A (library) |
| **8** | `11-VALIDATION.md` maps plan tasks to commands |

---

## RESEARCH COMPLETE

No blockers. Proceed with **`11-03-PLAN.md`** execution; research supports **T1** (port terrasort gd/gi pipeline into drift `diffcolors.go`).
