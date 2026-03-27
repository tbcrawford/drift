package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tylercrawford/drift/drift"
)

// stdinReader is swapped in tests via runCLI.
var stdinReader io.Reader = os.Stdin

var rootCmd = &cobra.Command{
	Use:   "drift [flags] OLD NEW | FILE",
	Short: "Pretty-print a diff between two inputs",
	Long: `Pretty-print a diff between two inputs.
With one path inside a git repository, diffs the working tree against HEAD.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          validateRootArgs,
	RunE:          runRoot,
}

func validateRootArgs(cmd *cobra.Command, args []string) error {
	from, err := cmd.Flags().GetString("from")
	if err != nil {
		return err
	}
	to, err := cmd.Flags().GetString("to")
	if err != nil {
		return err
	}

	if from != "" || to != "" {
		if len(args) != 0 {
			return fmt.Errorf("invalid arguments: with --from/--to do not pass file paths")
		}
		return nil
	}

	switch len(args) {
	case 0:
		return fmt.Errorf("invalid usage: expected drift [flags] OLD NEW, a single FILE in a git repo, or --from/--to")
	case 1, 2:
		return nil
	default:
		return fmt.Errorf("invalid usage: too many positional arguments")
	}
}

func parseAlgorithm(s string) (drift.Algorithm, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "myers":
		return drift.Myers, nil
	case "patience":
		return drift.Patience, nil
	case "histogram":
		return drift.Histogram, nil
	default:
		return 0, newExitCode(2, fmt.Sprintf("invalid algorithm: %q (use myers, patience, histogram)", s))
	}
}

func buildDriftOptions(cmd *cobra.Command) ([]drift.Option, error) {
	var opts []drift.Option

	algStr, err := cmd.Flags().GetString("algorithm")
	if err != nil {
		return nil, err
	}
	a, err := parseAlgorithm(algStr)
	if err != nil {
		return nil, err
	}
	opts = append(opts, drift.WithAlgorithm(a))

	n, err := cmd.Flags().GetInt("context")
	if err != nil {
		return nil, err
	}
	if n < 0 {
		return nil, newExitCode(2, "invalid context: must be non-negative")
	}
	opts = append(opts, drift.WithContext(n))

	noColor, err := cmd.Flags().GetBool("no-color")
	if err != nil {
		return nil, err
	}
	if noColor {
		opts = append(opts, drift.WithNoColor())
	}

	lang, err := cmd.Flags().GetString("lang")
	if err != nil {
		return nil, err
	}
	if lang != "" {
		opts = append(opts, drift.WithLang(lang))
	}

	theme, err := cmd.Flags().GetString("theme")
	if err != nil {
		return nil, err
	}
	if theme != "" {
		opts = append(opts, drift.WithTheme(theme))
	}

	showTheme, err := cmd.Flags().GetBool("show-theme")
	if err != nil {
		return nil, err
	}
	if showTheme {
		opts = append(opts, drift.WithThemeResolved(func(name string) {
			fmt.Fprintf(os.Stderr, "drift: resolved syntax theme: %s\n", name)
		}))
	}

	split, err := cmd.Flags().GetBool("split")
	if err != nil {
		return nil, err
	}
	if split {
		opts = append(opts, drift.WithSplit())
	}

	noLineNumbers, err := cmd.Flags().GetBool("no-line-numbers")
	if err != nil {
		return nil, err
	}
	if noLineNumbers {
		opts = append(opts, drift.WithoutLineNumbers())
	}

	return opts, nil
}

func runRoot(cmd *cobra.Command, args []string) error {
	from, err := cmd.Flags().GetString("from")
	if err != nil {
		return err
	}
	to, err := cmd.Flags().GetString("to")
	if err != nil {
		return err
	}

	old, newText, oldName, newName, err := resolveInputs(args, from, to, stdinReader)
	if err != nil {
		return err
	}

	opts, err := buildDriftOptions(cmd)
	if err != nil {
		return err
	}

	result, err := drift.Diff(old, newText, opts...)
	if err != nil {
		return newExitCode(2, err.Error())
	}

	if result.IsEqual {
		return nil
	}

	if err := drift.RenderWithNames(result, cmd.OutOrStdout(), oldName, newName, opts...); err != nil {
		return newExitCode(2, err.Error())
	}

	// Differences rendered successfully; exit 1 without stderr noise.
	return newExitCode(1, "")
}

func runCLI(stdout, stderr io.Writer, stdin io.Reader, args []string) int {
	prev := stdinReader
	stdinReader = stdin
	defer func() { stdinReader = prev }()

	// Tests and repeated invocations share rootCmd; clear flags so prior --from/--to do not leak.
	_ = rootCmd.Flags().Set("from", "")
	_ = rootCmd.Flags().Set("to", "")

	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()
	if err == nil {
		return 0
	}

	var ec *exitCodeErr
	if errors.As(err, &ec) {
		if ec.msg != "" {
			fmt.Fprintln(stderr, ec.msg)
		}
		return ec.code
	}

	fmt.Fprintln(stderr, err)
	return 2
}

func executeDrift() int {
	return runCLI(os.Stdout, os.Stderr, os.Stdin, os.Args[1:])
}

func init() {
	rootCmd.Flags().Bool("split", false, "side-by-side split view")
	rootCmd.Flags().Bool("no-line-numbers", false, "hide old/new line-number gutters")
	rootCmd.Flags().String("algorithm", "myers", "diff algorithm: myers, patience, histogram")
	rootCmd.Flags().String("lang", "", "Chroma language override (e.g. go, python)")
	rootCmd.Flags().String("theme", "", "Chroma style/theme override")
	rootCmd.Flags().Bool("no-color", false, "disable ANSI colors")
	rootCmd.Flags().Int("context", 3, "lines of context around hunks")
	rootCmd.Flags().String("from", "", "old text as a raw string (use with --to)")
	rootCmd.Flags().String("to", "", "new text as a raw string (use with --from)")
	rootCmd.Flags().Bool("show-theme", false, "print resolved Chroma theme to stderr after selection")
	_ = rootCmd.Flags().MarkHidden("show-theme")
}

func main() {
	os.Exit(executeDrift())
}
