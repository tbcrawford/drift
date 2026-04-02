package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

// filePair represents a single file entry in a directory diff.
type filePair struct {
	Name    string // relative path with forward slashes (display name)
	OldPath string // absolute path on old side; empty if file is added
	NewPath string // absolute path on new side; empty if file is removed
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

	var pairs []filePair

	// Walk newDir; compare against oldFiles.
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
		relSlash := filepath.ToSlash(rel)

		if oldPath, inOld := oldFiles[relSlash]; inOld {
			// File exists on both sides — compare contents.
			delete(oldFiles, relSlash)
			oldContent, err := os.ReadFile(oldPath)
			if err != nil {
				return fmt.Errorf("reading old %s: %w", relSlash, err)
			}
			newContent, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("reading new %s: %w", relSlash, err)
			}
			if !bytes.Equal(oldContent, newContent) {
				pairs = append(pairs, filePair{
					Name:    relSlash,
					OldPath: oldPath,
					NewPath: path,
				})
			}
		} else {
			// File only in new — added.
			pairs = append(pairs, filePair{
				Name:    relSlash,
				OldPath: "",
				NewPath: path,
			})
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walking new dir: %w", err)
	}

	// Remaining entries in oldFiles exist only on old side — removed.
	for relSlash, oldPath := range oldFiles {
		pairs = append(pairs, filePair{
			Name:    relSlash,
			OldPath: oldPath,
			NewPath: "",
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
