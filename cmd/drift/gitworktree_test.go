package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
