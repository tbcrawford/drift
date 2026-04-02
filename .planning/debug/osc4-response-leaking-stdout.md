---
status: awaiting_human_verify
trigger: "When running `drift --split` at repo root, OSC 4 palette query responses appear as raw escape sequences in diff output for one file."
created: 2026-04-02T00:00:00Z
updated: 2026-04-02T00:00:00Z
---

## Current Focus

hypothesis: Multiple parallel goroutines in runGitDirectoryDiff each call RenderWithNames → buildRenderPipeline → resolveChromaStyle → QueryANSIPalette concurrently. QueryANSIPalette opens /dev/tty, puts it in raw mode, writes 16 OSC 4 queries, and reads responses. When N goroutines do this simultaneously, one goroutine's read loop grabs responses that were intended for another goroutine's queries — or the raw mode state interferes — causing OSC 4 response bytes to leak into the wrong read buffer. Since the render is writing to a bytes.Buffer (not stdout), the only way those escape sequences reach visible output is if they come through tty.Read() into one goroutine's parse buffer and then get embedded in the rendered bytes.Buffer output. Actually the more likely path: /dev/tty has a single OS-level buffer; goroutine A sends 16 queries, goroutine B sends 16 queries, both are reading concurrently; goroutine A's read call returns bytes that belong to goroutine B's responses (they're interleaved in the kernel tty buffer); ParseOSC4Responses fails for goroutine A (gets garbled data); goroutine B's read call returns goroutine A's response bytes. The unread bytes from the terminal then flow through the terminal's cooked-mode echo or get forwarded to stdout after raw mode is restored.
test: Confirmed by code inspection - resolveChromaStyle line 181 calls terminal.QueryANSIPalette() unconditionally whenever profile is color-capable, and there is no caching at CLI level. flags.go caches WithIsDark and WithColorProfile but NOT the OSC 4 result / resolved theme name.
expecting: Fix: resolve the theme name once in resolveRootOptions (before goroutines start), then inject it via WithTheme so resolveChromaStyle short-circuits at the cfg.render.theme != "" branch and never calls QueryANSIPalette in the goroutines.
next_action: awaiting human verification that `drift --split` at repo root no longer shows raw OSC 4 sequences in output

## Symptoms

expected: Syntax highlighting works silently — no raw escape sequences visible in diff output
actual: For one file in the directory diff, raw OSC 4 responses appear inline in the diff output, e.g.: ^[]4;9;rgb:bfbf/6161/6a6a^G^[]4;10;rgb:a3a3/bebe/8c8c^G^[]4;11;rgb:ebeb/cbcb/8b8b^G...
errors: No crash, but visible garbage in output
reproduction: Run `drift --split` at the repo root (runs a git directory diff with many files)
started: After a fix was just applied — removed the `w.(*os.File)` guard from `resolveChromaStyle` in `render.go`, so `terminal.QueryANSIPalette()` now fires for any color-capable profile

## Eliminated

- hypothesis: The leaking is caused by raw mode not being restored between goroutines
  evidence: QueryANSIPalette defers term.Restore correctly; but this doesn't prevent concurrent access to the same /dev/tty fd from multiple goroutines
  timestamp: 2026-04-02T00:00:00Z

- hypothesis: The response bytes are being written to stdout directly
  evidence: QueryANSIPalette writes only to /dev/tty and reads only from /dev/tty; but the kernel tty read buffer is shared — interleaved responses from N concurrent goroutines corrupt each other's parse attempts, and unread bytes can end up being echoed or flushed when raw mode is restored
  timestamp: 2026-04-02T00:00:00Z

## Evidence

- timestamp: 2026-04-02T00:00:00Z
  checked: render.go:167-194 resolveChromaStyle
  found: The OSC 4 branch at line 181 has NO guard for concurrent calls. QueryANSIPalette() is called every time profile is color-capable AND no explicit theme is set. In directory diff, N goroutines each call buildRenderPipeline → resolveChromaStyle concurrently.
  implication: N concurrent QueryANSIPalette calls → N goroutines all open /dev/tty, all MakeRaw, all write 16 OSC 4 queries, all read from the same terminal buffer. Interleaving is inevitable.

- timestamp: 2026-04-02T00:00:00Z
  checked: flags.go:83-111 resolveRootOptions
  found: The fix for OSC 11 (WithIsDark) is in place — lipgloss.HasDarkBackground fires once before goroutines. But there is NO equivalent for the OSC 4 palette query. WithTheme is only set when flags.theme != "".
  implication: The pattern is clear — apply the same pre-goroutine resolution for OSC 4 that was already done for OSC 11.

- timestamp: 2026-04-02T00:00:00Z
  checked: options.go WithTheme
  found: WithTheme sets cfg.render.theme; resolveChromaStyle checks cfg.render.theme != "" FIRST (line 171) — before the OSC 4 branch. So injecting the resolved theme name via WithTheme perfectly short-circuits the concurrent QueryANSIPalette calls.
  implication: The fix requires no new Option type. Just call QueryANSIPalette + BestMatchTheme once in resolveRootOptions, then append WithTheme(name) to opts, mirroring the WithIsDark pattern.

- timestamp: 2026-04-02T00:00:00Z
  checked: internal/terminal/palette_unix.go QueryANSIPalette
  found: Opens /dev/tty, MakeRaw, writes 16 OSC 4 queries, reads until 16 BEL chars or timeout. No mutex, no once.Do — purely per-call.
  implication: Concurrent calls will race on /dev/tty — kernel tty buffer is shared; MakeRaw/Restore calls from multiple goroutines will corrupt each other's terminal state.

## Resolution

root_cause: After removing the *os.File guard from resolveChromaStyle, terminal.QueryANSIPalette() is called once per file in parallel goroutines during directory diff. Multiple concurrent goroutines open /dev/tty, set raw mode, write OSC 4 queries, and read from the shared tty buffer. The responses are interleaved in the kernel buffer, so goroutines receive each other's response bytes. Unread OSC 4 response bytes can be flushed/echoed to stdout when raw mode is restored, producing visible garbage in the diff output. The OSC 11 (isDark) problem was previously fixed by pre-resolving with WithIsDark in flags.go; the OSC 4 (theme) problem needs the identical treatment.
fix: In resolveRootOptions (flags.go), after the color profile and isDark detection block, call terminal.QueryANSIPalette() and highlight.BestMatchTheme() once (when no explicit --theme was given and the terminal is color-capable), then inject the result via drift.WithTheme(resolvedThemeName). This causes resolveChromaStyle to take the cfg.render.theme != "" branch in every goroutine, skipping QueryANSIPalette entirely.
verification: Build succeeded. 292/292 tests pass. Fix applies the identical pre-goroutine resolution pattern already used for OSC 11 (WithIsDark), now extended to OSC 4 (WithTheme). Awaiting human confirmation that raw sequences no longer appear in `drift --split` output.
files_changed: [cmd/drift/flags.go]
