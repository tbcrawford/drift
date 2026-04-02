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

// gitDirectoryVsHEAD walks absDir, compares each file against its HEAD blob,
// and returns a sorted slice of gitFilePair values for files that differ from HEAD
// (or are new/deleted relative to HEAD). absDir must be inside a git worktree.
func gitDirectoryVsHEAD(absDir string) ([]gitFilePair, error) {
	absDir = filepath.Clean(absDir)

	// Confirm inside a git worktree.
	inside, err := gitRevParseIsInsideWorkTree(absDir)
	if err != nil {
		return nil, err
	}
	if inside != "true" {
		return nil, fmt.Errorf("not a git worktree: %q", absDir)
	}

	repo, repoRoot, err := openRepoAt(absDir)
	if err != nil {
		return nil, err
	}

	// Compute the path of absDir relative to the repo root.
	dirRel, err := filepath.Rel(repoRoot, absDir)
	if err != nil || strings.HasPrefix(dirRel, "..") {
		return nil, fmt.Errorf("directory %q is outside repository root", absDir)
	}
	dirRelSlash := filepath.ToSlash(dirRel)

	// Walk the working-tree directory to collect all current files.
	type workFile struct{ rel, abs string }
	var workFiles []workFile
	if err := filepath.WalkDir(absDir, func(path string, d os.DirEntry, wErr error) error {
		if wErr != nil {
			return wErr
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, rErr := filepath.Rel(absDir, path)
		if rErr != nil {
			return rErr
		}
		workFiles = append(workFiles, workFile{rel: filepath.ToSlash(rel), abs: path})
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walking directory: %w", err)
	}

	// Filter out gitignored files from the working-tree list.
	repoRelPaths := make([]string, 0, len(workFiles))
	for _, wf := range workFiles {
		repoRel := dirRelSlash + "/" + wf.rel
		if dirRelSlash == "." {
			repoRel = wf.rel
		}
		repoRelPaths = append(repoRelPaths, repoRel)
	}
	keptRepoPaths, _ := filterGitIgnored(repoRoot, repoRelPaths)
	keptSet := make(map[string]struct{}, len(keptRepoPaths))
	for _, p := range keptRepoPaths {
		keptSet[p] = struct{}{}
	}
	filteredWorkFiles := workFiles[:0]
	for _, wf := range workFiles {
		repoRel := dirRelSlash + "/" + wf.rel
		if dirRelSlash == "." {
			repoRel = wf.rel
		}
		if _, ok := keptSet[repoRel]; ok {
			filteredWorkFiles = append(filteredWorkFiles, wf)
		}
	}
	workFiles = filteredWorkFiles

	// Build a set of working-tree relative paths for deleted-file detection.
	workSet := make(map[string]string, len(workFiles))
	for _, wf := range workFiles {
		workSet[wf.rel] = wf.abs
	}

	var pairs []gitFilePair

	// For each file in the working tree, compare against HEAD.
	for _, wf := range workFiles {
		repoRel := dirRelSlash + "/" + wf.rel
		if dirRelSlash == "." {
			repoRel = wf.rel
		}

		headContent, hErr := gitShowHEADBlob(repoRoot, repoRel)
		if hErr != nil {
			return nil, hErr
		}

		workBytes, rErr := os.ReadFile(wf.abs)
		if rErr != nil {
			return nil, fmt.Errorf("reading %s: %w", wf.abs, rErr)
		}
		workContent := string(workBytes)

		if headContent == workContent {
			continue // identical — skip
		}

		oldName := filepath.Base(absDir) + "/" + wf.rel + " (HEAD)"
		newName := filepath.Base(absDir) + "/" + wf.rel + " (working tree)"
		pairs = append(pairs, gitFilePair{
			Name:        wf.rel,
			HeadContent: headContent,
			WorkContent: workContent,
			OldName:     oldName,
			NewName:     newName,
		})
	}

	// Detect files that exist at HEAD but are deleted in the working tree.
	// Use go-git tree.Files() iterator to enumerate HEAD blobs.
	tree, tErr := headCommitTree(repo)
	if tErr == nil {
		prefix := dirRelSlash + "/"
		if dirRelSlash == "." {
			prefix = ""
		}
		fIter := tree.Files()
		_ = fIter.ForEach(func(f *object.File) error {
			name := f.Name
			// Only consider files under the target directory.
			if prefix != "" && !strings.HasPrefix(name, prefix) {
				return nil
			}
			relToDir := strings.TrimPrefix(name, prefix)
			if _, inWork := workSet[relToDir]; inWork {
				return nil // present in working tree — already handled
			}
			// File exists at HEAD but not on disk — deleted.
			headContent, hErr := f.Contents()
			if hErr != nil {
				return nil // skip on error
			}
			oldName := filepath.Base(absDir) + "/" + relToDir + " (HEAD)"
			newName := filepath.Base(absDir) + "/" + relToDir + " (working tree)"
			pairs = append(pairs, gitFilePair{
				Name:        relToDir,
				HeadContent: headContent,
				WorkContent: "",
				OldName:     oldName,
				NewName:     newName,
			})
			return nil
		})
	}
	// If HEAD doesn't exist (empty repo), there are no deleted files to detect — skip.

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
