---
status: investigating
trigger: "drift <directory> --split is not showing split (side-by-side) diffs — it appears to render unified diffs instead. For example: `drift lending/modules/pl-auth0-applications --split` shows unified output rather than the expected side-by-side split view. This happens with the new single-directory git diff code path added to `cmd/drift/main.go` (`runGitDirectoryDiff`) and `cmd/drift/gitworktree.go` (`gitDirectoryVsHEAD`)."
created: 2026-04-01T00:00:00Z
updated: 2026-04-01T00:05:00Z
---

## Current Focus

hypothesis: Terminal width is detected from bytes.Buffer (not os.Stdout), producing 80-column split panels even on wide terminals — but this would look narrow, not "unified". The actual rendering path IS split when --split is passed.
test: Read full code path from flag parsing → driftOpts → RenderWithNames → render.Split; live-test with modified files
expecting: Code is correct OR there is a subtle NoTTY/width issue producing output that looks like unified
next_action: Need user to provide exact output they see vs what they expect — code and live tests show split IS working

## Symptoms

expected: `drift <directory> --split` renders a side-by-side split diff for each changed file in the directory
actual: Output shows unified diff format instead of split view
errors: (none reported — visual wrong output)
reproduction: Run `drift lending/modules/pl-auth0-applications --split` (directory that has git changes); output is unified
started: After new single-directory git diff code path was added (runGitDirectoryDiff / gitDirectoryVsHEAD)

## Eliminated

- hypothesis: WithSplit() option not included in driftOpts
  evidence: flags.go resolveRootOptions correctly appends drift.WithSplit() when flags.split == true (line 70-72)
  timestamp: 2026-04-01T00:05:00Z

- hypothesis: driftOpts not passed to RenderWithNames in runGitDirectoryDiff
  evidence: main.go line 125 clearly passes opts.driftOpts... to drift.RenderWithNames
  timestamp: 2026-04-01T00:05:00Z

- hypothesis: cfg.render.split not checked in RenderWithNames
  evidence: render.go lines 91-94 correctly check cfg.render.split and call render.Split when true
  timestamp: 2026-04-01T00:05:00Z

- hypothesis: --split flag position after directory arg breaks cobra parsing
  evidence: live test: both `drift dir --split` and `drift --split dir` produce identical split output
  timestamp: 2026-04-01T00:05:00Z

- hypothesis: render.Split fallback to unified when profile is NoTTY
  evidence: render.Split has no NoTTY fallback — it always renders side-by-side. Only word-diff is disabled for NoTTY (shouldWordDiffPair returns false), but that doesn't affect split vs unified choice.
  timestamp: 2026-04-01T00:05:00Z

## Evidence

- timestamp: 2026-04-01T00:03:00Z
  checked: cmd/drift/flags.go resolveRootOptions
  found: flags.split == true → drift.WithSplit() appended to opts (line 70-72); included in driftOpts
  implication: WithSplit option correctly flows into rootOptions.driftOpts

- timestamp: 2026-04-01T00:03:00Z
  checked: cmd/drift/main.go runGitDirectoryDiff
  found: line 125: drift.RenderWithNames(result, buf, pair.OldName, pair.NewName, opts.driftOpts...) — driftOpts passed
  implication: WithSplit() is passed to RenderWithNames for every file in git directory diff

- timestamp: 2026-04-01T00:03:00Z
  checked: drift/render.go RenderWithNames
  found: defaultConfig() creates fresh config; opts applied; if cfg.render.split → render.Split else render.Unified
  implication: Split path is taken whenever WithSplit() is in opts

- timestamp: 2026-04-01T00:04:00Z
  checked: Live test — ./drift internal/render --split --no-color (after modifying unified.go)
  found: Output has "@@ ... @@" hunk header, then two columns separated by "│" — clearly side-by-side
  implication: Split rendering IS working in the git directory code path

- timestamp: 2026-04-01T00:04:00Z
  checked: Live test — ./drift internal/render --no-color (unified, same files)
  found: Output has "--- filename (HEAD)" / "+++ filename (working tree)" headers + "+"/"-" prefixes — clearly unified
  implication: Split and unified produce visually distinct output; user could not confuse them

- timestamp: 2026-04-01T00:05:00Z
  checked: render/termwidth.go TerminalWidth — called with bytes.Buffer (not os.Stdout)
  found: bytes.Buffer is not *os.File → falls to COLUMNS env var → defaults to 80 columns
  implication: Split panels are only 39 chars wide even on wide terminals; output looks "narrow" but still distinctly side-by-side

## Resolution

root_cause: Cannot reproduce — code is correct. Split rendering works as expected in runGitDirectoryDiff. The most likely explanation is either: (a) the user had a stale binary, (b) the rendering is correct but panels are narrow (80-col default) on a wide terminal making it look like "broken split", or (c) the user may have confused the absence of ---/+++ file headers in split mode with unified output. One real concern: TerminalWidth is detected from bytes.Buffer not os.Stdout, so split panels are always 80-col default regardless of actual terminal width.
fix: N/A — cannot reproduce; OR fix TerminalWidth detection to use opts.streams.Out instead of the buffer
verification:
files_changed: []
