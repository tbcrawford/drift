package main

import (
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
	"github.com/tbcrawford/drift"
	"github.com/tbcrawford/drift/internal/chrome"
	"github.com/tbcrawford/drift/internal/highlight"
	"github.com/tbcrawford/drift/internal/terminal"
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
	noPager       bool
	colorOnly     bool
	chrome        string
}

// rootOptions holds fully resolved values ready for execution.
// All decisions (algorithm parsing, option building, streams assignment)
// are made in resolveRootOptions — nothing is deferred to run time.
type rootOptions struct {
	streams      IOStreams
	driftOpts    []drift.Option
	from         string
	to           string
	args         []string
	showTheme    bool // retained for show-theme stderr callback wiring
	noPager      bool
	noColor      bool               // mirrors --no-color; used for plain-text header rendering
	termWidth    int                // terminal width for header rule; 0 means use default (80)
	colorOnly    bool               // mirrors --color-only; pager mode pass-through with ANSI coloring
	chromeTheme  chrome.ChromeTheme // resolved chrome decoration theme (default: DriftTheme)
	contextLines int                // resolved --context value; used for git diff backfill
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

	// Measure terminal width, detect color profile, and probe the terminal
	// background color — all from the real output stream now, before any
	// bytes.Buffer is involved. This ensures:
	//   - split-view panels fill the actual terminal width (not the 80-column default)
	//   - ANSI colors are preserved when output is buffered for paging
	//   - The OSC 11 background-color query fires exactly once (not once per file
	//     in a directory diff), preventing concurrent terminal queries that can
	//     cause raw escape sequences to appear in stdout
	// WithTermWidth, WithColorProfile, and WithIsDark short-circuit the per-call
	// probes in buildRenderPipeline, so no internal API changes are needed.
	var resolvedTermWidth int
	if f, ok := streams.Out.(*os.File); ok {
		if w, _, err := term.GetSize(f.Fd()); err == nil && w > 0 {
			resolvedTermWidth = w
			opts = append(opts, drift.WithTermWidth(w))
		}
		profile := colorprofile.Detect(f, os.Environ())
		opts = append(opts, drift.WithColorProfile(profile))

		// Probe the terminal background color once, before any goroutines start.
		// Only query the terminal when colors are enabled and the output is a TTY —
		// HasDarkBackground sends OSC 11 + DA2 to the terminal and reads the
		// response. Doing this inside parallel goroutines causes concurrent writes
		// to os.Stdout (via backgroundColor(os.Stdout, os.Stdout)) which races and
		// leaks raw escape-sequence responses into the diff output.
		if !flags.noColor && term.IsTerminal(f.Fd()) {
			isDark := lipgloss.HasDarkBackground(os.Stdin, f)
			opts = append(opts, drift.WithIsDark(isDark))

			// Resolve the Chroma theme once via OSC 4 palette query, before any
			// goroutines start. When rendering a directory diff, each file is
			// rendered in a parallel goroutine; if resolveChromaStyle calls
			// QueryANSIPalette inside each goroutine, N concurrent calls race on
			// /dev/tty (shared kernel tty buffer), causing OSC 4 responses to
			// interleave and leak as raw escape sequences into the diff output.
			//
			// Only do this when the caller hasn't explicitly set --theme (which
			// already short-circuits the OSC 4 branch in resolveChromaStyle).
			if flags.theme == "" && profile != colorprofile.NoTTY && profile != colorprofile.Ascii {
				if palette, err := terminal.QueryANSIPalette(); err == nil && len(palette) > 0 {
					opts = append(opts, drift.WithTheme(highlight.BestMatchTheme(palette)))
				}
			}
		}
	}

	// Resolve chrome decoration theme from --chrome flag value.
	// Empty string → DriftTheme (default). Unknown names return an error.
	chromeTheme, err := chrome.ParseChromeTheme(flags.chrome)
	if err != nil {
		return nil, newExitCode(2, err.Error())
	}
	opts = append(opts, drift.WithHunkHeaderRenderer(chromeTheme.RenderHunkHeader))

	return &rootOptions{
		streams:      streams,
		driftOpts:    opts,
		from:         flags.from,
		to:           flags.to,
		args:         args,
		showTheme:    flags.showTheme,
		noPager:      flags.noPager,
		noColor:      flags.noColor,
		termWidth:    resolvedTermWidth,
		colorOnly:    flags.colorOnly,
		chromeTheme:  chromeTheme,
		contextLines: flags.context,
	}, nil
}
