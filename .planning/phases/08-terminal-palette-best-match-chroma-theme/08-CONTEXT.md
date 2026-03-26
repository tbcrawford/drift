# Phase 8: Terminal palette best-match Chroma theme — Context

**Gathered:** 2026-03-25  
**Status:** Ready for planning  
**Source:** `/gsd-plan-phase` — user request to support Terrasort-style automatic theme matching

---

## Phase boundary

Deliver **palette-aware Chroma theme selection** for drift when the caller does not set an explicit theme: query the terminal’s 16-color ANSI palette via **OSC 4**, score every registered Chroma style using the same **Euclidean nearest-neighbor** idea as Terrasort’s `BestMatchTheme`, and use the winner for `SelectTheme` / rendering. If OSC 4 is unavailable, times out, or stdout is not a real terminal file, **fall back** to the existing **dark vs light** `SelectTheme("", isDark)` behavior (no regression for piped output or `NO_COLOR`).

OSC 4 batch query and raw TTY handling are **I/O**; keep pure scoring + parsing in `internal/highlight`, and place the TTY query in **`cmd/drift`** or a small `internal/terminal` package used only from CLI / `Render` when `*os.File`.

---

## Implementation decisions (locked)

1. **Algorithm parity:** Implement `BestMatchTheme(palette []color.RGBA) string` with the same scoring rules as Terrasort: sample token types (`Keyword`, `LiteralString`, `Comment`, `NameBuiltin`, `LiteralNumber`), sum minimum Euclidean distance to nearest palette slot per token, lowest score wins; tie-break alphabetical by theme name.
2. **No new algorithm in v1 for diff row chrome:** This phase adjusts **syntax** theme selection only — not Lip Gloss row backgrounds (Terrasort’s `UXTheme` is out of scope unless explicitly added later).
3. **Explicit theme wins:** If `WithTheme("x")` / `--theme` is set, skip OSC 4 and best-match entirely.
4. **Safety:** OSC 4 query uses a **short timeout** (e.g. 500ms) and must not block indefinitely; on failure, fall back silently to current behavior unless `--show-theme` requests visibility.
5. **Platform:** Document that OSC 4 + `/dev/tty` is **Unix-first**; Windows behavior is “fallback only” unless a tested implementation is added in the same phase.
6. **Dependencies:** Prefer existing modules (`golang.org/x/sys`, `charmbracelet/x/term` already in tree) for raw TTY; add `golang.org/x/term` only if required and justified in plan tasks.

---

## Canonical references

- Terrasort (reference implementation, same author):  
  - `/Users/tylercrawford/dev/github/terrasort/internal/highlight/themematch.go` — `BestMatchTheme`  
  - `/Users/tylercrawford/dev/github/terrasort/internal/highlight/osc4.go` — `ParseOSC4Responses`  
  - `/Users/tylercrawford/dev/github/terrasort/internal/cli/osc4query.go` — batch OSC 4 query + timeout  
- Drift: `internal/highlight/highlight.go` — `SelectTheme`, `DetectLexer`  
- Drift: `internal/theme/theme.go` — `DetectDarkBackground`  
- Drift: `render.go` — `resolveProfile`, `Render` / `RenderWithNames`

---

## Deferred

- Per-line diff background tints derived from terminal RGB (Terrasort `DeriveUXThemeFromColor`) — not in this phase.  
- Live theme switching without re-run — explicitly out of scope (see Terrasort `uxtheme.go` comments).
