package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	out, stderr, err := runGit(repoRoot, "show", "HEAD:"+relpathSlash)
	if err == nil {
		return out, nil
	}
	if errors.Is(err, exec.ErrNotFound) {
		return "", fmt.Errorf("git not found in PATH; use two paths to diff files without git: %w", err)
	}
	if headBlobMissing(stderr) {
		return "", nil
	}
	return "", fmt.Errorf("git show HEAD:%s failed: use two paths to diff without a working git repo: %w\n%s", relpathSlash, err, strings.TrimSpace(stderr))
}

func headBlobMissing(stderr string) bool {
	s := strings.ToLower(stderr)
	return strings.Contains(s, "exists on disk, but not in") ||
		strings.Contains(s, "did not match any file") ||
		strings.Contains(s, "fatal: path")
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
