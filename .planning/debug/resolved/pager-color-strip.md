---
status: resolved
trigger: "Colors have stopped showing when running `drift lending/modules/pl-auth0-applications/main.tf --split`. Likely broke in one of the last ~20 commits. Suspect PAGER env var. Goal: identify commit, root cause, fix color preservation through pagers."
created: 2026-04-02T00:00:00Z
updated: 2026-04-02T00:30:00Z
symptoms_prefilled: true
---

## Current Focus

hypothesis: CONFIRMED — two bugs:
  1. PRIMARY: render.go `resolveProfile(bytes.Buffer)` returns NoTTY → no ANSI colors generated into buffer at all
  2. SECONDARY: resolvePager() passes $PAGER verbatim; if PAGER=less (no -R), less strips any colors
test: traced execution path: runRoot → writeThroughPager → drift.RenderWithNames(&buf) → buildRenderPipeline(&buf) → resolveProfile(&buf) → NoTTY → no colors
expecting: fix by (1) adding WithColorProfile option to pass real TTY profile, (2) augmenting pager cmd with -R for less
next_action: implement fix

## Symptoms

expected: Colors rendered correctly in terminal output when using `drift --split`
actual: Colors have stopped showing when running `drift lending/modules/pl-auth0-applications/main.tf --split`
errors: no explicit error; output is plain/uncolored
reproduction: run `drift lending/modules/pl-auth0-applications/main.tf --split` with PAGER env var set
started: within last ~20 commits
user_suspect: PAGER env var causing issue; possibly related to recent `WithTermWidth` fix in flags.go

## Eliminated

(none yet)

## Evidence

- timestamp: 2026-04-02T00:05:00Z
  checked: render.go resolveProfile()
  found: resolveProfile(w, cfg) returns NoTTY for any non-*os.File writer (including bytes.Buffer)
  implication: ALL rendering through pager path produces no ANSI colors — colors are never written to buffer

- timestamp: 2026-04-02T00:05:00Z
  checked: main.go runRoot() execution flow
  found: drift.RenderWithNames(result, &buf, ...) passes bytes.Buffer as w → resolveProfile returns NoTTY → resolveChromaStyle skips colors → formatter is noop → output is plain text
  implication: The buffer never contains colors, so the pager isn't the one stripping them

- timestamp: 2026-04-02T00:05:00Z
  checked: flags.go resolveRootOptions(), the recent fc383d3 commit
  found: WithTermWidth was added but there's no analogous WithColorProfile injection — width is captured from real TTY but color profile is NOT
  implication: fc383d3 fixed width but exposed the parallel gap: color profile also needs to be captured from the real TTY before buffering

- timestamp: 2026-04-02T00:05:00Z
  checked: pager.go resolvePager()
  found: If PAGER=less (without -R), colors would also be stripped by less even if they were in the buffer
  implication: Secondary bug — need to ensure -R/--RAW-CONTROL-CHARS is present for less

- timestamp: 2026-04-02T00:05:00Z
  checked: options.go renderConfig struct
  found: No colorProfile field exists — only noColor bool. Must add WithColorProfile option.
  implication: Fix requires new option in options.go + corresponding use in render.go resolveProfile()

## Resolution

root_cause: Two bugs: (1) PRIMARY — buildRenderPipeline called resolveProfile(bytes.Buffer) which always returned colorprofile.NoTTY, so no ANSI codes were ever written into the render buffer. The pager wasn't stripping colors; they were never generated. (2) SECONDARY — resolvePager() passed $PAGER verbatim; if PAGER=less without -R, less would strip any ANSI it did receive.
fix: Added WithColorProfile(p colorprofile.Profile) option. When hasProfile is true, buildRenderPipeline constructs colorprofile.Writer with the explicit profile rather than re-detecting from bytes.Buffer. In flags.go resolveRootOptions, detect color profile from streams.Out before buffering (mirrors WithTermWidth fix). In pager.go, ensureLessColors() appends -R when less is invoked without --RAW-CONTROL-CHARS.
verification: go build ./... success, go test ./... 277 passed (270 original + 7 new), go vet clean. Commit 9f72ef1.
files_changed: [options.go, render.go, render_test.go, cmd/drift/flags.go, cmd/drift/pager.go, cmd/drift/pager_test.go]
