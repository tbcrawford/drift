package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// filePair represents a single file entry in a directory diff.
type filePair struct {
	Name       string // relative path with forward slashes (display name)
	OldPath    string // absolute path on old side; empty if file is added
	NewPath    string // absolute path on new side; empty if file is removed
	OldContent string // file content on old side; empty if file is added
	NewContent string // file content on new side; empty if file is removed
}

// IsAdded reports whether the file exists only on the new side.
func (fp filePair) IsAdded() bool { return fp.OldPath == "" }

// IsRemoved reports whether the file exists only on the old side.
func (fp filePair) IsRemoved() bool { return fp.NewPath == "" }

// diffDirectories walks two directory trees and returns a sorted slice of
// filePair values — one per file that differs or is exclusive to one side.
// Identical files (same byte content) are excluded from the result.
// Returns an error if either path is not a directory.
func diffDirectories(oldDir, newDir string) ([]filePair, error) {
	// Validate that both paths are directories.
	if err := requireDir(oldDir); err != nil {
		return nil, fmt.Errorf("old: %w", err)
	}
	if err := requireDir(newDir); err != nil {
		return nil, fmt.Errorf("new: %w", err)
	}

	// Walk oldDir and collect relative-name → absolute-path map.
	oldFiles := make(map[string]string) // relName → absPath
	if err := filepath.WalkDir(oldDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(oldDir, path)
		if err != nil {
			return err
		}
		oldFiles[filepath.ToSlash(rel)] = path
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walking old dir: %w", err)
	}

	// Filter oldFiles through gitignore if oldDir is inside a git repo.
	if repoRoot, ok := isInsideGitRepo(oldDir); ok {
		oldRelPaths := make([]string, 0, len(oldFiles))
		for rel := range oldFiles {
			oldRelPaths = append(oldRelPaths, rel)
		}
		kept, _ := filterGitIgnored(repoRoot, oldRelPaths)
		keptSet := make(map[string]struct{}, len(kept))
		for _, p := range kept {
			keptSet[p] = struct{}{}
		}
		for rel := range oldFiles {
			if _, ok := keptSet[rel]; !ok {
				delete(oldFiles, rel)
			}
		}
	}

	// Walk newDir, collecting all entries first, then filter through gitignore.
	type newEntry struct {
		relSlash string
		absPath  string
	}
	var newEntries []newEntry
	if err := filepath.WalkDir(newDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(newDir, path)
		if err != nil {
			return err
		}
		newEntries = append(newEntries, newEntry{relSlash: filepath.ToSlash(rel), absPath: path})
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walking new dir: %w", err)
	}

	// Filter newEntries through gitignore if newDir is inside a git repo.
	if repoRoot, ok := isInsideGitRepo(newDir); ok {
		newRelPaths := make([]string, 0, len(newEntries))
		for _, e := range newEntries {
			newRelPaths = append(newRelPaths, e.relSlash)
		}
		kept, _ := filterGitIgnored(repoRoot, newRelPaths)
		keptSet := make(map[string]struct{}, len(kept))
		for _, p := range kept {
			keptSet[p] = struct{}{}
		}
		filtered := newEntries[:0]
		for _, e := range newEntries {
			if _, ok := keptSet[e.relSlash]; ok {
				filtered = append(filtered, e)
			}
		}
		newEntries = filtered
	}

	var pairs []filePair

	// Compare newEntries against oldFiles.
	for _, ne := range newEntries {
		if oldPath, inOld := oldFiles[ne.relSlash]; inOld {
			// File exists on both sides — compare contents.
			delete(oldFiles, ne.relSlash)
			oldBytes, err := os.ReadFile(oldPath)
			if err != nil {
				return nil, fmt.Errorf("reading old %s: %w", ne.relSlash, err)
			}
			newBytes, err := os.ReadFile(ne.absPath)
			if err != nil {
				return nil, fmt.Errorf("reading new %s: %w", ne.relSlash, err)
			}
			if !bytes.Equal(oldBytes, newBytes) {
				pairs = append(pairs, filePair{
					Name:       ne.relSlash,
					OldPath:    oldPath,
					NewPath:    ne.absPath,
					OldContent: string(oldBytes),
					NewContent: string(newBytes),
				})
			}
		} else {
			// File only in new — added.
			newBytes, err := os.ReadFile(ne.absPath)
			if err != nil {
				return nil, fmt.Errorf("reading new %s: %w", ne.relSlash, err)
			}
			pairs = append(pairs, filePair{
				Name:       ne.relSlash,
				OldPath:    "",
				NewPath:    ne.absPath,
				OldContent: "",
				NewContent: string(newBytes),
			})
		}
	}

	// Remaining entries in oldFiles exist only on old side — removed. Read content now.
	for relSlash, oldPath := range oldFiles {
		oldBytes, err := os.ReadFile(oldPath)
		if err != nil {
			return nil, fmt.Errorf("reading old %s: %w", relSlash, err)
		}
		pairs = append(pairs, filePair{
			Name:       relSlash,
			OldPath:    oldPath,
			NewPath:    "",
			OldContent: string(oldBytes),
			NewContent: "",
		})
	}

	// Sort by Name for deterministic output.
	slices.SortFunc(pairs, func(a, b filePair) int {
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

// requireDir returns an error if path does not exist or is not a directory.
func requireDir(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%q is not a directory", path)
	}
	return nil
}

// isInsideGitRepo reports whether dir is inside a git worktree and returns the
// repo root if so. Returns ("", false) if not in a repo or if git is unavailable.
func isInsideGitRepo(dir string) (repoRoot string, ok bool) {
	inside, err := gitRevParseIsInsideWorkTree(dir)
	if err != nil || inside != "true" {
		return "", false
	}
	root, err := gitRevParseShowToplevel(dir)
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(root), true
}
