package main

import (
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

// --- resolveGitWorkingTreeVsHEAD tests ---

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
