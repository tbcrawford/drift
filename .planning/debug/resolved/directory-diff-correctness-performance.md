---
status: resolved
trigger: "directory diff suspected issues: 1. correctness - equal files leaking through or changed files missing. 2. performance - should use parallel goroutines per file then merge deterministically"
created: 2026-04-02T00:00:00Z
updated: 2026-04-02T00:05:00Z
---

## Current Focus

hypothesis: RESOLVED
test: complete
expecting: n/a
next_action: archive

## Symptoms

expected: Only files that actually differ show in output; diffing is fast (parallel)
actual: Sequential diffing; double file reads in two-dir path; content already in memory not reused
errors: none reported
reproduction: drift dir-a/ dir-b/ OR drift dir/
started: unknown

## Eliminated

- hypothesis: "equal files leak through IsEqual check"
  evidence: Both runDirectoryDiff and runGitDirectoryDiff check result.IsEqual and skip — correctly implemented. dirwalk.go also pre-filters with bytes.Equal. No leak.
  timestamp: 2026-04-02T00:01:00Z

- hypothesis: "gitDirectoryVsHEAD misses new/deleted files"
  evidence: New files handled (gitShowHEADBlob returns "" when cat-file -e fails). Deleted files handled via git ls-tree. Both code paths correct.
  timestamp: 2026-04-02T00:01:00Z

- hypothesis: "CRLF mismatch causes false equals"
  evidence: gitDirectoryVsHEAD uses string equality (headContent == workContent) as optimization, drift.Diff normalizes CRLF separately. A CRLF-vs-LF file would pass the first filter into pairs[] but then result.IsEqual=true in runGitDirectoryDiff would catch it. Not a correctness bug, just sub-optimal.
  timestamp: 2026-04-02T00:01:00Z

## Evidence

- timestamp: 2026-04-02T00:01:00Z
  checked: drift.Diff() concurrency safety
  found: All package-level vars are read-only (compiled regexes, interface assertions). drift.Diff and drift.RenderWithNames are stateless — safe to call concurrently.
  implication: Parallel diffing is safe with no lock needed.

- timestamp: 2026-04-02T00:01:00Z
  checked: runDirectoryDiff + diffDirectories interaction
  found: diffDirectories reads file contents (bytes.Equal check) then DISCARDS content. runDirectoryDiff then reads files AGAIN to pass to drift.Diff. Two reads per file.
  implication: Performance bug: files read twice. Fixed by carrying OldContent/NewContent in filePair.

- timestamp: 2026-04-02T00:01:00Z
  checked: runGitDirectoryDiff + gitDirectoryVsHEAD interaction
  found: gitDirectoryVsHEAD already loads both headContent and workContent into memory (gitFilePair). runGitDirectoryDiff then passes them directly to drift.Diff. No double read here.
  implication: Git path is efficient on reads but sequential on diff computation — fixed with parallel errgroup.

- timestamp: 2026-04-02T00:01:00Z
  checked: Parallel safety of runGit subprocess
  found: exec.Command spawns independent subprocesses. No shared state between goroutines. Safe to parallelize.
  implication: errgroup parallelization is safe for both paths.

- timestamp: 2026-04-02T00:01:00Z
  checked: Output ordering requirements
  found: Both functions return sorted []filePair / []gitFilePair. Pre-sized indexed result slice + post-Wait merge preserves sort order.
  implication: Deterministic output guaranteed despite parallel execution.

## Resolution

root_cause: |
  1. runDirectoryDiff read file contents twice: once in diffDirectories (bytes.Equal) and again when passing to drift.Diff. The bytes read for equality checking were discarded, forcing a redundant second read.
  2. Both runDirectoryDiff and runGitDirectoryDiff processed files sequentially. drift.Diff and drift.RenderWithNames are fully stateless (all package-level vars are read-only), making parallel execution safe with no synchronization needed.

fix: |
  1. filePair struct extended with OldContent/NewContent string fields, populated during diffDirectories walk (reusing the bytes already read for bytes.Equal). runDirectoryDiff uses pair content directly — zero additional ReadFile calls.
  2. runDirectoryDiff and runGitDirectoryDiff now use errgroup.Group to launch one goroutine per file pair. Each goroutine writes to its own pre-allocated bytes.Buffer (indexed by position). After g.Wait(), results are merged in sorted order into the output buffer.
  3. IsEqual guard retained inside each goroutine to handle CRLF-only differences that bytes.Equal would flag as different but drift.Diff normalizes away.
  4. golang.org/x/sync promoted from indirect to direct dependency in go.mod.

verification: 277 tests pass, go build succeeds. All existing directory diff tests (TestRunCLI_directoryDiff_*, TestDiffDirectories) pass with parallel implementation.
files_changed: [cmd/drift/dirwalk.go, cmd/drift/main.go, go.mod]
