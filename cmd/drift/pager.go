package main

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/x/term"
)

// resolvePager returns the pager command to use.
// Resolution order: $PAGER env → "less -R" (if less on PATH) → "more".
func resolvePager() string {
	if p := os.Getenv("PAGER"); p != "" {
		return p
	}
	if _, err := exec.LookPath("less"); err == nil {
		return "less -R"
	}
	return "more"
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
	cleanup := func() {
		_ = pw.Close()
		_ = pr.Close()
		_ = cmd.Wait()
	}
	return pw, cleanup, nil
}
