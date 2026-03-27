# Summary 11-02 — Gap closure: word diff render integration

**Completed:** 2026-03-26

## Delivered

- Paired delete/insert rows use per-token Chroma highlighting with `Comment`-muted unchanged spans and `gutterTintStyle` backgrounds on changed spans (matches gutter palette).
- Word diff activates only for color profiles (TrueColor / ANSI256 / ANSI), avoiding `FormatterFunc` equality (not comparable) and matching piped/no-color paths.
- Public `WithWordDiff(bool)` (default true) and builder `WordDiff`; `RenderConfig.WordDiff` plumbed from `Render` / `RenderWithNames`.

## Notes

Optional: terminal visual check of a multi-token substitution in unified and split modes.
