package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "drift [flags] OLD NEW",
	Short:         "Pretty-print a diff between two inputs",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.MaximumNArgs(2),
	RunE:          runRoot,
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
	_, _, _, _, err = resolveInputs(args, from, to, os.Stdin)
	if err != nil {
		return err
	}
	return fmt.Errorf("not implemented: complete plan 05-03")
}

func init() {
	rootCmd.Flags().Bool("split", false, "side-by-side split view")
	rootCmd.Flags().String("algorithm", "myers", "diff algorithm: myers, patience, histogram")
	rootCmd.Flags().String("lang", "", "Chroma language override (e.g. go, python)")
	rootCmd.Flags().String("theme", "", "Chroma style/theme override")
	rootCmd.Flags().Bool("no-color", false, "disable ANSI colors")
	rootCmd.Flags().Int("context", 3, "lines of context around hunks")
	rootCmd.Flags().String("from", "", "old text as a raw string (use with --to)")
	rootCmd.Flags().String("to", "", "new text as a raw string (use with --from)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
