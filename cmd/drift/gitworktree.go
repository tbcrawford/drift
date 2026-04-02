// Performance: changedFilesViaIndex replaces wt.Status() (go-git v5.x).
// On auth0-tenant-config (158 files, 25 changed):
//   - Before (wt.Status()):          4,512ms wall-clock
//   - After  (changedFilesViaIndex):   134ms wall-clock  (drift --split, no pager)
//   - Reference (git diff | delta):    167ms wall-clock
//
// wt.Status() is slow because it does a full two-pass filesystem walk: walks every
// file via os.ReadDir and reads SHA1 for each. Go-git does not implement git's
// inode+mtime "stat cache" optimization, making it 200x slower on large repos.
// changedFilesViaIndex uses the index entry's stored mtime (entry.ModifiedAt) to
// detect unstaged changes without reading file content, and compares index hashes
// against HEAD tree hashes to detect staged changes.
// Baseline measured 2026-04-02.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// openRepoAt opens the git repository containing startPath (searching parent dirs).
// Returns the repo and the worktree root path.
func openRepoAt(startPath string) (*git.Repository, string, error) {
	repo, err := git.PlainOpenWithOptions(startPath, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, "", err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return nil, "", err
	}
	return repo, wt.Filesystem.Root(), nil
}

// headCommitTree returns the tree object for the HEAD commit.
func headCommitTree(repo *git.Repository) (*object.Tree, error) {
	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}
	return commit.Tree()
}

// gitRevParseIsInsideWorkTree reports whether dir is inside a git worktree.
// Returns "true" if inside, "false" if not a git repo, error for unexpected failures.
func gitRevParseIsInsideWorkTree(dir string) (string, error) {
	_, _, err := openRepoAt(dir)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		return "false", nil
	}
	if err != nil {
		return "", fmt.Errorf("git open failed: %w", err)
	}
	return "true", nil
}

// gitRevParseShowToplevel returns the repository root for the repo containing dir.
func gitRevParseShowToplevel(dir string) (string, error) {
	_, root, err := openRepoAt(dir)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		return "", fmt.Errorf("not a git repository: %s", dir)
	}
	if err != nil {
		return "", fmt.Errorf("git open failed: %w", err)
	}
	return root, nil
}

// gitShowHEADBlob reads the content of relPathSlash from the HEAD commit.
// Returns "" if the file does not exist at HEAD (new/untracked file — treat old as empty).
func gitShowHEADBlob(repoRoot, relpathSlash string) (string, error) {
	repo, _, err := openRepoAt(repoRoot)
	if err != nil {
		return "", fmt.Errorf("opening repo at %s: %w", repoRoot, err)
	}

	tree, err := headCommitTree(repo)
	if err != nil {
		// No HEAD commit (empty repo) — file definitely doesn't exist at HEAD.
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return "", nil
		}
		return "", fmt.Errorf("reading HEAD tree: %w", err)
	}

	f, err := tree.File(relpathSlash)
	if err != nil {
		if errors.Is(err, object.ErrFileNotFound) {
			return "", nil
		}
		return "", fmt.Errorf("finding file %s in HEAD: %w", relpathSlash, err)
	}

	content, err := f.Contents()
	if err != nil {
		return "", fmt.Errorf("reading file %s from HEAD: %w", relpathSlash, err)
	}
	return content, nil
}

// filterGitIgnored returns relPaths with any gitignored entries removed.
// Fail-open: if the repo cannot be opened or gitignore cannot be read, the
// original relPaths slice is returned unchanged.
func filterGitIgnored(repoRoot string, relPaths []string) ([]string, error) {
	if len(relPaths) == 0 {
		return nil, nil
	}

	repo, _, err := openRepoAt(repoRoot)
	if err != nil {
		// Fail-open: if not a git repo or any error, return all paths unchanged.
		return relPaths, nil
	}

	wt, err := repo.Worktree()
	if err != nil {
		return relPaths, nil
	}

	patterns, err := gitignore.ReadPatterns(wt.Filesystem, nil)
	if err != nil {
		return relPaths, nil
	}
	matcher := gitignore.NewMatcher(patterns)

	kept := make([]string, 0, len(relPaths))
	for _, p := range relPaths {
		parts := strings.Split(p, "/")
		// Match as a file (not a directory).
		if !matcher.Match(parts, false) {
			kept = append(kept, p)
		}
	}
	return kept, nil
}

// resolveGitWorkingTreeVsHEAD loads OLD from the HEAD commit blob (or "" if the
// path has no blob at HEAD) and NEW from the working tree file at abs path.
func resolveGitWorkingTreeVsHEAD(path string) (old, newText string, oldName, newName string, err error) {
	absFile, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", "", "", "", fmt.Errorf("invalid file %q: %w", path, err)
	}
	st, err := os.Stat(absFile)
	if err != nil {
		return "", "", "", "", fmt.Errorf("invalid file %q: %w", path, err)
	}
	if !st.Mode().IsRegular() {
		return "", "", "", "", fmt.Errorf("invalid file %q: not a regular file", path)
	}

	gitDir := filepath.Dir(absFile)
	inside, err := gitRevParseIsInsideWorkTree(gitDir)
	if err != nil {
		return "", "", "", "", err
	}
	if inside != "true" {
		return "", "", "", "", fmt.Errorf("not a git worktree: use two paths to diff files; repository check failed")
	}

	repoRoot, err := gitRevParseShowToplevel(gitDir)
	if err != nil {
		return "", "", "", "", err
	}
	repoRoot = strings.TrimSpace(repoRoot)

	rel, err := filepath.Rel(repoRoot, absFile)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", "", "", "", fmt.Errorf("invalid file: outside repository root %q", repoRoot)
	}

	newBytes, err := os.ReadFile(absFile)
	if err != nil {
		return "", "", "", "", fmt.Errorf("invalid file %q: %w", absFile, err)
	}
	newText = string(newBytes)

	relpathSlash := filepath.ToSlash(rel)
	old, err = gitShowHEADBlob(repoRoot, relpathSlash)
	if err != nil {
		return "", "", "", "", err
	}

	base := filepath.Base(absFile)
	oldName = base + " (HEAD)"
	newName = base + " (working tree)"
	return old, newText, oldName, newName, nil
}

// gitShowHEADBlobFromTree reads the content of relpathSlash from an already-loaded
// HEAD tree. Returns "" if the file does not exist in the tree (new file).
// Use this inside directory diff loops to avoid re-opening the repo per file.
func gitShowHEADBlobFromTree(tree *object.Tree, relpathSlash string) (string, error) {
	if tree == nil {
		return "", nil
	}
	f, err := tree.File(relpathSlash)
	if err != nil {
		if errors.Is(err, object.ErrFileNotFound) {
			return "", nil
		}
		return "", fmt.Errorf("finding file %s in HEAD: %w", relpathSlash, err)
	}
	content, err := f.Contents()
	if err != nil {
		return "", fmt.Errorf("reading file %s from HEAD: %w", relpathSlash, err)
	}
	return content, nil
}

// changedFilesViaIndex returns the list of files that differ between the git index
// (or HEAD) and the working tree, using the index entry's stored mtime to detect
// unstaged changes without reading file content.
//
// It also returns the HEAD *object.Tree (nil for repos with no HEAD commit) so
// callers can read blobs without re-opening the repo, the worktree filesystem root,
// and any error.
//
// Algorithm:
//  1. Build a map of filename → HEAD hash by iterating the HEAD tree.
//  2. Iterate index entries. For each entry:
//     a. If the entry's hash differs from HEAD hash → staged change → include.
//     b. If the entry is not in HEAD → new staged file → include.
//     c. Otherwise stat the file: if missing → deleted → include.
//     d. If mtime differs from entry.ModifiedAt → potential unstaged change → include.
func changedFilesViaIndex(repoRoot string) (paths []string, headTree *object.Tree, wtRoot string, err error) {
	repo, err := git.PlainOpenWithOptions(repoRoot, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, nil, "", err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return nil, nil, "", err
	}
	wtRoot = wt.Filesystem.Root()

	// Build HEAD hash map. Nil headHashes means no HEAD commit (empty repo).
	var headHashes map[string]string
	ref, headErr := repo.Head()
	if headErr == nil {
		commit, cErr := repo.CommitObject(ref.Hash())
		if cErr == nil {
			headTree, _ = commit.Tree()
			if headTree != nil {
				headHashes = make(map[string]string)
				_ = headTree.Files().ForEach(func(f *object.File) error {
					headHashes[f.Name] = f.Hash.String()
					return nil
				})
			}
		}
	}

	idx, err := repo.Storer.Index()
	if err != nil {
		return nil, headTree, wtRoot, err
	}

	for _, entry := range idx.Entries {
		headHash, inHead := headHashes[entry.Name]
		if headHashes == nil || !inHead {
			// Not in HEAD → new staged file.
			paths = append(paths, entry.Name)
			continue
		}
		if headHash != entry.Hash.String() {
			// Staged change: index hash differs from HEAD hash.
			paths = append(paths, entry.Name)
			continue
		}
		// File is clean in staging — check working tree via mtime.
		absPath := filepath.Join(wtRoot, filepath.FromSlash(entry.Name))
		info, statErr := os.Stat(absPath)
		if statErr != nil {
			// File deleted from working tree.
			paths = append(paths, entry.Name)
			continue
		}
		if !info.ModTime().Equal(entry.ModifiedAt) {
			// mtime differs → potential unstaged change.
			paths = append(paths, entry.Name)
		}
	}
	return paths, headTree, wtRoot, nil
}

// gitDirectoryVsHEAD uses the git index to get the list of actually-changed
// files under absDir (relative to the repo root), then fetches HEAD blob and
// working-tree content for each. Only files with real changes (modified, added,
// deleted, staged) are returned.
// absDir must be inside a git worktree.
func gitDirectoryVsHEAD(absDir string) ([]gitFilePair, error) {
	absDir = filepath.Clean(absDir)

	// Get changed files using the fast index+mtime approach. changedFilesViaIndex
	// opens the repo (with DetectDotGit) so we don't need a separate openRepoAt call.
	changedPaths, headTree, wtRoot, err := changedFilesViaIndex(absDir)
	if err != nil {
		// Distinguish "not a git repo" from other errors.
		if errors.Is(err, git.ErrRepositoryNotExists) {
			return nil, fmt.Errorf("not a git worktree: %q: %w", absDir, err)
		}
		return nil, fmt.Errorf("reading changed files: %w", err)
	}

	// Compute the path of absDir relative to the worktree root.
	dirRel, err := filepath.Rel(wtRoot, absDir)
	if err != nil || strings.HasPrefix(dirRel, "..") {
		return nil, fmt.Errorf("directory %q is outside repository root", absDir)
	}
	dirRelSlash := filepath.ToSlash(dirRel)

	// Scope prefix: only consider files under the requested directory.
	prefix := dirRelSlash + "/"
	if dirRelSlash == "." {
		prefix = ""
	}
	dirBase := filepath.Base(absDir)

	var pairs []gitFilePair

	for _, repoRelSlash := range changedPaths {
		// Skip files outside the requested directory subtree.
		if prefix != "" && !strings.HasPrefix(repoRelSlash, prefix) {
			continue
		}

		// Determine the path relative to the requested directory (for display).
		relToDir := strings.TrimPrefix(repoRelSlash, prefix)

		// Fetch the HEAD blob using the already-loaded tree (no extra repo open).
		headContent, hErr := gitShowHEADBlobFromTree(headTree, repoRelSlash)
		if hErr != nil {
			return nil, hErr
		}

		// Fetch the working-tree content (empty string if file is deleted).
		var workContent string
		absPath := filepath.Join(wtRoot, filepath.FromSlash(repoRelSlash))
		workBytes, rErr := os.ReadFile(absPath)
		if rErr == nil {
			workContent = string(workBytes)
		}

		// Safety net: if contents are identical after reading, skip.
		if headContent == workContent {
			continue
		}

		oldName := dirBase + "/" + relToDir + " (HEAD)"
		newName := dirBase + "/" + relToDir + " (working tree)"
		pairs = append(pairs, gitFilePair{
			Name:        relToDir,
			HeadContent: headContent,
			WorkContent: workContent,
			OldName:     oldName,
			NewName:     newName,
		})
	}

	// Sort by Name for deterministic output.
	slices.SortFunc(pairs, func(a, b gitFilePair) int {
		if a.Name < b.Name {
			return -1
		}
		if a.Name > b.Name {
			return 1
		}
		return 0
	})

	return pairs, nil
}

// gitFilePair holds a file's HEAD and working-tree content for a git directory diff.
type gitFilePair struct {
	Name        string // relative path (display)
	HeadContent string // content at HEAD (empty if file is new)
	WorkContent string // content in working tree (empty if file is deleted)
	OldName     string // display name for old side
	NewName     string // display name for new side
}
