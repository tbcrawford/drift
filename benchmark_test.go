// Benchmarks diff and render on ~10,000-line inputs in unified and split modes.
//
// Expect each benchmark iteration to finish in under one second on a modern laptop;
// smoke test with: go test -bench=. -benchtime=100ms
package drift

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
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
	old, newText := generateLines(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Diff(old, newText, WithAlgorithm(Myers)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRenderUnified10k(b *testing.B) {
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
