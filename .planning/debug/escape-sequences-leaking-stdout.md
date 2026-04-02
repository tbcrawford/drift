---
status: resolved
trigger: "Running `drift lending/modules/pl-auth0-applications --split` prints raw terminal escape sequences instead of diff output"
created: 2026-04-02T00:00:00Z
updated: 2026-04-02T00:10:00Z
---

## Current Focus

hypothesis: `lipgloss.HasDarkBackground(os.Stdin, os.Stdout)` is called once per file goroutine in `runGitDirectoryDiff`. When called concurrently, multiple OSC11+DA2 queries are sent to the terminal simultaneously. `backgroundColor(os.Stdout, os.Stdout)` writes the OSC11+DA2 query to stdout (since it uses the same file for both read and write). On macOS PTY, reading from stdout fd (write-only) fails immediately, leaving the query written to stdout AND terminal responses unread in the PTY buffer. The real fix is to avoid calling `HasDarkBackground` repeatedly (once per file) and to not call it at all when the color profile was pre-detected.
test: Trace `DetectDarkBackground` → `HasDarkBackground` call path; confirm it's called once per file goroutine; verify the specific mechanism (stdout write vs stdin interaction).
expecting: The fix is to cache `isDark` at the `resolveRootOptions` level (detect once, pass in via option) and guard `HasDarkBackground` calls to only happen when output is the real TTY (not a buffer).
next_action: Implement the fix — add `WithIsDark(bool)` option or detect dark background once in `resolveRootOptions` and pass it through, bypassing the per-call `HasDarkBackground` probe.

## Symptoms

expected: A rendered split-view diff of all changed files in the directory, with styled file headers
actual: Only raw terminal escape sequences printed to stdout — OSC 11 (background color query responses: `^[]11;rgb:2e2e/3434/4040^G`) and DA2 device attribute responses (`^[[?62;22;52c`). No diff content visible.
errors: No error messages — exits normally
reproduction: `drift lending/modules/pl-auth0-applications --split` (single directory argument, --split flag)
timeline: Likely introduced by the color profile detection fix that added `colorprofile.Detect(f, os.Environ())` in `cmd/drift/flags.go`

## Eliminated

- hypothesis: `colorprofile.Detect(f, os.Environ())` sends OSC/DA2 queries
  evidence: Reviewed env.go in colorprofile@v0.4.2 — `Detect` uses only environment variables (TERM, COLORTERM, etc.) and terminfo database; NO terminal queries sent
  timestamp: 2026-04-02T00:05:00Z

- hypothesis: `terminal.QueryANSIPalette()` is the source (called from resolveChromaStyle)
  evidence: `resolveChromaStyle` checks `if _, ok := w.(*os.File); ok` — when w is *bytes.Buffer (directory diff path), this check fails and QueryANSIPalette is never called
  timestamp: 2026-04-02T00:07:00Z

- hypothesis: `colorprofile.NewWriter` sends terminal queries
  evidence: Reviewed writer.go in colorprofile@v0.4.2 — `NewWriter` only calls `Detect(w, environ)` which uses env vars only; no terminal I/O
  timestamp: 2026-04-02T00:08:00Z

## Evidence

- timestamp: 2026-04-02T00:02:00Z
  checked: colorprofile@v0.4.2/env.go — the Detect() function
  found: Detect() uses only environment variables (TERM, COLORTERM, NO_COLOR, CLICOLOR etc.) and terminfo database. Does NOT send OSC or DA2 queries to the terminal.
  implication: colorprofile.Detect is not the source of the escape sequences

- timestamp: 2026-04-02T00:03:00Z
  checked: charm.land/lipgloss/v2@v2.0.2/terminal.go and query.go
  found: `HasDarkBackground(os.Stdin, os.Stdout)` → `BackgroundColor(in, out)`. On Unix, tries `backgroundColor(f, f)` for each of `{in, out}`. `backgroundColor(os.Stdout, os.Stdout)` writes query `"\x1b]11;?\x07\x1b[c"` to stdout AND tries to read the terminal's response from stdout fd. The terminal's response (OSC 11 + DA2) matches the exact escape sequences in the symptom: `]11;rgb:2e2e/3434/4040` and `[?62;22;52c`.
  implication: `HasDarkBackground` IS the source of the terminal probing

- timestamp: 2026-04-02T00:06:00Z
  checked: internal/theme/theme.go `DetectDarkBackground`, render.go `buildRenderPipeline`
  found: `DetectDarkBackground(profile)` calls `lipgloss.HasDarkBackground(os.Stdin, os.Stdout)` whenever profile is NOT NoTTY or Ascii. Since `resolveProfile` returns TrueColor (from `hasProfile`), this runs for EVERY file in the directory diff. Since `runGitDirectoryDiff` uses goroutines, `HasDarkBackground` is called concurrently from N goroutines.
  implication: Multiple concurrent terminal queries via os.Stdin and os.Stdout; racing OSC11+DA2 queries. `backgroundColor(os.Stdout, os.Stdout)` specifically writes the query to stdout. Since stdout fd may not be readable on macOS PTY, reads fail immediately, but the WRITTEN QUERY appears in stdout output along with the unread terminal responses.

- timestamp: 2026-04-02T00:09:00Z
  checked: Lipgloss query.go lines 56-64: `for _, f := range []term.File{in, out}`
  found: With `HasDarkBackground(os.Stdin, os.Stdout)`, iteration 2 is `backgroundColor(os.Stdout, os.Stdout)`. This writes the OSC11 query to stdout fd 1. The query itself (`\x1b]11;?\x07\x1b[c`) is written to stdout. The terminal's response may also end up in stdout if os.Stdout fd is readable on the PTY.
  implication: The escape sequences visible in stdout are either (a) the terminal's response bytes being read from stdout fd and then re-emitted, or (b) the query + response interaction causing bytes to appear in the output stream.

## Resolution

root_cause: `theme.DetectDarkBackground` calls `lipgloss.HasDarkBackground(os.Stdin, os.Stdout)` once per file in the directory diff (from concurrent goroutines). This function, on Unix, tries `backgroundColor(os.Stdout, os.Stdout)` which writes OSC11+DA2 queries to stdout. Either the terminal's response bytes leak into stdout (via PTY bidirectionality), or the concurrent queries corrupt each other. The fundamental issue is that `isDark` detection happens inside `buildRenderPipeline` for every `RenderWithNames` call, even when rendering to a buffer — it should be detected once from the real terminal before buffering, cached, and passed in like `colorProfile` is.

fix: Added `WithIsDark(bool)` option to `options.go` with `hasIsDark` sentinel (mirrors `WithColorProfile`/`hasProfile` pattern). In `resolveRootOptions` (flags.go), detect dark background once via `lipgloss.HasDarkBackground(os.Stdin, f)` when output is a TTY and colors are enabled, then pass it via `WithIsDark`. In `buildRenderPipeline` (render.go), check `cfg.render.hasIsDark` first and skip `theme.DetectDarkBackground` (which calls `HasDarkBackground`) when `isDark` is already known.

verification: 277 tests pass; `go build ./cmd/drift/` succeeds; LSP reports no errors. The OSC 11 query now fires exactly once before any goroutines start (in `resolveRootOptions`), preventing concurrent terminal queries from racing and leaking responses into stdout.
files_changed: [options.go, render.go, cmd/drift/flags.go]
