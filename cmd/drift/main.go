package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"github.com/tbcrawford/drift"
	"golang.org/x/sync/errgroup"
)

// parseAlgorithm maps a flag string to a drift.Algorithm value.
// Preserved here because flags.go calls it during resolveRootOptions.
func parseAlgorithm(s string) (drift.Algorithm, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "auto":
		return drift.Auto, nil
	case "myers":
		return drift.Myers, nil
	case "patience":
		return drift.Patience, nil
	case "histogram":
		return drift.Histogram, nil
	default:
		return 0, newExitCode(2, fmt.Sprintf("invalid algorithm: %q (use auto, myers, patience, histogram)", s))
	}
}

// isStdinPipe reports whether streams.In is a real pipe (not a TTY).
// Used to detect when git is piping its diff output into drift.
// For non-file readers (bytes.Buffer in tests), always returns true —
// tests inject piped input without needing a real OS pipe.
// Readers that implement isTTYReader (returning true from IsTTY()) are treated
// as TTY inputs — this allows tests to simulate a TTY stdin for zero-arg git mode.
func isStdinPipe(in io.Reader) bool {
	// Allow test helpers to mark a reader as TTY (not a pipe).
	type isTTYReader interface {
		IsTTY() bool
	}
	if t, ok := in.(isTTYReader); ok {
		return !t.IsTTY()
	}
	f, ok := in.(*os.File)
	if !ok {
		return true // non-file readers (bytes.Buffer in tests) are always "piped"
	}
	return !term.IsTerminal(f.Fd())
}

// colorizeUnifiedLine applies simple ANSI green/red to +/- diff lines.
// Context lines (space prefix) and headers (@@, diff, index) pass through unchanged.
func colorizeUnifiedLine(line string, noColor bool) string {
	if noColor || len(line) == 0 {
		return line
	}
	switch line[0] {
	case '+':
		return "\x1b[32m" + line + "\x1b[0m" // green
	case '-':
		return "\x1b[31m" + line + "\x1b[0m" // red
	default:
		return line
	}
}

// runColorOnlyMode reads unified diff lines from r and writes them to w with
// ANSI color applied to +/- lines. Line structure is preserved exactly so that
// git add -p can process the output (interactive.diffFilter requirement).
func runColorOnlyMode(r io.Reader, w io.Writer, opts *rootOptions) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintln(w, colorizeUnifiedLine(line, opts.noColor))
	}
	if err := scanner.Err(); err != nil {
		return newExitCode(2, err.Error())
	}
	return nil
}

// runPagerMode reads a multi-file unified diff from r, re-renders each file
// using drift's full pipeline, and writes the result to the pager (or stdout).
// This is invoked when drift is used as git's core.pager.
func runPagerMode(r io.Reader, opts *rootOptions) error {
	files, err := parseUnifiedDiff(r)
	if err != nil {
		return newExitCode(2, fmt.Sprintf("drift: failed to parse diff input: %v", err))
	}
	if len(files) == 0 {
		return nil // no diff → exit 0
	}
	return streamThroughPager(opts, func(w io.Writer) (bool, error) {
		hasDiff := false
		for _, f := range files {
			if f.IsBinary {
				continue // skip binary files
			}
			result, dErr := drift.Diff(f.OldContent, f.NewContent, opts.driftOpts...)
			if dErr != nil {
				return false, newExitCode(2, dErr.Error())
			}
			if result.IsEqual {
				continue
			}
			var buf bytes.Buffer
			writeFileHeader(&buf, f.Name, opts.noColor, opts.termWidth)
			if rErr := drift.RenderWithNames(result, &buf, f.OldName, f.NewName, opts.driftOpts...); rErr != nil {
				return false, newExitCode(2, rErr.Error())
			}
			_, _ = buf.WriteTo(w)
			hasDiff = true
		}
		return hasDiff, nil
	})
}

// newRootCmd constructs the root cobra command for a single CLI invocation.
// All state is local — no package-level variables are shared across invocations.
func newRootCmd(streams IOStreams) *cobra.Command {
	flags := &rootFlags{}

	cmd := &cobra.Command{
		Use:   "drift [flags] OLD NEW | FILE",
		Short: "Pretty-print a diff between two inputs",
		Long: `Pretty-print a diff between two inputs.
With one path inside a git repository, diffs the working tree against HEAD.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := resolveRootOptions(flags, streams, args)
			if err != nil {
				return err
			}
			return runRoot(opts)
		},
	}

	cmd.Flags().BoolVar(&flags.split, "split", false, "side-by-side split view")
	cmd.Flags().BoolVar(&flags.noLineNumbers, "no-line-numbers", false, "hide old/new line-number gutters")
	cmd.Flags().StringVar(&flags.algorithm, "algorithm", "auto", "diff algorithm: auto, myers, patience, histogram")
	cmd.Flags().StringVar(&flags.lang, "lang", "", "Chroma language override (e.g. go, python)")
	cmd.Flags().StringVar(&flags.theme, "theme", "", "Chroma style/theme override")
	cmd.Flags().BoolVar(&flags.noColor, "no-color", false, "disable ANSI colors")
	cmd.Flags().IntVar(&flags.context, "context", 3, "lines of context around hunks")
	cmd.Flags().StringVar(&flags.from, "from", "", "old text as a raw string (use with --to)")
	cmd.Flags().StringVar(&flags.to, "to", "", "new text as a raw string (use with --from)")
	cmd.Flags().BoolVar(&flags.showTheme, "show-theme", false, "print resolved Chroma theme to stderr after selection")
	_ = cmd.Flags().MarkHidden("show-theme")
	cmd.Flags().BoolVar(&flags.noPager, "no-pager", false, "disable automatic pager for large output")
	cmd.Flags().BoolVar(&flags.colorOnly, "color-only", false, "re-color stdin diff without restructuring (for interactive.diffFilter)")

	installPagerCmd := &cobra.Command{
		Use:   "install-pager",
		Short: "Print a ~/.gitconfig snippet to use drift as git's pager",
		RunE: func(cmd *cobra.Command, args []string) error {
			snippet := `# Add this to ~/.gitconfig to use drift as a drop-in for delta:
[core]
    pager = drift

[interactive]
    diffFilter = drift --color-only
`
			fmt.Fprint(streams.Out, snippet)
			return nil
		},
	}
	cmd.AddCommand(installPagerCmd)

	return cmd
}

// fileHeaderName returns the display name to use in the file header for the
// single-file and two-file diff paths. Returns "" when no meaningful file name
// is available (--from/--to flags, stdin "-", or no args).
func fileHeaderName(args []string) string {
	switch len(args) {
	case 1:
		if args[0] == "-" {
			return ""
		}
		return args[0]
	case 2:
		a, b := args[0], args[1]
		if a == "-" && b == "-" {
			return ""
		}
		if a == "-" {
			return b
		}
		if b == "-" {
			return a
		}
		if a == b {
			return a
		}
		return a + " → " + b
	default:
		return ""
	}
}

// writeFileHeader writes a styled file header into buf before each file's diff output.
//
// Styled (color) format:
//
//	▸ filename
//	────────────────────────────────────────────────────────────
//
// Plain (--no-color / NoTTY) format:
//
//	▸ filename
//	------------------------------------------------------------
//
// A blank line follows the rule so the diff hunk below has breathing room.
func writeFileHeader(buf *bytes.Buffer, name string, noColor bool, termWidth int) {
	const fallbackWidth = 80
	width := termWidth
	if width <= 0 {
		width = fallbackWidth
	}

	if noColor {
		buf.WriteString("▸ " + name + "\n")
		buf.WriteString(strings.Repeat("-", width) + "\n")
		buf.WriteString("\n")
		return
	}

	// Accent color for the ▸ glyph — a muted slate-blue (ANSI 256 #63).
	chevronStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Bold(true)

	// Filename in a muted foreground (bright white on dark / dark gray on light).
	nameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250"))

	// Rule in a dimmer tone so it recedes behind the filename.
	ruleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("238"))

	rule := strings.Repeat("─", width)

	buf.WriteString(chevronStyle.Render("▸") + " " + nameStyle.Render(name) + "\n")
	buf.WriteString(ruleStyle.Render(rule) + "\n")
	buf.WriteString("\n")
}

// runDirectoryDiff iterates over file pairs produced by diffDirectories, renders
// each changed/added/removed file's diff into w preceded by a styled file header.
// Diffs are computed in parallel (one goroutine per file) and merged in sorted order.
// Returns hasDiff=true if any file pair produced output.
func runDirectoryDiff(pairs []filePair, opts *rootOptions, w io.Writer) (hasDiff bool, err error) {
	if len(pairs) == 0 {
		return false, nil
	}

	// Render each file's diff into its own buffer in parallel, preserving pair order.
	results := make([]bytes.Buffer, len(pairs))
	g := new(errgroup.Group)

	for i, pair := range pairs {
		i, pair := i, pair // capture loop vars
		g.Go(func() error {
			oldName := "a/" + pair.Name
			newName := "b/" + pair.Name

			result, dErr := drift.Diff(pair.OldContent, pair.NewContent, opts.driftOpts...)
			if dErr != nil {
				return newExitCode(2, dErr.Error())
			}
			if result.IsEqual {
				return nil // skip (e.g. CRLF-only difference normalised away)
			}

			writeFileHeader(&results[i], pair.Name, opts.noColor, opts.termWidth)
			return drift.RenderWithNames(result, &results[i], oldName, newName, opts.driftOpts...)
		})
	}

	if err := g.Wait(); err != nil {
		return false, err
	}

	for i := range results {
		if results[i].Len() > 0 {
			hasDiff = true
			_, _ = results[i].WriteTo(w)
		}
	}
	return hasDiff, nil
}

// runGitDirectoryDiff renders git HEAD-vs-working-tree diffs for all changed
// files under a single directory into w, with styled file headers.
// Diffs are computed in parallel (one goroutine per file) and merged in sorted order.
func runGitDirectoryDiff(pairs []gitFilePair, opts *rootOptions, w io.Writer) (hasDiff bool, err error) {
	if len(pairs) == 0 {
		return false, nil
	}

	// Render each file's diff into its own buffer in parallel, preserving pair order.
	results := make([]bytes.Buffer, len(pairs))
	g := new(errgroup.Group)

	for i, pair := range pairs {
		i, pair := i, pair // capture loop vars
		g.Go(func() error {
			result, dErr := drift.Diff(pair.HeadContent, pair.WorkContent, opts.driftOpts...)
			if dErr != nil {
				return newExitCode(2, dErr.Error())
			}
			if result.IsEqual {
				return nil
			}
			writeFileHeader(&results[i], pair.Name, opts.noColor, opts.termWidth)
			return drift.RenderWithNames(result, &results[i], pair.OldName, pair.NewName, opts.driftOpts...)
		})
	}

	if err := g.Wait(); err != nil {
		return false, err
	}

	for i := range results {
		if results[i].Len() > 0 {
			hasDiff = true
			_, _ = results[i].WriteTo(w)
		}
	}
	return hasDiff, nil
}

// streamThroughPager starts the pager immediately (if stdout is a TTY and paging is
// enabled), then calls renderFn which writes rendered diff content directly to the
// pager/stdout writer. This eliminates the full-buffer-then-page pattern, allowing
// the first file's diff to appear in the terminal as soon as it's computed.
//
// If the pager cannot be started, falls back to writing directly to opts.streams.Out.
// renderFn returns (hasDiff, error): hasDiff=false means no output, so exit 0.
//
// When drift is invoked as git's core.pager, git sets GIT_PAGER_IN_USE in the
// environment. In that case we must NOT start a nested pager — git already owns
// the terminal session and has configured its own pager environment (e.g. LESS=FRX).
// Launching another less inside that context produces the chain git→drift→less,
// where the nested less inherits -F (quit-if-one-screen) and silently exits without
// displaying anything. Instead, write rendered output directly to stdout (the TTY).
func streamThroughPager(opts *rootOptions, renderFn func(w io.Writer) (hasDiff bool, err error)) error {
	_, gitPagerInUse := os.LookupEnv("GIT_PAGER_IN_USE")
	if f, ok := opts.streams.Out.(*os.File); ok && !opts.noPager && !gitPagerInUse {
		// TTY path: start pager before any rendering so output streams immediately.
		pagerWriter, cleanup, pErr := startPager(resolvePager(), opts.streams)
		if pErr == nil {
			hasDiff, renderErr := renderFn(pagerWriter)
			cleanup()
			if renderErr != nil && !isPipeClosedErr(renderErr) {
				return newExitCode(2, renderErr.Error())
			}
			if !hasDiff {
				return nil
			}
			return newExitCode(1, "")
		}
		// Pager failed to start — fall through to direct write.
		_ = f
	}
	// Non-TTY, git-pager context, or pager fallback: write directly to stdout.
	hasDiff, err := renderFn(opts.streams.Out)
	if err != nil {
		return newExitCode(2, err.Error())
	}
	if !hasDiff {
		return nil
	}
	return newExitCode(1, "")
}

// writeThroughPager writes buf to a pager (or directly to opts.streams.Out if paging
// is not needed). It always returns newExitCode(1, "") so callers can do:
//
//	return writeThroughPager(&buf, opts)
//
// If the user quits the pager early (e.g. presses q in less), the write may
// return io.ErrClosedPipe or syscall.EPIPE. Both are treated as a clean exit:
// the user intentionally dismissed the output, which is not an error condition.
//
// When drift is invoked as git's core.pager (GIT_PAGER_IN_USE is set), we skip
// launching a nested pager and write directly to stdout — git already owns the
// terminal session.
func writeThroughPager(buf *bytes.Buffer, opts *rootOptions) error {
	_, gitPagerInUse := os.LookupEnv("GIT_PAGER_IN_USE")
	var termHeight int
	if f, ok := opts.streams.Out.(*os.File); ok {
		if _, h, tErr := term.GetSize(f.Fd()); tErr == nil {
			termHeight = h
		}
	}
	lineCount := strings.Count(buf.String(), "\n")
	if !gitPagerInUse && shouldPage(opts.streams.Out, lineCount, termHeight, opts.noPager) {
		pagerWriter, cleanup, pErr := startPager(resolvePager(), opts.streams)
		if pErr != nil {
			_, _ = buf.WriteTo(opts.streams.Out)
		} else {
			_, writeErr := buf.WriteTo(pagerWriter)
			cleanup()
			// io.ErrClosedPipe means our background goroutine closed pr because
			// the pager exited (user pressed q). syscall.EPIPE is the OS-level
			// equivalent on some platforms. Both mean "user quit intentionally" —
			// treat as success.
			if writeErr != nil && !isPipeClosedErr(writeErr) {
				return newExitCode(2, writeErr.Error())
			}
		}
	} else {
		_, _ = buf.WriteTo(opts.streams.Out)
	}
	return newExitCode(1, "")
}

// isPipeClosedErr reports whether err indicates that the write end of a pipe
// was closed because the reader (pager process) exited. This is the normal
// outcome when the user quits the pager before all output is consumed.
func isPipeClosedErr(err error) bool {
	if errors.Is(err, io.ErrClosedPipe) {
		return true
	}
	// syscall.EPIPE is returned on some platforms/Go versions when writing to
	// a pipe whose read end has been closed by the OS.
	var errno syscall.Errno
	if errors.As(err, &errno) && errno == syscall.EPIPE {
		return true
	}
	return false
}

// runRoot is a thin orchestrator: it calls resolveInputs, drift.Diff, and drift.RenderWithNames
// in sequence. No flag parsing or I/O decisions live here.
func runRoot(opts *rootOptions) error {
	// interactive.diffFilter mode: git add -p pipes a colorized diff to drift.
	// --color-only means: highlight each +/- line with ANSI color but preserve
	// the exact unified diff line structure (git add -p requires 1:1 line correspondence).
	if opts.colorOnly && isStdinPipe(opts.streams.In) {
		return runColorOnlyMode(opts.streams.In, opts.streams.Out, opts)
	}

	// Git pager mode: when stdin is a pipe and no positional args + no --from/--to,
	// treat stdin as a unified diff from git and re-render it with drift styling.
	// This enables: core.pager = drift in ~/.gitconfig.
	// This MUST come before the zero-arg git worktree block, which only fires when stdin is a TTY.
	if len(opts.args) == 0 && opts.from == "" && opts.to == "" && isStdinPipe(opts.streams.In) {
		return runPagerMode(opts.streams.In, opts)
	}

	// Zero-argument mode: no args, no --from/--to → diff the entire repo working tree vs HEAD.
	if len(opts.args) == 0 && opts.from == "" && opts.to == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return newExitCode(2, fmt.Sprintf("fatal: could not determine working directory: %v", err))
		}
		pairs, err := gitDirectoryVsHEAD(cwd)
		if err != nil {
			// Distinguish "not a git repo" from other errors.
			if errors.Is(err, git.ErrRepositoryNotExists) {
				fmt.Fprintln(opts.streams.Err, "fatal: not a git repository (or any parent up to root)")
				return newExitCode(2, "")
			}
			return newExitCode(2, err.Error())
		}
		return streamThroughPager(opts, func(w io.Writer) (bool, error) {
			return runGitDirectoryDiff(pairs, opts, w)
		})
	}

	// Single-directory git diff: one arg that is a directory → diff vs HEAD.
	if len(opts.args) == 1 && isDir(opts.args[0]) {
		absDir, err := filepath.Abs(opts.args[0])
		if err != nil {
			return newExitCode(2, fmt.Sprintf("invalid path %q: %v", opts.args[0], err))
		}
		pairs, err := gitDirectoryVsHEAD(absDir)
		if err != nil {
			return newExitCode(2, err.Error())
		}
		return streamThroughPager(opts, func(w io.Writer) (bool, error) {
			return runGitDirectoryDiff(pairs, opts, w)
		})
	}

	// Two-directory diff: both positional args are directories.
	if len(opts.args) == 2 && isDir(opts.args[0]) && isDir(opts.args[1]) {
		pairs, err := diffDirectories(opts.args[0], opts.args[1])
		if err != nil {
			return newExitCode(2, err.Error())
		}
		return streamThroughPager(opts, func(w io.Writer) (bool, error) {
			return runDirectoryDiff(pairs, opts, w)
		})
	}

	old, newText, oldName, newName, err := resolveInputs(opts.args, opts.from, opts.to, opts.streams.In)
	if err != nil {
		return err
	}

	result, err := drift.Diff(old, newText, opts.driftOpts...)
	if err != nil {
		return newExitCode(2, err.Error())
	}

	if result.IsEqual {
		return nil
	}

	// Derive a display name for the file header, when a real file path is available.
	// Skip when input comes from --from/--to flags or stdin ("-").
	headerName := fileHeaderName(opts.args)

	// Render to a buffer so we can count lines and decide whether to page.
	var buf bytes.Buffer
	if headerName != "" {
		writeFileHeader(&buf, headerName, opts.noColor, opts.termWidth)
	}
	if err := drift.RenderWithNames(result, &buf, oldName, newName, opts.driftOpts...); err != nil {
		return newExitCode(2, err.Error())
	}

	return writeThroughPager(&buf, opts)
}

// runCLI creates a fresh cobra command tree for each invocation, executes it,
// and returns an exit code. Injecting IOStreams makes every call independently testable.
func runCLI(streams IOStreams, args []string) int {
	cmd := newRootCmd(streams)
	cmd.SetIn(streams.In)
	cmd.SetOut(streams.Out)
	cmd.SetErr(streams.Err)
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err == nil {
		return 0
	}

	var ec *exitCodeErr
	if errors.As(err, &ec) {
		if ec.msg != "" {
			fmt.Fprintln(streams.Err, ec.msg)
		}
		return ec.code
	}

	fmt.Fprintln(streams.Err, err)
	return 2
}

// executeDrift wires real OS I/O and runs the CLI. Called from main() and testscript.
func executeDrift() int {
	return runCLI(System(), os.Args[1:])
}

func main() {
	os.Exit(executeDrift())
}
