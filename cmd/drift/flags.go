package main

import (
	"fmt"

	"github.com/tbcrawford/drift"
)

// rootFlags holds the raw values parsed by cobra from the command line.
// Every flag defined on rootCmd maps to exactly one field here.
// No resolution logic lives here — fields are populated by cobra directly.
type rootFlags struct {
	split         bool
	noLineNumbers bool
	algorithm     string
	lang          string
	theme         string
	noColor       bool
	context       int
	from          string
	to            string
	showTheme     bool
}

// rootOptions holds fully resolved values ready for execution.
// All decisions (algorithm parsing, option building, streams assignment)
// are made in resolveRootOptions — nothing is deferred to run time.
type rootOptions struct {
	streams   IOStreams
	driftOpts []drift.Option
	from      string
	to        string
	args      []string
	showTheme bool // retained for show-theme stderr callback wiring
}

// resolveRootOptions converts raw cobra flags into a fully populated rootOptions.
// It returns an error for invalid flag values (e.g. unknown algorithm or negative context).
// All I/O decisions are made here; runRoot() orchestrates work only.
func resolveRootOptions(flags *rootFlags, streams IOStreams, args []string) (*rootOptions, error) {
	var opts []drift.Option

	a, err := parseAlgorithm(flags.algorithm)
	if err != nil {
		return nil, err
	}
	opts = append(opts, drift.WithAlgorithm(a))

	if flags.context < 0 {
		return nil, newExitCode(2, "invalid context: must be non-negative")
	}
	opts = append(opts, drift.WithContext(flags.context))

	if flags.noColor {
		opts = append(opts, drift.WithNoColor())
	}
	if flags.lang != "" {
		opts = append(opts, drift.WithLang(flags.lang))
	}
	if flags.theme != "" {
		opts = append(opts, drift.WithTheme(flags.theme))
	}
	if flags.showTheme {
		opts = append(opts, drift.WithThemeResolved(func(name string) {
			fmt.Fprintf(streams.Err, "drift: resolved syntax theme: %s\n", name)
		}))
	}
	if flags.split {
		opts = append(opts, drift.WithSplit())
	}
	if flags.noLineNumbers {
		opts = append(opts, drift.WithoutLineNumbers())
	}

	return &rootOptions{
		streams:   streams,
		driftOpts: opts,
		from:      flags.from,
		to:        flags.to,
		args:      args,
		showTheme: flags.showTheme,
	}, nil
}
