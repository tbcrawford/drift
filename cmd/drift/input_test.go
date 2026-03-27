package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveInputs_singleArg_gitMode(t *testing.T) {
	t.Helper()
	bin := t.TempDir()
	repo := t.TempDir()
	file := filepath.Join(repo, "f.go")
	if err := os.WriteFile(file, []byte("new\n"), 0o600); err != nil {
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
		"  *cat-file*-e*HEAD:f.go*) exit 0 ;;\n" +
		"  *show*HEAD:f.go*) printf '%s' 'old\n'; exit 0 ;;\n" +
		"esac\n" +
		"echo \"fake git: $joined\" >&2\n" +
		"exit 99\n"
	writeFakeGit(t, bin, script)
	prependPath(t, bin)

	old, new_, on, nn, err := resolveInputs([]string{file}, "", "", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if old != "old\n" || new_ != "new\n" {
		t.Fatalf("content old=%q new=%q", old, new_)
	}
	if !strings.HasSuffix(on, "(HEAD)") || !strings.HasSuffix(nn, "(working tree)") {
		t.Fatalf("names on=%q nn=%q", on, nn)
	}
}

func TestResolveInputs_singleArg_notGitWorktree(t *testing.T) {
	t.Helper()
	bin := t.TempDir()
	repo := t.TempDir()
	file := filepath.Join(repo, "f.go")
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

	_, _, _, _, err := resolveInputs([]string{file}, "", "", strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "two") {
		t.Fatalf("expected 'two' in error: %v", err)
	}
}

func TestResolveInputs_zeroArgs(t *testing.T) {
	t.Helper()
	_, _, _, _, err := resolveInputs([]string{}, "", "", strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "expected") || !strings.Contains(err.Error(), "NEW") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveInputs_twoFiles(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	p1 := filepath.Join(dir, "a.txt")
	p2 := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(p1, []byte("alpha\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p2, []byte("beta\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	old, new_, on, nn, err := resolveInputs([]string{p1, p2}, "", "", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if old != "alpha\n" || new_ != "beta\n" {
		t.Fatalf("content mismatch old=%q new=%q", old, new_)
	}
	if on != "a.txt" || nn != "b.txt" {
		t.Fatalf("names on=%q nn=%q", on, nn)
	}
}

func TestResolveInputs_stdinAndFile(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	p2 := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(p2, []byte("file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	stdin := strings.NewReader("x\n")
	old, new_, on, nn, err := resolveInputs([]string{"-", p2}, "", "", stdin)
	if err != nil {
		t.Fatal(err)
	}
	if old != "x\n" || new_ != "file\n" {
		t.Fatalf("got old=%q new=%q", old, new_)
	}
	if on != "-" || nn != "b.txt" {
		t.Fatalf("names on=%q nn=%q", on, nn)
	}
}

func TestResolveInputs_fromToFlags(t *testing.T) {
	t.Helper()
	old, new_, on, nn, err := resolveInputs([]string{}, "hello", "world", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if old != "hello" || new_ != "world" {
		t.Fatal("flag content")
	}
	if on != "a/string" || nn != "b/string" {
		t.Fatalf("names on=%q nn=%q", on, nn)
	}
}

func TestResolveInputs_stdinStdin(t *testing.T) {
	t.Helper()
	r := strings.NewReader("same\n")
	old, new_, _, _, err := resolveInputs([]string{"-", "-"}, "", "", r)
	if err != nil {
		t.Fatal(err)
	}
	if old != "same\n" || new_ != "same\n" {
		t.Fatalf("expected identical stdin content, old=%q new=%q", old, new_)
	}
}

func TestResolveInputs_onlyFrom(t *testing.T) {
	t.Helper()
	_, _, _, _, err := resolveInputs([]string{}, "only", "", strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid") && !strings.Contains(strings.ToLower(err.Error()), "both") {
		t.Fatalf("unexpected error: %v", err)
	}
}
