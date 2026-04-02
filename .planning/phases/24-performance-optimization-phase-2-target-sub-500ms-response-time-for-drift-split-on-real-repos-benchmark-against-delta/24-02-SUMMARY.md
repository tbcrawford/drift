# Phase 24-02 Summary: Eliminate Redundant Repo Opens + Streaming Pager Output

## What Was Built

Two changes to eliminate remaining latency after Plan 24-01:

1. **Eliminated redundant repo opens** — `gitDirectoryVsHEAD` now calls
   `changedFilesViaIndex(absDir)` directly (with `DetectDotGit: true`), removing
   the separate `gitRevParseIsInsideWorkTree` + `openRepoAt` calls that previously
   opened the repo 2-3 times per invocation.

2. **Streaming pager output** — Added `streamThroughPager` which starts the pager
   process *before* rendering begins, so rendered output streams directly to the
   pager rather than buffering the full diff first. All three directory-diff call
   sites (zero-arg, single-dir git, two-dir) were updated to use it.

## Files Modified

- `cmd/drift/main.go`
  - `runDirectoryDiff` signature: `*bytes.Buffer` → `io.Writer`
  - `runGitDirectoryDiff` signature: `*bytes.Buffer` → `io.Writer`
  - Added `streamThroughPager(opts, renderFn)` function
  - Updated zero-arg, single-dir git, and two-dir call sites to use `streamThroughPager`
  - `runRoot` zero-arg path: removed `openRepoAt(cwd)` + `headCommitTree` check;
    calls `gitDirectoryVsHEAD(cwd)` directly and handles `git.ErrRepositoryNotExists` inline

- `cmd/drift/gitworktree.go`
  - `gitDirectoryVsHEAD`: removed separate `gitRevParseIsInsideWorkTree` + `openRepoAt` calls;
    calls `changedFilesViaIndex(absDir)` directly
  - Fixed `ErrRepositoryNotExists` wrap to use `%w` (preserves `errors.Is` chain)

## Key Function Added

### `streamThroughPager(opts *rootOptions, renderFn func(w io.Writer) (bool, error)) error`

Starts the pager immediately (before rendering), then calls `renderFn` which writes
directly to the pager writer. Falls back to `opts.streams.Out` if pager can't start
or if stdout is not a TTY. Eliminates the full-buffer-then-page round-trip.

## Performance Results

| Scenario | Time |
|----------|------|
| Original baseline (wt.Status()) | 4,512ms |
| After Plan 24-01 (changedFilesViaIndex) | ~134ms |
| After Plan 24-02 (streaming + no redundant opens) | ~125ms |
| Reference (git diff \| delta) | ~167ms |
| **Target** | **< 500ms** ✅ |

`drift --split` is now ~25% faster than `git diff | delta` on auth0-tenant-config
(158 files, 25 changed).

## Tests

All 305 tests pass. Fixed `TestRunCLI_zeroArg_notInRepo` which checked for
`"not a git repository"` in stderr — the `errors.Is` chain was restored by
switching the wrap in `gitDirectoryVsHEAD` from `%q` to `%w`.
