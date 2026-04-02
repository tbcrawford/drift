// Benchmarks diff and render on ~10,000-line inputs in unified and split modes.
//
// Expect each benchmark iteration to finish in under one second on a modern laptop;
// smoke test with: go test -bench=. -benchtime=100ms
//
// Phase 23 Performance Optimization Results:
//
// Baseline (pre-optimization, Phase 23-01):
//
//	BenchmarkDiff10k-14                                3308    361504 ns/op      855252 B/op        16 allocs/op
//	BenchmarkRenderUnified10k-14                        650   2020439 ns/op     1029465 B/op     26436 allocs/op
//	BenchmarkRenderSplit10k-14                          500   2473840 ns/op     1558587 B/op     36039 allocs/op
//	BenchmarkRenderUnified10kColor-14                   234   5700141 ns/op     1955308 B/op     52387 allocs/op
//	BenchmarkRenderSplit10kColor-14                     211   5862243 ns/op     3363175 B/op     58522 allocs/op
//	BenchmarkRenderSplitWithLineNumbers10kColor-14      217   5528098 ns/op     3349645 B/op     58521 allocs/op
//
// Post-optimization (Plan 23-02 — gutter cache + direct ANSI SGR builder):
//
//	BenchmarkDiff10k-14                                3396    352790 ns/op      855250 B/op        16 allocs/op  (unchanged)
//	BenchmarkRenderUnified10k-14                        610   1972759 ns/op     1036870 B/op     26445 allocs/op  (unchanged)
//	BenchmarkRenderSplit10k-14                          505   2427972 ns/op     1559211 B/op     36048 allocs/op  (unchanged)
//	BenchmarkRenderUnified10kColor-14                   280   4516417 ns/op     1834636 B/op     43827 allocs/op  (-21% ns/op, -16.4% allocs/op)
//	BenchmarkRenderSplit10kColor-14                     233   5072443 ns/op     3336400 B/op     54690 allocs/op  (-6.5% ns/op)
//	BenchmarkRenderSplitWithLineNumbers10kColor-14      227   5293071 ns/op     3345171 B/op     54690 allocs/op  (-4.2% ns/op)
package drift

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
)

func generateLines(n int) (old, new string) {
	lines := make([]string, n)
	for i := 0; i < n; i++ {
		lines[i] = fmt.Sprintf("line %08d\n", i)
	}
	old = strings.Join(lines, "")
	mid := n / 2
	chunk := 200
	if chunk > n {
		chunk = n
	}
	newLines := append([]string(nil), lines...)
	for i := mid; i < mid+chunk && i < n; i++ {
		newLines[i] = fmt.Sprintf("line %08d changed\n", i)
	}
	new = strings.Join(newLines, "")
	return old, new
}

func BenchmarkDiff10k(b *testing.B) {
	b.ReportAllocs()
	old, newText := generateLines(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Diff(old, newText, WithAlgorithm(Myers)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRenderUnified10k(b *testing.B) {
	b.ReportAllocs()
	old, newText := generateLines(10000)
	result, err := Diff(old, newText, WithAlgorithm(Myers))
	if err != nil {
		b.Fatal(err)
	}
	var buf bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := Render(result, &buf, WithNoColor()); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRenderSplit10k(b *testing.B) {
	b.ReportAllocs()
	old, newText := generateLines(10000)
	result, err := Diff(old, newText, WithAlgorithm(Myers))
	if err != nil {
		b.Fatal(err)
	}
	var buf bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := Render(result, &buf, WithNoColor(), WithSplit()); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRenderUnified10kColor benchmarks the expensive color rendering path
// (no WithNoColor — exercises per-line lipgloss gutter style and Chroma token highlighting).
func BenchmarkRenderUnified10kColor(b *testing.B) {
	b.ReportAllocs()
	old, newText := generateLines(10000)
	result, err := Diff(old, newText, WithAlgorithm(Myers))
	if err != nil {
		b.Fatal(err)
	}
	var buf bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := Render(result, &buf,
			WithTermWidth(200),
			WithColorProfile(colorprofile.TrueColor),
		); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRenderSplit10kColor benchmarks the expensive color rendering path for split view.
func BenchmarkRenderSplit10kColor(b *testing.B) {
	b.ReportAllocs()
	old, newText := generateLines(10000)
	result, err := Diff(old, newText, WithAlgorithm(Myers))
	if err != nil {
		b.Fatal(err)
	}
	var buf bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := Render(result, &buf,
			WithSplit(),
			WithTermWidth(200),
			WithColorProfile(colorprofile.TrueColor),
		); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRenderSplitWithLineNumbers10kColor benchmarks split view with line numbers
// and color enabled — exercises gutterStyleForCell for every line in the hot path.
func BenchmarkRenderSplitWithLineNumbers10kColor(b *testing.B) {
	b.ReportAllocs()
	old, newText := generateLines(10000)
	result, err := Diff(old, newText, WithAlgorithm(Myers))
	if err != nil {
		b.Fatal(err)
	}
	var buf bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := Render(result, &buf,
			WithSplit(),
			WithLineNumbers(true),
			WithTermWidth(200),
			WithColorProfile(colorprofile.TrueColor),
		); err != nil {
			b.Fatal(err)
		}
	}
}
