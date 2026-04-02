package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// testSig returns a fixed author signature for test commits.
func testSig() *object.Signature {
	return &object.Signature{
		Name:  "Test",
		Email: "test@test.com",
		When:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

// makeTestRepo creates a real git repo in a temp dir with the given files,
// stages and commits them all, and returns the repo directory.
func makeTestRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()

	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("PlainInit: %v", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Worktree: %v", err)
	}

	for name, content := range files {
		absPath := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			t.Fatalf("MkdirAll for %s: %v", name, err)
		}
		if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile %s: %v", name, err)
		}
		if _, err := wt.Add(name); err != nil {
			t.Fatalf("wt.Add %s: %v", name, err)
		}
	}

	if _, err := wt.Commit("initial commit", &git.CommitOptions{Author: testSig()}); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	return dir
}

// --- filterGitIgnored tests ---

func TestFilterGitIgnored_emptyInput(t *testing.T) {
	// Empty input must return nil without opening a repo.
	got, err := filterGitIgnored(t.TempDir(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

func TestFilterGitIgnored_noIgnored(t *testing.T) {
	// Real repo with no .gitignore → all paths returned.
	repoDir := makeTestRepo(t, map[string]string{
		"main.go":   "package main",
		"README.md": "# readme",
	})

	paths := []string{"main.go", "README.md"}
	got, err := filterGitIgnored(repoDir, paths)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != "main.go" || got[1] != "README.md" {
		t.Fatalf("expected all paths returned; got %v", got)
	}
}

func TestFilterGitIgnored_someIgnored(t *testing.T) {
	// Real repo with .gitignore containing "dist/" → dist/app excluded.
	repoDir := makeTestRepo(t, map[string]string{
		"main.go":    "package main",
		"README.md":  "# readme",
		".gitignore": "dist/\n",
	})

	// Write dist/app to disk (not committed — gitignore matches regardless).
	distDir := filepath.Join(repoDir, "dist")
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(distDir, "app"), []byte("artifact"), 0o644); err != nil {
		t.Fatal(err)
	}

	paths := []string{"main.go", "dist/app", "README.md"}
	got, err := filterGitIgnored(repoDir, paths)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"main.go", "README.md"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i, p := range want {
		if got[i] != p {
			t.Errorf("got[%d] = %q, want %q", i, got[i], p)
		}
	}
}

func TestFilterGitIgnored_notInRepo(t *testing.T) {
	// Non-repo dir → fail-open, return all paths unchanged.
	plainDir := t.TempDir()
	paths := []string{"a.go", "b.go"}
	got, err := filterGitIgnored(plainDir, paths)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected fail-open (all paths); got %v", got)
	}
}

// --- gitDirectoryVsHEAD gitignore integration ---

func TestGitDirectoryVsHEAD_skipsIgnored(t *testing.T) {
	// Create repo: commit keep.go + .gitignore (dist/); modify keep.go on disk; add dist/app (not committed).
	repoDir := makeTestRepo(t, map[string]string{
		"keep.go":    "oldcontent",
		".gitignore": "dist/\n",
	})

	// Modify keep.go in working tree (not staged/committed).
	if err := os.WriteFile(filepath.Join(repoDir, "keep.go"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write dist/app to disk (not staged — ignored).
	distDir := filepath.Join(repoDir, "dist")
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(distDir, "app"), []byte("artifact"), 0o644); err != nil {
		t.Fatal(err)
	}

	pairs, err := gitDirectoryVsHEAD(repoDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// dist/app must not appear (it is gitignored).
	for _, p := range pairs {
		if strings.Contains(p.Name, "dist") || strings.Contains(p.Name, "app") {
			t.Errorf("ignored file dist/app appeared in pairs: %+v", p)
		}
	}

	// keep.go must appear with correct content.
	found := false
	for _, p := range pairs {
		if p.Name == "keep.go" {
			found = true
			if p.HeadContent != "oldcontent" {
				t.Errorf("keep.go HeadContent = %q, want %q", p.HeadContent, "oldcontent")
			}
			if p.WorkContent != "changed" {
				t.Errorf("keep.go WorkContent = %q, want %q", p.WorkContent, "changed")
			}
		}
	}
	if !found {
		t.Errorf("expected keep.go in pairs, got: %+v", pairs)
	}
}

func TestGitDirectoryVsHEAD_negationPatternNoChange(t *testing.T) {
	// Regression test: a negation pattern like `!.idea/auth.iml` un-ignores the
	// file so it is visible on disk, but if the file has not actually changed
	// relative to HEAD it must NOT appear in the diff output.
	repoDir := makeTestRepo(t, map[string]string{
		"main.go":        "package main",
		".idea/auth.iml": "committed content",
		".gitignore":     ".idea/\n!.idea/auth.iml\n",
	})

	// The file matches the negation — it's tracked and committed. Do NOT modify
	// it on disk. Only modify main.go so there is at least one real change.
	if err := os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\n// changed"), 0o644); err != nil {
		t.Fatal(err)
	}

	pairs, err := gitDirectoryVsHEAD(repoDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// .idea/auth.iml must NOT appear — it hasn't changed.
	for _, p := range pairs {
		if strings.Contains(p.Name, "auth.iml") || strings.Contains(p.Name, ".idea") {
			t.Errorf("unchanged negation-pattern file appeared in pairs: %+v", p)
		}
	}

	// main.go must appear because it was actually modified.
	found := false
	for _, p := range pairs {
		if p.Name == "main.go" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected main.go in pairs (it was modified), got: %+v", pairs)
	}
}

func TestGitDirectoryVsHEAD_negationPatternWithChange(t *testing.T) {
	// A file un-ignored by a negation pattern that IS actually changed must
	// appear in the diff output.
	repoDir := makeTestRepo(t, map[string]string{
		".idea/auth.iml": "old content",
		".gitignore":     ".idea/\n!.idea/auth.iml\n",
	})

	// Modify the negated file on disk.
	ideaDir := filepath.Join(repoDir, ".idea")
	if err := os.WriteFile(filepath.Join(ideaDir, "auth.iml"), []byte("new content"), 0o644); err != nil {
		t.Fatal(err)
	}

	pairs, err := gitDirectoryVsHEAD(repoDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, p := range pairs {
		if p.Name == ".idea/auth.iml" {
			found = true
			if p.HeadContent != "old content" {
				t.Errorf(".idea/auth.iml HeadContent = %q, want %q", p.HeadContent, "old content")
			}
			if p.WorkContent != "new content" {
				t.Errorf(".idea/auth.iml WorkContent = %q, want %q", p.WorkContent, "new content")
			}
		}
	}
	if !found {
		t.Errorf("expected .idea/auth.iml in pairs (it was modified), got: %+v", pairs)
	}
}

// --- changedFilesViaIndex tests ---

// modifyFile overwrites a file's content on disk and bumps its mtime by 1 second
// to reliably trigger the mtime-based change detection in changedFilesViaIndex.
func modifyFile(t *testing.T, absPath, content string) {
	t.Helper()
	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", absPath, err)
	}
	future := time.Now().Add(time.Second)
	if err := os.Chtimes(absPath, future, future); err != nil {
		t.Fatalf("Chtimes %s: %v", absPath, err)
	}
}

func TestChangedFilesViaIndex_noChanges(t *testing.T) {
	// Committed repo with no disk modifications → empty result.
	repoDir := makeTestRepo(t, map[string]string{
		"file.txt": "hello",
	})
	paths, _, _, err := changedFilesViaIndex(repoDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 0 {
		t.Fatalf("expected no changed files, got %v", paths)
	}
}

func TestChangedFilesViaIndex_modifiedWorktree(t *testing.T) {
	// file.txt committed, then modified on disk → returns ["file.txt"].
	repoDir := makeTestRepo(t, map[string]string{
		"file.txt": "original",
	})
	modifyFile(t, filepath.Join(repoDir, "file.txt"), "modified")

	paths, _, _, err := changedFilesViaIndex(repoDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 1 || paths[0] != "file.txt" {
		t.Fatalf("expected [file.txt], got %v", paths)
	}
}

func TestChangedFilesViaIndex_staged(t *testing.T) {
	// file2.txt added and staged but not committed → returns ["file2.txt"].
	repoDir := makeTestRepo(t, map[string]string{
		"file.txt": "original",
	})

	// Stage a new file.
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		t.Fatalf("PlainOpen: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Worktree: %v", err)
	}
	newFile := filepath.Join(repoDir, "file2.txt")
	if err := os.WriteFile(newFile, []byte("new file"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := wt.Add("file2.txt"); err != nil {
		t.Fatalf("wt.Add: %v", err)
	}

	paths, _, _, err := changedFilesViaIndex(repoDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 1 || paths[0] != "file2.txt" {
		t.Fatalf("expected [file2.txt], got %v", paths)
	}
}

func TestChangedFilesViaIndex_deleted(t *testing.T) {
	// file.txt committed then deleted from disk → returns ["file.txt"].
	repoDir := makeTestRepo(t, map[string]string{
		"file.txt": "original",
	})
	if err := os.Remove(filepath.Join(repoDir, "file.txt")); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	paths, _, _, err := changedFilesViaIndex(repoDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 1 || paths[0] != "file.txt" {
		t.Fatalf("expected [file.txt], got %v", paths)
	}
}

func TestChangedFilesViaIndex_deletedAndNew(t *testing.T) {
	// old.txt committed+deleted from disk; new.txt staged → both returned.
	repoDir := makeTestRepo(t, map[string]string{
		"old.txt": "old content",
	})

	// Delete old.txt from disk.
	if err := os.Remove(filepath.Join(repoDir, "old.txt")); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	// Stage new.txt.
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		t.Fatalf("PlainOpen: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Worktree: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "new.txt"), []byte("new"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := wt.Add("new.txt"); err != nil {
		t.Fatalf("wt.Add: %v", err)
	}

	paths, _, _, err := changedFilesViaIndex(repoDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 changed files, got %v", paths)
	}
	// Both files should be present (order may vary).
	found := map[string]bool{}
	for _, p := range paths {
		found[p] = true
	}
	if !found["old.txt"] || !found["new.txt"] {
		t.Fatalf("expected old.txt and new.txt in result, got %v", paths)
	}
}

func TestChangedFilesViaIndex_subdir(t *testing.T) {
	// sub/a.tf and sub/b.tf committed; sub/a.tf modified → returns ["sub/a.tf"].
	repoDir := makeTestRepo(t, map[string]string{
		"sub/a.tf": "resource a {}",
		"sub/b.tf": "resource b {}",
	})
	modifyFile(t, filepath.Join(repoDir, "sub", "a.tf"), "resource a { modified = true }")

	paths, _, _, err := changedFilesViaIndex(repoDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 1 || paths[0] != "sub/a.tf" {
		t.Fatalf("expected [sub/a.tf], got %v", paths)
	}
}

func BenchmarkChangedFilesViaIndex(b *testing.B) {
	// Create a 200-file repo with 20 modified files to benchmark change detection.
	files := make(map[string]string, 200)
	for i := range 200 {
		files[fmt.Sprintf("dir/file%03d.go", i)] = fmt.Sprintf("package main\n// file %d\n", i)
	}

	// Inline repo creation using b.TempDir() (makeTestRepo requires *testing.T).
	dir := b.TempDir()
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		b.Fatalf("PlainInit: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		b.Fatalf("Worktree: %v", err)
	}
	for name, content := range files {
		absPath := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			b.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
			b.Fatalf("WriteFile: %v", err)
		}
		if _, err := wt.Add(name); err != nil {
			b.Fatalf("wt.Add: %v", err)
		}
	}
	if _, err := wt.Commit("bench commit", &git.CommitOptions{Author: testSig()}); err != nil {
		b.Fatalf("Commit: %v", err)
	}

	// Modify 20 files on disk with mtime 1s ahead to trigger change detection.
	for i := range 20 {
		p := filepath.Join(dir, fmt.Sprintf("dir/file%03d.go", i))
		if err := os.WriteFile(p, []byte(fmt.Sprintf("package main\n// modified %d\n", i)), 0o644); err != nil {
			b.Fatalf("WriteFile: %v", err)
		}
		future := time.Now().Add(time.Second)
		if err := os.Chtimes(p, future, future); err != nil {
			b.Fatalf("Chtimes: %v", err)
		}
	}

	b.ResetTimer()
	for range b.N {
		changed, _, _, err := changedFilesViaIndex(dir)
		if err != nil {
			b.Fatal(err)
		}
		if len(changed) < 20 {
			b.Fatalf("expected ≥20 changed files, got %d", len(changed))
		}
	}
}

func TestResolveGitWorkingTreeVsHEAD_happyPath(t *testing.T) {
	// Create repo with file.txt = "oldcontent", then modify to "new" on disk.
	repoDir := makeTestRepo(t, map[string]string{
		"file.txt": "oldcontent",
	})

	// Overwrite with new content (not staged/committed).
	filePath := filepath.Join(repoDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	old, newText, oldName, newName, err := resolveGitWorkingTreeVsHEAD(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if old != "oldcontent" || newText != "new" {
		t.Fatalf("content old=%q new=%q", old, newText)
	}
	if !strings.HasSuffix(oldName, "(HEAD)") || !strings.HasSuffix(newName, "(working tree)") {
		t.Fatalf("names oldName=%q newName=%q", oldName, newName)
	}
}

func TestResolveGitWorkingTreeVsHEAD_notInRepo(t *testing.T) {
	// Plain temp dir — not a git repo.
	plainDir := t.TempDir()
	filePath := filepath.Join(plainDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, _, _, _, err := resolveGitWorkingTreeVsHEAD(filePath)
	if err == nil {
		t.Fatal("expected error for non-repo dir")
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "git") || !strings.Contains(msg, "two") {
		t.Fatalf("error should mention git and two paths: %v", err)
	}
}

func TestResolveGitWorkingTreeVsHEAD_missingHEADBlob(t *testing.T) {
	// Empty repo (no commits) — file.txt on disk but not in HEAD.
	dir := t.TempDir()
	if _, err := git.PlainInit(dir, false); err != nil {
		t.Fatalf("PlainInit: %v", err)
	}

	filePath := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(filePath, []byte("diskonly"), 0o600); err != nil {
		t.Fatal(err)
	}

	old, newText, _, _, err := resolveGitWorkingTreeVsHEAD(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if old != "" || newText != "diskonly" {
		t.Fatalf("expected empty old and disk content new; old=%q new=%q", old, newText)
	}
}
