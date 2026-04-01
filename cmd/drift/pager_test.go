package main

import (
	"bytes"
	"os"
	"testing"
)

// TestPagerResolvePager verifies the $PAGER env var resolution order.
func TestPagerResolvePager(t *testing.T) {
	t.Run("uses $PAGER when set", func(t *testing.T) {
		t.Setenv("PAGER", "bat")
		got := resolvePager()
		if got != "bat" {
			t.Errorf("resolvePager() = %q, want %q", got, "bat")
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
