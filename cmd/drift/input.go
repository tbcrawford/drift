package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// resolveInputs returns old/new content and display names for diff headers.
// Either two positional paths (or "-" for stdin), or both --from and --to (non-empty) with no positionals.
func resolveInputs(args []string, fromFlag, toFlag string, stdin io.Reader) (old, new string, oldName, newName string, err error) {
	if fromFlag != "" || toFlag != "" {
		if fromFlag == "" || toFlag == "" {
			return "", "", "", "", fmt.Errorf("invalid arguments: --from and --to must both be set")
		}
		if len(args) != 0 {
			return "", "", "", "", fmt.Errorf("invalid arguments: use either two paths or --from and --to together")
		}
		return fromFlag, toFlag, "a/string", "b/string", nil
	}

	if len(args) != 2 {
		return "", "", "", "", fmt.Errorf("invalid usage: expected drift [flags] OLD NEW (two paths or stdin '-')")
	}

	a, b := args[0], args[1]

	switch {
	case a == "-" && b == "-":
		body, err := io.ReadAll(stdin)
		if err != nil {
			return "", "", "", "", fmt.Errorf("invalid stdin: %w", err)
		}
		s := string(body)
		return s, s, "-", "-", nil
	case a == "-":
		oldb, err := io.ReadAll(stdin)
		if err != nil {
			return "", "", "", "", fmt.Errorf("invalid stdin: %w", err)
		}
		newb, err := os.ReadFile(b)
		if err != nil {
			return "", "", "", "", fmt.Errorf("invalid file %q: %w", b, err)
		}
		return string(oldb), string(newb), "-", filepath.Base(b), nil
	case b == "-":
		oldb, err := os.ReadFile(a)
		if err != nil {
			return "", "", "", "", fmt.Errorf("invalid file %q: %w", a, err)
		}
		newb, err := io.ReadAll(stdin)
		if err != nil {
			return "", "", "", "", fmt.Errorf("invalid stdin: %w", err)
		}
		return string(oldb), string(newb), filepath.Base(a), "-", nil
	default:
		oldb, err := os.ReadFile(a)
		if err != nil {
			return "", "", "", "", fmt.Errorf("invalid file %q: %w", a, err)
		}
		newb, err := os.ReadFile(b)
		if err != nil {
			return "", "", "", "", fmt.Errorf("invalid file %q: %w", b, err)
		}
		return string(oldb), string(newb), filepath.Base(a), filepath.Base(b), nil
	}
}
