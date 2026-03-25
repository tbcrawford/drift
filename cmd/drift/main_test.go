package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunCLI_identicalFromTo(t *testing.T) {
	t.Helper()
	var out, errOut bytes.Buffer
	code := runCLI(&out, &errOut, strings.NewReader(""), []string{"--from", "same", "--to", "same"})
	if code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, errOut.String())
	}
}

func TestRunCLI_differs(t *testing.T) {
	t.Helper()
	var out, errOut bytes.Buffer
	code := runCLI(&out, &errOut, strings.NewReader(""), []string{"--from", "a", "--to", "b"})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d stderr=%q", code, errOut.String())
	}
	if out.Len() == 0 {
		t.Fatal("expected diff on stdout")
	}
}

func TestRunCLI_invalidAlgorithm(t *testing.T) {
	t.Helper()
	var out, errOut bytes.Buffer
	code := runCLI(&out, &errOut, strings.NewReader(""), []string{"--algorithm", "nope", "--from", "a", "--to", "b"})
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
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, flag := range []string{
		"--split",
		"--algorithm",
		"--lang",
		"--theme",
		"--no-color",
		"--context",
		"--from",
		"--to",
	} {
		if !strings.Contains(out, flag) {
			t.Errorf("help output missing flag %q\noutput:\n%s", flag, out)
		}
	}
}
