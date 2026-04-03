package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestPagerResolvePager verifies the $PAGER env var resolution order.
func TestPagerResolvePager(t *testing.T) {
	t.Run("uses $PAGER when set to non-less pager", func(t *testing.T) {
		t.Setenv("PAGER", "bat")
		got := resolvePager()
		if got != "bat" {
			t.Errorf("resolvePager() = %q, want %q", got, "bat")
		}
	})

	t.Run("adds -R when PAGER=less without flags", func(t *testing.T) {
		t.Setenv("PAGER", "less")
		got := resolvePager()
		if got != "less -R" {
			t.Errorf("resolvePager() = %q, want %q", got, "less -R")
		}
	})

	t.Run("preserves -R when PAGER already has it", func(t *testing.T) {
		t.Setenv("PAGER", "less -R")
		got := resolvePager()
		if got != "less -R" {
			t.Errorf("resolvePager() = %q, want %q", got, "less -R")
		}
	})

	t.Run("preserves other less flags and adds -R", func(t *testing.T) {
		t.Setenv("PAGER", "less -F -X")
		got := resolvePager()
		if got != "less -F -X -R" {
			t.Errorf("resolvePager() = %q, want %q", got, "less -F -X -R")
		}
	})

	t.Run("preserves --RAW-CONTROL-CHARS when already set", func(t *testing.T) {
		t.Setenv("PAGER", "less --RAW-CONTROL-CHARS")
		got := resolvePager()
		if got != "less --RAW-CONTROL-CHARS" {
			t.Errorf("resolvePager() = %q, want %q", got, "less --RAW-CONTROL-CHARS")
		}
	})

	t.Run("ignores empty $PAGER", func(t *testing.T) {
		t.Setenv("PAGER", "")
		got := resolvePager()
		if got == "" {
			t.Error("resolvePager() returned empty string; expected less -R or more")
		}
	})

	t.Run("returns less -R or more when $PAGER unset", func(t *testing.T) {
		t.Setenv("PAGER", "")
		got := resolvePager()
		if got != "less -R" && got != "more" {
			t.Errorf("resolvePager() = %q, want %q or %q", got, "less -R", "more")
		}
	})
}

// TestEnsureLessColors verifies the -R injection logic for less.
func TestEnsureLessColors(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"less", "less -R"},
		{"less -R", "less -R"},
		{"less --RAW-CONTROL-CHARS", "less --RAW-CONTROL-CHARS"},
		{"less -F -X", "less -F -X -R"},
		{"less -F -R -X", "less -F -R -X"},
		{"/usr/bin/less", "/usr/bin/less -R"},
		{"/usr/bin/less -R", "/usr/bin/less -R"},
		{"bat", "bat"},
		{"bat --style=plain", "bat --style=plain"},
		{"more", "more"},
		{"delta", "delta"},
	}
	for _, tc := range tests {
		got := ensureLessColors(tc.input)
		if got != tc.want {
			t.Errorf("ensureLessColors(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// TestPagerShouldPage verifies the short-circuit conditions in shouldPage.
func TestPagerShouldPage(t *testing.T) {
	t.Run("returns false when noPager is true", func(t *testing.T) {
		if shouldPage(os.Stdout, 1000, 24, true) {
			t.Error("shouldPage() = true, want false when noPager=true")
		}
	})

	t.Run("returns false for non-file writer", func(t *testing.T) {
		buf := &bytes.Buffer{}
		if shouldPage(buf, 1000, 24, false) {
			t.Error("shouldPage() = true, want false for non-*os.File writer")
		}
	})

	t.Run("returns false when lineCount <= termHeight", func(t *testing.T) {
		if shouldPage(os.Stdout, 10, 24, false) {
			t.Error("shouldPage() = true, want false when lineCount <= termHeight")
		}
	})

	t.Run("returns false when termHeight is zero", func(t *testing.T) {
		if shouldPage(os.Stdout, 1000, 0, false) {
			t.Error("shouldPage() = true, want false when termHeight is 0")
		}
	})

	t.Run("returns false for *os.File that is not a TTY", func(t *testing.T) {
		// In tests, os.Stdout is never a real TTY (piped or redirected).
		// This verifies that shouldPage correctly requires TTY detection.
		if shouldPage(os.Stdout, 1000, 24, false) {
			t.Error("shouldPage() = true for non-TTY os.Stdout, want false")
		}
	})

	t.Run("returns false for temp file (not a TTY)", func(t *testing.T) {
		f, err := os.CreateTemp(t.TempDir(), "pager-test-*")
		if err != nil {
			t.Fatalf("CreateTemp: %v", err)
		}
		defer f.Close()
		if shouldPage(f, 1000, 24, false) {
			t.Error("shouldPage() = true for temp file (not a TTY), want false")
		}
	})
}

// TestPagerStart verifies that startPager launches a subprocess and wires stdin.
func TestPagerStart(t *testing.T) {
	var buf bytes.Buffer
	streams := IOStreams{
		In:  os.Stdin,
		Out: &buf,
		Err: &buf,
	}

	wc, cleanup, err := startPager("cat", streams)
	if err != nil {
		t.Fatalf("startPager: %v", err)
	}

	payload := []byte("hello from pager test\n")
	if _, err := wc.Write(payload); err != nil {
		t.Fatalf("Write to pager stdin: %v", err)
	}

	cleanup()

	got := buf.String()
	if got != string(payload) {
		t.Errorf("pager output = %q, want %q", got, string(payload))
	}
}

// TestPagerStartEarlyExit verifies that startPager does not deadlock when the
// pager subprocess exits before all input has been written (simulating the user
// pressing q in less). The write to the pipe must unblock promptly and cleanup
// must return within a reasonable timeout.
//
// We use "head -1" as the pager: it reads exactly one line and then exits,
// leaving the writer with unread data in the pipe — exactly the early-exit case.
func TestPagerStartEarlyExit(t *testing.T) {
	var buf bytes.Buffer
	streams := IOStreams{
		In:  os.Stdin,
		Out: &buf,
		Err: os.Stderr,
	}

	wc, cleanup, err := startPager("head -1", streams)
	if err != nil {
		t.Skipf("head not available: %v", err)
	}

	// Write more data than head -1 will consume. The first line is read; after
	// that, head exits and closes its stdin, which triggers the early-exit path.
	largePayload := strings.Repeat("x", 1024) + "\n" + strings.Repeat("y", 1024) + "\n"

	done := make(chan struct{})
	go func() {
		defer close(done)
		// Write may return io.ErrClosedPipe (or succeed partially) — both are fine.
		_, _ = wc.Write([]byte(largePayload))
		cleanup()
	}()

	select {
	case <-done:
		// Good: no deadlock.
	case <-time.After(5 * time.Second):
		t.Fatal("deadlock: startPager early-exit did not unblock within 5 seconds")
	}
}

// TestStreamThroughPager_SkipsNestedPagerWhenGitPagerInUse verifies that
// streamThroughPager does not launch a nested pager subprocess when
// GIT_PAGER_IN_USE is set in the environment.
//
// Regression test for: when drift runs as git's core.pager, git sets
// GIT_PAGER_IN_USE. Without this guard, drift would start a nested less
// subprocess. Git also sets LESS=FRX, so the inner less inherits -F
// (quit-if-one-screen) and silently exits without displaying any output.
//
// We detect pager invocation by using a sentinel PAGER ("sed s/^/PAGED:/")
// that transforms every output line. If the pager is invoked, rendered output
// will contain "PAGED:" prefixes. If GIT_PAGER_IN_USE is respected, drift
// writes directly to stdout and the output is untransformed.
func TestStreamThroughPager_SkipsNestedPagerWhenGitPagerInUse(t *testing.T) {
	if _, err := exec.LookPath("sed"); err != nil {
		t.Skip("sed not available")
	}

	t.Setenv("GIT_PAGER_IN_USE", "1")
	// Sentinel pager: if invoked, adds "PAGED:" to every line.
	// streamThroughPager must NOT invoke this when GIT_PAGER_IN_USE is set.
	t.Setenv("PAGER", "sed s/^/PAGED:/")

	// Use os.Pipe() so Out is *os.File, satisfying the guard that triggers
	// the pager-start branch when GIT_PAGER_IN_USE is absent.
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	defer pr.Close()

	streams := IOStreams{In: strings.NewReader(""), Out: pw, Err: os.Stderr}
	opts := &rootOptions{streams: streams}

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = streamThroughPager(opts, func(w io.Writer) (bool, error) {
			_, _ = io.WriteString(w, "rendered output\n")
			return true, nil
		})
		pw.Close()
	}()

	output, _ := io.ReadAll(pr)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: streamThroughPager did not return within 5s")
	}

	got := string(output)
	if strings.Contains(got, "PAGED:") {
		t.Error("nested pager was invoked despite GIT_PAGER_IN_USE being set; output lines were transformed by sed")
	}
	if !strings.Contains(got, "rendered output") {
		t.Errorf("expected direct output when GIT_PAGER_IN_USE is set, got: %q", got)
	}
}

// TestIsPipeClosedErr verifies the helper correctly identifies pipe-closed errors.
func TestIsPipeClosedErr(t *testing.T) {
	// os.ErrClosed is not a pipe error, so isPipeClosedErr should return false.
	// This is a sanity check that the func doesn't over-match.
	if isPipeClosedErr(os.ErrClosed) {
		t.Error("isPipeClosedErr: os.ErrClosed should not be considered a pipe-closed error")
	}

	pr, pw := func() (*os.File, *os.File) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("os.Pipe: %v", err)
		}
		return r, w
	}()

	// Close the read end, then write to the write end — should get EPIPE or similar.
	_ = pr.Close()
	_, writeErr := pw.Write([]byte("hello"))
	_ = pw.Close()

	if writeErr == nil {
		// On some platforms the write may succeed if the OS buffers it; skip.
		t.Skip("write to closed pipe did not return an error on this platform")
	}

	if !isPipeClosedErr(writeErr) {
		t.Errorf("isPipeClosedErr(%v) = false, want true for OS broken-pipe error", writeErr)
	}
}
