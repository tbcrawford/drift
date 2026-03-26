# Phase 8 — Research: Palette best-match Chroma theme

**Phase:** 8 — Terminal palette best-match Chroma theme  
**Question:** What must drift implement to match Terrasort’s automatic theme selection quality without breaking non-TTY or `NO_COLOR` users?

---

## 1. Current drift behavior

- `highlight.SelectTheme(requested, isDark)` with empty `requested` picks **`monokai`** (dark) or **`github`** (light) from `lipgloss.HasDarkBackground` (via `theme.DetectDarkBackground` + color profile guards).
- No knowledge of the terminal emulator’s **indexed 16-color palette** — so themes can look “off” on Solarized, One Dark, etc.

---

## 2. Terrasort reference behavior (summary)

| Piece | Role |
|-------|------|
| **`ParseOSC4Responses`** | Parse raw bytes from OSC 4 replies `\033]4;n;rgb:r/g/b\007` into `[]color.RGBA` indexed by slot. |
| **`BestMatchTheme(palette)`** | For each Chroma theme name (sorted), sum over sample token types: min Euclidean distance from each token’s RGB to **nearest** palette color. Lowest total wins. |
| **`queryOSC4Palette`** (CLI) | Open `/dev/tty`, **raw mode** (`golang.org/x/term`), write 16× `\033]4;n;?\007`, read until 16 BELs or **500ms timeout**, parse with `ParseOSC4Responses`. |
| **Fallback** | If OSC 4 fails → `BackgroundColor` / `HasDarkBackground` → `SelectDefaultTheme` (`github-dark` / `github` in Terrasort). |

Drift should keep its existing **fallback chain** but replace only the “happy path” when palette is available.

---

## 3. Design constraints for drift

1. **Library vs CLI:** Pure functions (`BestMatchTheme`, `ParseOSC4Responses`) belong in **`internal/highlight`** (testable, no I/O). OSC 4 **query** touches `/dev/tty` and raw mode → **`cmd/drift`** or **`internal/terminal`** — not required for `go test` of scoring.
2. **`Render` wiring:** `resolveProfile` already distinguishes TTY `*os.File` vs piped. Best-match should run only when profile indicates color capability, `cfg.theme == ""`, and writer is suitable (and optionally only on Unix for v1).
3. **Performance:** Scoring loops **all** Chroma themes — acceptable at startup once per `Render` call; document if caching theme name on `RenderConfig` is needed to avoid double work in split/unified.
4. **Windows:** Terrasort’s query uses `/dev/tty` — drift should **document** fallback on Windows or use ConPTY APIs only if in scope (default: **skip OSC 4**, use boolean dark/light).

---

## 4. Risks

- **Escape leakage:** Raw TTY read/write must restore terminal state (defer `term.Restore`).
- **CI / non-interactive:** Must never hang — timeout mandatory.
- **Nested drift usage:** Library callers passing `bytes.Buffer` never trigger OSC 4 (already `NoTTY` or non-file).

---

## 5. Verification approach

- **Unit:** `BestMatchTheme` deterministic on fixed palette; `ParseOSC4Responses` golden strings from Terrasort tests pattern.
- **Integration:** Optional `-run OSC4` test behind `GOOS=linux` + `t.Skip` without `/dev/tty`, or table-driven mock of raw response bytes.
- **CLI:** `--show-theme` prints resolved theme name to stderr (optional flag) for manual confirmation.

---

## Validation Architecture

| Dimension | Approach |
|-----------|------------|
| Automated unit | `go test ./internal/highlight/...` — Euclidean + parser + deterministic winner |
| Automated CLI | `go test ./cmd/drift/...` — flag parsing; stderr theme line when `--show-theme` |
| Manual | Real terminal: run `drift` without `--theme`, compare theme name with `--show-theme` vs expectation |

**Quick command:** `go test ./...`

---

## RESEARCH COMPLETE

Findings support a two-wave plan: (1) pure highlight helpers + tests, (2) OSC 4 query + `Render`/CLI wiring + docs.
