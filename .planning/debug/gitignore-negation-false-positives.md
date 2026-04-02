---
status: awaiting_human_verify
trigger: "gitignore-negation-false-positives"
created: 2026-04-02T00:00:00Z
updated: 2026-04-02T00:01:00Z
---

## Current Focus

hypothesis: CONFIRMED AND FIXED. Root cause verified. Fix implemented using worktree.Status() instead of filesystem walk + gitignore filter.
test: All 294 tests pass including 2 new regression tests for the negation pattern scenario.
expecting: User confirms the original issue is resolved in their real workflow.
next_action: await human verification

## Symptoms

expected: Only files that actually have changes (modified, added, deleted) appear in the diff. Files listed with negation patterns in `.gitignore` (e.g. `!.idea/auth0-tenant-config.iml`) should only appear if they are actually changed.
actual: Files matched by negation patterns in `.gitignore` are incorrectly included in the diff output even when they have no actual changes.
errors: No crash or error — just spurious file entries appearing in the output.
reproduction: Have a `.gitignore` with a line like `!.idea/auth0-tenant-config.iml`. Run `drift` (zero-arg mode) or `drift <dir> <dir>`. The file shows up in the diff output even if it hasn't changed.
timeline: This has likely never worked correctly — the directory-walking approach has always had this limitation.

## Eliminated

- hypothesis: Content equality check (line 294 in old code) prevents false positives
  evidence: The check only covers working-tree files. But if a negation un-ignores a file that is NOT tracked in HEAD, headContent="" and workContent=actual content → file incorrectly appears as "added". Also for the case where `.idea/` parent dir is ignored but negation un-ignores a child file, go-git's matcher un-ignores the child, the file passes filterGitIgnored, and if HEAD content differs (or file not in HEAD at all) it shows up.
  timestamp: 2026-04-02T00:00:30Z

## Evidence

- timestamp: 2026-04-02T00:00:00Z
  checked: orchestrator hypothesis
  found: Architecture is walk-FS → filter gitignore → diff all; negation patterns un-ignore files, which then get included regardless of whether they changed
  implication: Need to use worktree.Status() to get only truly changed files in git mode

- timestamp: 2026-04-02T00:00:30Z
  checked: gitworktree.go lines 220-266 (old filterGitIgnored approach)
  found: gitDirectoryVsHEAD walked filesystem, then called filterGitIgnored which uses gitignore.Matcher. Negation pattern causes matcher.Match() to return false (un-ignored) → file passes filter → included in workFiles even when content unchanged.
  implication: The fundamental issue is asking "is this file ignored?" instead of "has this file changed?"

- timestamp: 2026-04-02T00:01:00Z
  checked: go-git status.go and worktree_status.go
  found: worktree.Status() returns git.Status map[path]*FileStatus with Staging and Worktree codes. Unmodified files are NOT in the map (with Empty strategy) or have both codes = Unmodified (with Preload). Only actually changed files appear with Modified/Added/Deleted/etc codes.
  implication: worktree.Status() is the correct source of truth — it already handles gitignore semantics correctly by nature of how git tracks changes

## Resolution

root_cause: gitDirectoryVsHEAD walked the filesystem and used filterGitIgnored (gitignore pattern matcher) to decide which files to include. A negation pattern like `!.idea/auth0-tenant-config.iml` causes the matcher to "un-ignore" the file, so it passes the filter and gets included in the diff candidate list. If the file exists on disk but is not tracked in HEAD (or is tracked but headContent mismatches due to how path is resolved), it incorrectly appears as added/changed even with no real change.

fix: Replaced the entire filesystem walk + gitignore filter approach in gitDirectoryVsHEAD with a single worktree.Status() call. Status() returns only files that git considers actually changed (Modified, Added, Deleted, Staged, etc.). Files that are un-ignored by negation patterns but have no actual changes correctly do not appear in the Status map. Also removed the separate "deleted files" detection loop (covered by Status with Deleted code).

verification: All 294 existing tests pass. Added 2 new regression tests:
  - TestGitDirectoryVsHEAD_negationPatternNoChange: confirms un-ignored but unchanged file is excluded
  - TestGitDirectoryVsHEAD_negationPatternWithChange: confirms un-ignored AND changed file is included

files_changed:
  - cmd/drift/gitworktree.go (replaced gitDirectoryVsHEAD implementation; removed io/fs import)
  - cmd/drift/gitworktree_test.go (added 2 regression tests)
