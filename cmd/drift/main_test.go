package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCLI_identicalFromTo(t *testing.T) {
	t.Helper()
	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	code := runCLI(streams, []string{"--from", "same", "--to", "same"})
	if code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, errOut.String())
	}
}

func TestRunCLI_differs(t *testing.T) {
	t.Helper()
	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	code := runCLI(streams, []string{"--from", "a", "--to", "b"})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d stderr=%q", code, errOut.String())
	}
	if out.Len() == 0 {
		t.Fatal("expected diff on stdout")
	}
}

func TestRunCLI_gitSingleArg_differs(t *testing.T) {
	t.Helper()
	bin := t.TempDir()
	repo := t.TempDir()
	file := filepath.Join(repo, "x.txt")
	if err := os.WriteFile(file, []byte("working\n"), 0o600); err != nil {
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
		"  *show*HEAD:x.txt*) printf '%s' 'head\n'; exit 0 ;;\n" +
		"esac\n" +
		"echo \"fake git: $joined\" >&2\n" +
		"exit 99\n"
	writeFakeGit(t, bin, script)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	code := runCLI(streams, []string{file})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d stderr=%q stdout=%q", code, errOut.String(), out.String())
	}
	if out.Len() == 0 {
		t.Fatal("expected diff on stdout")
	}
}

func TestRunCLI_invalidAlgorithm(t *testing.T) {
	t.Helper()
	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	code := runCLI(streams, []string{"--algorithm", "nope", "--from", "a", "--to", "b"})
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "algorithm") {
		t.Fatalf("stderr should mention algorithm: %q", errOut.String())
	}
}

func TestHelpListsAllFlags(t *testing.T) {
	t.Helper()
	buf := new(bytes.Buffer)
	streams := IOStreams{In: strings.NewReader(""), Out: buf, Err: buf}
	code := runCLI(streams, []string{"--help"})
	// --help exits 0
	if code != 0 {
		t.Fatalf("expected exit 0 from --help, got %d", code)
	}
	out := buf.String()
	for _, flag := range []string{
		"--split",
		"--no-line-numbers",
		"--algorithm",
		"--lang",
		"--theme",
		"--no-color",
		"--context",
		"--from",
		"--to",
		"--no-pager",
		"git",
	} {
		if !strings.Contains(out, flag) {
			t.Errorf("help output missing flag %q\noutput:\n%s", flag, out)
		}
	}
}

func TestRunCLI_noPagerFlag(t *testing.T) {
	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	code := runCLI(streams, []string{"--from", "a\n", "--to", "b\n", "--no-pager"})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d stderr=%q", code, errOut.String())
	}
	if !strings.Contains(out.String(), "@@") {
		t.Fatalf("expected diff hunk header in output: %q", out.String())
	}
}

func TestRunCLI_pagerSkippedOnNonTTY(t *testing.T) {
	// Non-TTY out: shouldPage must return false → output written directly to buffer
	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	code := runCLI(streams, []string{"--from", "a\n", "--to", "b\n"})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d stderr=%q", code, errOut.String())
	}
	if out.Len() == 0 {
		t.Fatal("expected diff output in buffer (pager should not have consumed it)")
	}
}
