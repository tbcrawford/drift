package main

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/x/term"
)

// resolvePager returns the pager command to use, ensuring ANSI color codes are
// preserved. Resolution order: $PAGER env → "less -R" (if less on PATH) → "more".
//
// When the user-configured pager is "less" (or a path ending in /less), -R /
// --RAW-CONTROL-CHARS is added if neither flag is already present. This matches
// the behaviour of delta and git, which both force -R so that ANSI sequences are
// passed through rather than displayed as escape characters.
func resolvePager() string {
	if p := os.Getenv("PAGER"); p != "" {
		return ensureLessColors(p)
	}
	if _, err := exec.LookPath("less"); err == nil {
		return "less -R"
	}
	return "more"
}

// ensureLessColors returns cmd unchanged unless cmd invokes less, in which case
// it appends -R if neither -R nor --RAW-CONTROL-CHARS is already present.
// This preserves any flags the user already set (e.g. PAGER="less -F -X").
func ensureLessColors(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return cmd
	}
	// Check whether the first word is "less" (bare name or absolute/relative path).
	base := parts[0]
	if idx := strings.LastIndex(base, "/"); idx >= 0 {
		base = base[idx+1:]
	}
	if base != "less" {
		return cmd
	}
	// Already has a flag that enables raw ANSI passthrough.
	for _, p := range parts[1:] {
		if p == "-R" || p == "--RAW-CONTROL-CHARS" {
			return cmd
		}
	}
	return cmd + " -R"
}

// shouldPage reports whether the pager should be invoked for the given output.
// Returns true only when: out is a TTY *os.File, noPager is false,
// termHeight > 0, and lineCount > termHeight.
func shouldPage(out io.Writer, lineCount int, termHeight int, noPager bool) bool {
	if noPager || termHeight <= 0 || lineCount <= termHeight {
		return false
	}
	f, ok := out.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(f.Fd())
}

// startPager launches the pager subprocess with its stdin connected to a pipe.
// The caller writes diff output to the returned WriteCloser, then calls the
// cleanup func to flush and wait. Pager stdout and stderr are wired to streams.
//
// Early-exit handling: a background goroutine waits for the pager process to
// exit and then closes the read end of the pipe (pr). This causes any in-flight
// pw.Write() call to return io.ErrClosedPipe instead of blocking forever, which
// happens when the user quits the pager (e.g. presses q in less) before all
// output has been written. cleanup() signals that the caller is done writing,
// then waits for the background goroutine to finish.
func startPager(pagerCmd string, streams IOStreams) (io.WriteCloser, func(), error) {
	parts := strings.Fields(pagerCmd)
	if len(parts) == 0 {
		parts = []string{"more"}
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = streams.Out
	cmd.Stderr = streams.Err
	pr, pw := io.Pipe()
	cmd.Stdin = pr
	if err := cmd.Start(); err != nil {
		_ = pr.Close()
		_ = pw.Close()
		return nil, nil, err
	}

	// done is closed by the background goroutine once cmd.Wait() returns.
	done := make(chan struct{})

	// Background goroutine: wait for the pager to exit (for any reason —
	// normal EOF or user pressing q), then close pr so that any blocked
	// pw.Write() unblocks with io.ErrClosedPipe.
	go func() {
		defer close(done)
		_ = cmd.Wait()
		_ = pr.Close()
	}()

	cleanup := func() {
		// Signal EOF to the pager: close the write end so the pager can drain
		// and exit naturally (if it hasn't already).
		_ = pw.Close()
		// Wait for the background goroutine (cmd.Wait + pr.Close) to finish.
		<-done
	}
	return pw, cleanup, nil
}
