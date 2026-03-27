// Basic example: functional API (drift.Diff + drift.Render).
// Output is deterministic with drift.WithNoColor(); you can also export NO_COLOR=1.
package main

import (
	"fmt"
	"os"

	"github.com/tylercrawford/drift/drift"
)

func main() {
	old := "// drift-example v1\npackage main\n\nfunc main() {}\n"
	newText := "// drift-example v2\npackage main\n\nfunc main() {\n\tprintln(\"hi\")\n}\n"

	result, err := drift.Diff(old, newText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "diff: %v\n", err)
		os.Exit(1)
	}
	if result.IsEqual {
		return
	}
	if err := drift.Render(result, os.Stdout, drift.WithNoColor(), drift.WithLang("go")); err != nil {
		fmt.Fprintf(os.Stderr, "render: %v\n", err)
		os.Exit(1)
	}
}
