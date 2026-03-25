// Package drift provides a production-quality text diff library and CLI for Go.
//
// Drift computes line-level diffs between two multi-line strings using the Myers,
// Patience, or Histogram algorithms and renders output with Chroma syntax highlighting
// via unified or side-by-side split layouts.
//
// Quick start:
//
//	result, err := drift.Diff(oldText, newText)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result)
//
// See the examples/ directory for runnable examples.
package drift
