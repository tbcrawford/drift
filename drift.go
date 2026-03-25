// Package drift provides a production-ready text diffing library with rich
// terminal output, syntax highlighting, and support for unified and side-by-side
// diff formats.
//
// Usage:
//
//	result, err := drift.Diff(original, modified, drift.WithSyntaxHighlight("go"))
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Print(result)
package drift
