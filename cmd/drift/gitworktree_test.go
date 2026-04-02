package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- filterGitIgnored tests ---

func TestFilterGitIgnored_emptyInput(t *testing.T) {
	// Empty input must return nil without calling git.
	got, err := filterGitIgnored(t.TempDir(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

func TestFilterGitIgnored_noIgnored(t *testing.T) {
	// Fake git that exits 1 with no output → "no paths are ignored" → return all paths unchanged.
	bin := t.TempDir()
	script := "#!/bin/sh\njoined=\"$*\"\ncase \"$joined\" in\n  *check-ignore*) exit 1 ;;\nesac\nexit 99\n"
	writeFakeGit(t, bin, script)
	prependPath(t, bin)

	paths := []string{"main.go", "README.md"}
	got, err := filterGitIgnored(t.TempDir(), paths)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != "main.go" || got[1] != "README.md" {
		t.Fatalf("expected all paths returned; got %v", got)
	}
}

func TestFilterGitIgnored_someIgnored(t *testing.T) {
	// Fake git that prints "dist/app\x00" to stdout → dist/app is ignored.
	bin := t.TempDir()
	script := "#!/bin/sh\njoined=\"$*\"\ncase \"$joined\" in\n  *check-ignore*) printf 'dist/app\\0'; exit 0 ;;\nesac\nexit 99\n"
	writeFakeGit(t, bin, script)
	prependPath(t, bin)

	paths := []string{"main.go", "dist/app", "README.md"}
	got, err := filterGitIgnored(t.TempDir(), paths)
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

func TestFilterGitIgnored_gitNotFound(t *testing.T) {
	// No git in PATH → fail open, return all paths.
	// Use an empty temp dir as PATH so 'git' is not found.
	bin := t.TempDir()
	prependPath(t, bin)
	// Override PATH to be just the empty bin dir (no git).
	t.Setenv("PATH", bin)

	paths := []string{"a.go", "b.go"}
	got, err := filterGitIgnored(t.TempDir(), paths)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected fail-open (all paths); got %v", got)
	}
}

// --- gitDirectoryVsHEAD gitignore integration ---

func TestGitDirectoryVsHEAD_skipsIgnored(t *testing.T) {
	bin := t.TempDir()
	repo := t.TempDir()
	repoAbs, err := filepath.Abs(repo)
	if err != nil {
		t.Fatal(err)
	}

	// Create working-tree files: keep.go (should appear), dist/app (ignored, should not appear).
	if err := os.WriteFile(filepath.Join(repo, "keep.go"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	distDir := filepath.Join(repo, "dist")
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(distDir, "app"), []byte("artifact"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Fake git: rev-parse → inside=true, toplevel=repoAbs
	// cat-file -e → exists (for both), show → "oldcontent", ls-tree → no files (skip deleted detection)
	// check-ignore → marks dist/app as ignored.
	script := "#!/bin/sh\n" +
		"joined=\"$*\"\n" +
		"case \"$joined\" in\n" +
		"  *rev-parse*--is-inside-work-tree*) echo true; exit 0 ;;\n" +
		"  *rev-parse*--show-toplevel*) echo \"" + repoAbs + "\"; exit 0 ;;\n" +
		"  *cat-file*-e*HEAD:keep.go*) exit 0 ;;\n" +
		"  *cat-file*-e*) exit 1 ;;\n" +
		"  *show*HEAD:keep.go*) printf '%s' 'oldcontent'; exit 0 ;;\n" +
		"  *check-ignore*) printf 'dist/app\\0'; exit 0 ;;\n" +
		"  *ls-tree*) echo ''; exit 0 ;;\n" +
		"esac\n" +
		"echo \"fake git: $joined\" >&2\n" +
		"exit 99\n"
	writeFakeGit(t, bin, script)
	prependPath(t, bin)

	pairs, err := gitDirectoryVsHEAD(repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only keep.go should appear — dist/app is ignored.
	for _, p := range pairs {
		if strings.Contains(p.Name, "dist") || strings.Contains(p.Name, "app") {
			t.Errorf("ignored file dist/app appeared in pairs: %+v", p)
		}
	}
	// keep.go should appear (its content differs from HEAD).
	found := false
	for _, p := range pairs {
		if p.Name == "keep.go" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected keep.go in pairs, got: %+v", pairs)
	}
}

func writeFakeGit(t *testing.T, dir, script string) string {
	t.Helper()
	p := filepath.Join(dir, "git")
	if err := os.WriteFile(p, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func prependPath(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestResolveGitWorkingTreeVsHEAD_happyPath(t *testing.T) {
	t.Helper()
	bin := t.TempDir()
	repo := t.TempDir()
	file := filepath.Join(repo, "file.txt")
	if err := os.WriteFile(file, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	repoAbs, err := filepath.Abs(repo)
	if err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\n" +
		"joined=\"$*\"\n" +
		"case \"$joined\" in\n" +
		"  *rev-parse*--is-inside-work-tree*) echo true; exit 0 ;;\n" +
		"  *rev-parse*--show-toplevel*) echo \"" + repoAbs + "\"; exit 0 ;;\n" +
		"  *cat-file*-e*HEAD:file.txt*) exit 0 ;;\n" +
		"  *show*HEAD:file.txt*) printf '%s' 'oldcontent'; exit 0 ;;\n" +
		"esac\n" +
		"echo \"fake git: $joined\" >&2\n" +
		"exit 99\n"
	writeFakeGit(t, bin, script)
	prependPath(t, bin)

	old, new_, on, nn, err := resolveGitWorkingTreeVsHEAD(file)
	if err != nil {
		t.Fatal(err)
	}
	if old != "oldcontent" || new_ != "new" {
		t.Fatalf("content old=%q new=%q", old, new_)
	}
	if !strings.HasSuffix(on, "(HEAD)") || !strings.HasSuffix(nn, "(working tree)") {
		t.Fatalf("names on=%q nn=%q", on, nn)
	}
}

func TestResolveGitWorkingTreeVsHEAD_notInRepo(t *testing.T) {
	t.Helper()
	bin := t.TempDir()
	repo := t.TempDir()
	file := filepath.Join(repo, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\n" +
		"joined=\"$*\"\n" +
		"case \"$joined\" in\n" +
		"  *rev-parse*--is-inside-work-tree*) echo false; exit 0 ;;\n" +
		"esac\n" +
		"exit 99\n"
	writeFakeGit(t, bin, script)
	prependPath(t, bin)

	_, _, _, _, err := resolveGitWorkingTreeVsHEAD(file)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "git") || !strings.Contains(msg, "two") {
		t.Fatalf("error should mention git and two: %v", err)
	}
}

func TestResolveGitWorkingTreeVsHEAD_missingHEADBlob(t *testing.T) {
	t.Helper()
	bin := t.TempDir()
	repo := t.TempDir()
	file := filepath.Join(repo, "file.txt")
	if err := os.WriteFile(file, []byte("diskonly"), 0o600); err != nil {
		t.Fatal(err)
	}
	repoAbs, err := filepath.Abs(repo)
	if err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\n" +
		"joined=\"$*\"\n" +
		"case \"$joined\" in\n" +
		"  *rev-parse*--is-inside-work-tree*) echo true; exit 0 ;;\n" +
		"  *rev-parse*--show-toplevel*) echo \"" + repoAbs + "\"; exit 0 ;;\n" +
		"  *show*HEAD:file.txt*)\n" +
		"    echo \"fatal: path 'file.txt' exists on disk, but not in 'HEAD'\" >&2\n" +
		"    exit 1 ;;\n" +
		"esac\n" +
		"exit 99\n"
	writeFakeGit(t, bin, script)
	prependPath(t, bin)

	old, new_, _, _, err := resolveGitWorkingTreeVsHEAD(file)
	if err != nil {
		t.Fatal(err)
	}
	if old != "" || new_ != "diskonly" {
		t.Fatalf("expected empty old and disk content new; old=%q new=%q", old, new_)
	}
}
