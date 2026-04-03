---
status: resolved
trigger: "git diff shows nothing when drift is configured as a git pager, but drift --split works standalone"
created: 2026-04-03T00:00:00Z
updated: 2026-04-03T03:30:00Z
---

## Current Focus

hypothesis: CONFIRMED — git sends ANSI-colored output to its pager when stdout is a TTY.
Every line is prefixed with escape sequences (e.g. `\x1b[1m`). `parseUnifiedDiff` checks
`strings.HasPrefix(line, "diff --git ")` which fails on `\x1b[1mdiff --git ...`.
Result: 0 files parsed → no output → exit 0.

test: PTY test with /tmp/drift-fixed binary confirms output appears (less opens with rendered diff)
expecting: Human runs `git diff` with installed drift and confirms it renders correctly

next_action: Human verification — run `git diff` in auth0-tenant-config to confirm output appears

## Symptoms

expected: Running `git diff` in /Users/tylercrawford/dev/figure/auth0-tenant-config should display diff output rendered through drift as the git pager
actual: No output is shown at all — terminal returns to prompt immediately with nothing rendered
errors: None visible — no error messages, just silence
reproduction: In /Users/tylercrawford/dev/figure/auth0-tenant-config, run `git diff`. Pager is configured (core.pager = drift).
started: Potentially after a recent change that collapsed two-module workspace into single go.mod and fixed lint errors in pager_test.go and unifieddiff.go

## Eliminated

- hypothesis: lint fix commit (f45d61c) caused regression
  evidence: confirmed not the regression; pipeline correctly parses and renders real git diff output
  timestamp: 2026-04-03T01:00:00Z

- hypothesis: lipgloss.HasDarkBackground hangs by consuming stdin pipe data
  evidence: term.MakeRaw on a pipe fd fails immediately ("inappropriate ioctl for device"); no hang, no data consumed; confirmed via dedicated test binary run in PTY as git pager
  timestamp: 2026-04-03T01:00:00Z

- hypothesis: QueryANSIPalette consumes stdin
  evidence: opens /dev/tty directly, never touches stdin pipe
  timestamp: 2026-04-03T01:00:00Z

- hypothesis: parseUnifiedDiff fails on real git diff output
  evidence: fails because git sends ANSI-colored output — this was the actual root cause, not a parser bug per se
  timestamp: 2026-04-03T03:00:00Z

- hypothesis: git bypasses pager when stdout is piped (all pipe-based tests were invalid)
  evidence: git only invokes the pager when its own stdout is a TTY; confirmed via `script` PTY test
  timestamp: 2026-04-03T01:00:00Z

- hypothesis: Nested `less` from streamThroughPager caused silent exit
  evidence: GIT_PAGER_IN_USE check added; fix was necessary but not sufficient — runPagerMode itself was returning 0 files
  timestamp: 2026-04-03T02:00:00Z

- hypothesis: MakeRaw(stdout) corrupts or empties stdin pipe
  evidence: Tested directly in PTY as git pager: stdin had 22 lines after MakeRaw(stdout)+Restore cycle; confirmed clean
  timestamp: 2026-04-03T03:00:00Z

## Evidence

- timestamp: 2026-04-03T01:00:00Z
  checked: git pager environment (via `script` PTY wrapper)
  found: LESS=FRX, PAGER=less, GIT_PAGER_IN_USE= (set but empty)
  implication: git sets LESS=FRX every time it invokes the pager; -F means "quit if one screen"

- timestamp: 2026-04-03T01:00:00Z
  checked: streamThroughPager in main.go
  found: was unconditionally starting nested less; fixed with GIT_PAGER_IN_USE check
  implication: GIT_PAGER_IN_USE fix was necessary but not sufficient

- timestamp: 2026-04-03T03:00:00Z
  checked: debug binary with fmt.Fprintln in runPagerMode
  found: runPagerMode reached; parseUnifiedDiff returned files=0, err=nil
  implication: the parser sees the input but doesn't recognize any "diff --git" lines

- timestamp: 2026-04-03T03:00:00Z
  checked: bytes sent by git to pager in PTY (via `git -c "core.pager=cat"` in script)
  found: first bytes are `\x1b[1mdiff --git ...` — ANSI bold escape prefix on every line
  implication: git enables color output when it detects a TTY; strings.HasPrefix(line, "diff --git ") NEVER matches; 0 files parsed

- timestamp: 2026-04-03T03:00:00Z
  checked: available dependencies for ANSI stripping
  found: `github.com/charmbracelet/x/ansi` is already a dep; provides `ansi.Strip(s string) string`
  implication: no new dependencies needed for the fix

- timestamp: 2026-04-03T03:00:00Z
  checked: PTY test with fixed binary (/tmp/drift-fixed)
  found: less opened with fully rendered drift output — file header, syntax highlighting, line numbers, +/- rows
  implication: fix is correct and working

- timestamp: 2026-04-03T03:00:00Z
  checked: full test suite with `rtk go test ./...`
  found: 324 tests pass, 0 failures
  implication: no regressions introduced

## Resolution

root_cause: Git sends ANSI-colored output to its pager when stdout is a TTY (which it always is when a pager is invoked). Every line is prefixed with ANSI escape sequences (e.g. `\x1b[1m` for bold headers, `\x1b[31m` for red deletions). `parseUnifiedDiff` used `strings.HasPrefix(line, "diff --git ")` and similar checks, which always failed on the escape-prefixed lines. As a result, no files were ever parsed and drift exited with 0 and no output. The GIT_PAGER_IN_USE fix (preventing nested less) was also necessary — without it, the chain was git→drift→less and less exited silently — but even after that fix, the parser returned 0 files due to the ANSI color issue.

fix: In `parseUnifiedDiff` (unifieddiff.go), import `github.com/charmbracelet/x/ansi` and call `ansi.Strip(scanner.Text())` before the switch/case prefix checks. This removes all ANSI escape sequences from each line before parsing, so `strings.HasPrefix(line, "diff --git ")` works correctly regardless of whether git colored its output.

verification: PTY test confirmed: `script -q ... /tmp/drift-fixed diff HEAD~1 -- main.tf` shows less opening with fully rendered drift output. All 324 tests pass. Binary installed via `just install`.
files_changed: [cmd/drift/unifieddiff.go, cmd/drift/main.go]
