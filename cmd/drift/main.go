package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
	"github.com/tbcrawford/drift"
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
// each changed/added/removed file's diff into buf preceded by a styled file header.
// Returns hasDiff=true if any file pair produced output.
func runDirectoryDiff(pairs []filePair, opts *rootOptions, buf *bytes.Buffer) (hasDiff bool, err error) {
	for _, pair := range pairs {
		var oldContent, newContent string
		oldName := "a/" + pair.Name
		newName := "b/" + pair.Name

		if pair.OldPath != "" {
			b, rErr := os.ReadFile(pair.OldPath)
			if rErr != nil {
				return false, newExitCode(2, fmt.Sprintf("reading %s: %v", pair.OldPath, rErr))
			}
			oldContent = string(b)
		}
		if pair.NewPath != "" {
			b, rErr := os.ReadFile(pair.NewPath)
			if rErr != nil {
				return false, newExitCode(2, fmt.Sprintf("reading %s: %v", pair.NewPath, rErr))
			}
			newContent = string(b)
		}

		result, dErr := drift.Diff(oldContent, newContent, opts.driftOpts...)
		if dErr != nil {
			return false, newExitCode(2, dErr.Error())
		}
		if result.IsEqual {
			continue // skip identical files
		}

		hasDiff = true
		writeFileHeader(buf, pair.Name, opts.noColor, opts.termWidth)
		if rErr := drift.RenderWithNames(result, buf, oldName, newName, opts.driftOpts...); rErr != nil {
			return false, newExitCode(2, rErr.Error())
		}
	}
	return hasDiff, nil
}

// runGitDirectoryDiff renders git HEAD-vs-working-tree diffs for all changed
// files under a single directory into buf, with styled file headers.
func runGitDirectoryDiff(pairs []gitFilePair, opts *rootOptions, buf *bytes.Buffer) (hasDiff bool, err error) {
	for _, pair := range pairs {
		result, dErr := drift.Diff(pair.HeadContent, pair.WorkContent, opts.driftOpts...)
		if dErr != nil {
			return false, newExitCode(2, dErr.Error())
		}
		if result.IsEqual {
			continue
		}
		hasDiff = true
		writeFileHeader(buf, pair.Name, opts.noColor, opts.termWidth)
		if rErr := drift.RenderWithNames(result, buf, pair.OldName, pair.NewName, opts.driftOpts...); rErr != nil {
			return false, newExitCode(2, rErr.Error())
		}
	}
	return hasDiff, nil
}

// writeThroughPager writes buf to a pager (or directly to opts.streams.Out if paging
// is not needed). It always returns newExitCode(1, "") so callers can do:
//
//	return writeThroughPager(&buf, opts)
func writeThroughPager(buf *bytes.Buffer, opts *rootOptions) error {
	var termHeight int
	if f, ok := opts.streams.Out.(*os.File); ok {
		if _, h, tErr := term.GetSize(f.Fd()); tErr == nil {
			termHeight = h
		}
	}
	lineCount := strings.Count(buf.String(), "\n")
	if shouldPage(opts.streams.Out, lineCount, termHeight, opts.noPager) {
		pagerWriter, cleanup, pErr := startPager(resolvePager(), opts.streams)
		if pErr != nil {
			_, _ = buf.WriteTo(opts.streams.Out)
		} else {
			_, _ = buf.WriteTo(pagerWriter)
			cleanup()
		}
	} else {
		_, _ = buf.WriteTo(opts.streams.Out)
	}
	return newExitCode(1, "")
}

// runRoot is a thin orchestrator: it calls resolveInputs, drift.Diff, and drift.RenderWithNames
// in sequence. No flag parsing or I/O decisions live here.
func runRoot(opts *rootOptions) error {
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
		var buf bytes.Buffer
		hasDiff, err := runGitDirectoryDiff(pairs, opts, &buf)
		if err != nil {
			return err
		}
		if !hasDiff {
			return nil
		}
		return writeThroughPager(&buf, opts)
	}

	// Two-directory diff: both positional args are directories.
	if len(opts.args) == 2 && isDir(opts.args[0]) && isDir(opts.args[1]) {
		pairs, err := diffDirectories(opts.args[0], opts.args[1])
		if err != nil {
			return newExitCode(2, err.Error())
		}
		var buf bytes.Buffer
		hasDiff, err := runDirectoryDiff(pairs, opts, &buf)
		if err != nil {
			return err
		}
		if !hasDiff {
			return nil // identical dirs → exit 0, no output
		}
		return writeThroughPager(&buf, opts)
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
