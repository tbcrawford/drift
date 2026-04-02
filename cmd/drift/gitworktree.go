package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// resolveGitWorkingTreeVsHEAD loads OLD from git show HEAD:<relpath> (or "" if the path
// has no blob at HEAD) and NEW from the working tree file at abs path.
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

func gitEnv() []string {
	return append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
}

func gitRevParseIsInsideWorkTree(gitDir string) (string, error) {
	out, stderr, err := runGit(gitDir, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("git not found in PATH; use two paths to diff files without git: %w", err)
		}
		return "", fmt.Errorf("git rev-parse failed: %w\n%s", err, strings.TrimSpace(stderr))
	}
	return strings.TrimSpace(out), nil
}

func gitRevParseShowToplevel(gitDir string) (string, error) {
	out, stderr, err := runGit(gitDir, "rev-parse", "--show-toplevel")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("git not found in PATH; use two paths to diff files without git: %w", err)
		}
		return "", fmt.Errorf("git rev-parse failed: %w\n%s", err, strings.TrimSpace(stderr))
	}
	return out, nil
}

func gitShowHEADBlob(repoRoot, relpathSlash string) (string, error) {
	// Use git cat-file -e to check existence before show — this is locale-independent
	// (exit code 0 = object exists, non-zero = absent), unlike parsing English stderr.
	_, _, checkErr := runGit(repoRoot, "cat-file", "-e", "HEAD:"+relpathSlash)
	if checkErr != nil {
		if errors.Is(checkErr, exec.ErrNotFound) {
			return "", fmt.Errorf("git not found in PATH; use two paths to diff files without git: %w", checkErr)
		}
		// Object not present at HEAD — file is new/untracked; treat old content as empty.
		return "", nil
	}

	out, stderr, err := runGit(repoRoot, "show", "HEAD:"+relpathSlash)
	if err == nil {
		return out, nil
	}
	if errors.Is(err, exec.ErrNotFound) {
		return "", fmt.Errorf("git not found in PATH; use two paths to diff files without git: %w", err)
	}
	return "", fmt.Errorf("git show HEAD:%s failed: use two paths to diff without a working git repo: %w\n%s", relpathSlash, err, strings.TrimSpace(stderr))
}

// gitDirectoryVsHEAD walks absDir, compares each file against its HEAD blob,
// and returns a sorted slice of filePair values for files that differ from HEAD
// (or are new/deleted relative to HEAD). absDir must be inside a git worktree.
// OldPath is always empty — HEAD content is returned inline via gitShowHEADBlob
// so callers read it from the pair's HeadContent field (see gitFilePair).
//
// To keep the filePair type unchanged, we return a gitFilePair slice instead.
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

	repoRoot, err := gitRevParseShowToplevel(absDir)
	if err != nil {
		return nil, err
	}
	repoRoot = strings.TrimSpace(repoRoot)

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
	// Built after gitignore filtering so ignored files don't mask deleted entries.
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
	// Use git ls-tree to list files under the directory at HEAD.
	prefix := dirRelSlash + "/"
	if dirRelSlash == "." {
		prefix = ""
	}
	lsOut, _, lsErr := runGit(repoRoot, "ls-tree", "-r", "--name-only", "HEAD", "--", dirRelSlash)
	if lsErr == nil {
		for _, line := range strings.Split(strings.TrimSpace(lsOut), "\n") {
			if line == "" {
				continue
			}
			relToDir := strings.TrimPrefix(line, prefix)
			if relToDir == line && prefix != "" {
				continue // not under our directory
			}
			if _, inWork := workSet[relToDir]; inWork {
				continue // present in working tree — already handled
			}
			// File exists at HEAD but not on disk — deleted.
			headContent, hErr := gitShowHEADBlob(repoRoot, line)
			if hErr != nil {
				return nil, hErr
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
		}
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

func runGit(gitDir string, args ...string) (stdout, stderr string, err error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return "", "", err
	}
	cmd := exec.Command(gitPath, append([]string{"-C", gitDir}, args...)...)
	cmd.Env = gitEnv()
	var outb, errb strings.Builder
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	return outb.String(), errb.String(), err
}

// runGitStdin is like runGit but feeds stdin to the git process.
func runGitStdin(gitDir, stdin string, args ...string) (stdout, stderr string, err error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return "", "", err
	}
	cmd := exec.Command(gitPath, append([]string{"-C", gitDir}, args...)...)
	cmd.Env = gitEnv()
	cmd.Stdin = strings.NewReader(stdin)
	var outb, errb strings.Builder
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	return outb.String(), errb.String(), err
}

// filterGitIgnored runs `git check-ignore -z --stdin` in repoRoot and returns
// relPaths with any gitignored entries removed. Fail-open: if git is not
// found, the input exits 1 with no output (nothing ignored), or any other
// unexpected error occurs, the original relPaths slice is returned unchanged.
func filterGitIgnored(repoRoot string, relPaths []string) ([]string, error) {
	if len(relPaths) == 0 {
		return nil, nil
	}
	// NUL-separated stdin.
	stdin := strings.Join(relPaths, "\x00") + "\x00"
	out, _, err := runGitStdin(repoRoot, stdin, "check-ignore", "-z", "--stdin")
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			// git not found or other non-exit error — fail open.
			return relPaths, nil
		}
		if exitErr.ExitCode() != 1 {
			// Unexpected exit code — fail open.
			return relPaths, nil
		}
		// Exit code 1 means "no paths are ignored" — not an error.
	}
	// Parse NUL-separated ignored paths into a set.
	ignoredSet := make(map[string]struct{})
	for _, p := range strings.Split(out, "\x00") {
		if p != "" {
			ignoredSet[p] = struct{}{}
		}
	}
	// Return paths not in the ignored set.
	kept := make([]string, 0, len(relPaths))
	for _, p := range relPaths {
		if _, ignored := ignoredSet[p]; !ignored {
			kept = append(kept, p)
		}
	}
	return kept, nil
}
