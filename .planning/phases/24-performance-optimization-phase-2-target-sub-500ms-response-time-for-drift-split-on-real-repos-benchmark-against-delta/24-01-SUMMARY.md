# Phase 24-01 Summary: Fast Index-Based Change Detection

## What Was Built

Replaced `wt.Status()` (go-git's slow two-pass filesystem walk) with a fast
`changedFilesViaIndex` function that uses stored mtime values from the git index
to detect working-tree changes without reading file content.

## Files Modified

- `cmd/drift/gitworktree.go` — added `changedFilesViaIndex`, `gitShowHEADBlobFromTree`;
  rewired `gitDirectoryVsHEAD` to use both; added performance comment block at top
- `cmd/drift/gitworktree_test.go` — added 6 `TestChangedFilesViaIndex_*` tests,
  `BenchmarkChangedFilesViaIndex`, and `modifyFile` helper

## Key Functions Added

### `changedFilesViaIndex(repoRoot string) (paths []string, headTree *object.Tree, wtRoot string, err error)`

Algorithm:
1. Open repo with `DetectDotGit: true` (one repo open, no redundant calls)
2. Build `map[filename]headHash` by iterating HEAD tree
3. Iterate index entries:
   - Entry not in HEAD → new staged file → include
   - Entry hash ≠ HEAD hash → staged change → include
   - Otherwise `os.Stat` the file:
     - Missing → deleted → include
     - `mtime ≠ entry.ModifiedAt` → potential unstaged change → include

Returns the HEAD `*object.Tree` so callers can read blobs without re-opening.

### `gitShowHEADBlobFromTree(tree *object.Tree, relpathSlash string) (string, error)`

Reads blob content from an already-loaded tree. Used inside `gitDirectoryVsHEAD`
loops to avoid one repo-open per file.

## Performance Results

| Scenario | Time |
|----------|------|
| Before (wt.Status()) | 4,512ms |
| After (changedFilesViaIndex) | ~134ms |
| Reference (git diff \| delta) | ~167ms |

Benchmark: `BenchmarkChangedFilesViaIndex` on 200-file synthetic repo — 14.6ms/op.

## Tests

- `TestChangedFilesViaIndex_NoChanges` — clean repo returns empty slice
- `TestChangedFilesViaIndex_ModifiedFile` — mtime-changed file is detected
- `TestChangedFilesViaIndex_StagedChange` — staged change (index hash ≠ HEAD) detected
- `TestChangedFilesViaIndex_NewStagedFile` — new file not in HEAD is detected
- `TestChangedFilesViaIndex_DeletedFile` — deleted file (stat fails) is detected
- `TestChangedFilesViaIndex_EmptyRepo` — repo with no HEAD commit returns empty, no panic

All 305 tests passing after this plan.
