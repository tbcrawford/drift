package main

import (
	"io"
	"os"
)

// IOStreams holds the standard I/O channels for a CLI invocation.
// It is constructed once in main() and injected everywhere; no code below
// main() accesses os.Stdout, os.Stdin, or os.Stderr directly.
type IOStreams struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

// System returns an IOStreams wired to the real OS I/O channels.
// Call once in main() or executeDrift().
func System() IOStreams {
	return IOStreams{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}
}
