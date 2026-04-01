package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

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

// runRoot is a thin orchestrator: it calls resolveInputs, drift.Diff, and drift.RenderWithNames
// in sequence. No flag parsing or I/O decisions live here.
func runRoot(opts *rootOptions) error {
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

	if err := drift.RenderWithNames(result, opts.streams.Out, oldName, newName, opts.driftOpts...); err != nil {
		return newExitCode(2, err.Error())
	}

	// Differences rendered successfully; exit 1 without stderr noise.
	return newExitCode(1, "")
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
