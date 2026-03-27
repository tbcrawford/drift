// Builder example: drift.New() chain with NoColor for stable stdout (or set NO_COLOR=1).
package main

import (
	"fmt"
	"os"

	"github.com/tylercrawford/drift/drift"
)

func main() {
	old := "name: app\nversion: 1\n"
	newText := "name: app\nversion: 2\n"

	b := drift.New().Algorithm(drift.Myers).NoColor()
	result, err := b.Diff(old, newText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "diff: %v\n", err)
		os.Exit(1)
	}
	if result.IsEqual {
		return
	}
	if err := b.Render(result, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "render: %v\n", err)
		os.Exit(1)
	}
}
