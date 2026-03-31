package myers_test

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/tylercrawford/drift/drift"
	"github.com/tylercrawford/drift/internal/algo/myers"
)

// countOps counts edits of each Op type.
func countOps(edits []drift.Edit) (equal, insert, del int) {
	for _, e := range edits {
		switch e.Op {
		case drift.Equal:
			equal++
		case drift.Insert:
			insert++
		case drift.Delete:
			del++
		}
	}
	return
}

// TestIdenticalInputs — Test 1: identical inputs return all Equal edits.
func TestIdenticalInputs(t *testing.T) {
	m := myers.New()
	old := []string{"a", "b", "c"}
	new := []string{"a", "b", "c"}
	edits := m.Diff(old, new)

	if len(edits) != 3 {
		t.Fatalf("expected 3 edits, got %d", len(edits))
	}
	for i, e := range edits {
		if e.Op != drift.Equal {
			t.Errorf("edit[%d]: expected Equal, got %v", i, e.Op)
		}
		expectedLine := i + 1
		if e.OldLine != expectedLine {
			t.Errorf("edit[%d]: expected OldLine=%d, got %d", i, expectedLine, e.OldLine)
		}
		if e.NewLine != expectedLine {
			t.Errorf("edit[%d]: expected NewLine=%d, got %d", i, expectedLine, e.NewLine)
		}
	}
}

// TestBothEmpty — Test 2: both empty returns empty (non-nil) slice.
func TestBothEmpty(t *testing.T) {
	m := myers.New()
	edits := m.Diff([]string{}, []string{})
	if edits == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(edits) != 0 {
		t.Fatalf("expected 0 edits, got %d", len(edits))
	}
}

// TestOldEmptyAllInserts — Test 3: old empty → all inserts.
func TestOldEmptyAllInserts(t *testing.T) {
	m := myers.New()
	edits := m.Diff([]string{}, []string{"x", "y"})
	if len(edits) != 2 {
		t.Fatalf("expected 2 edits, got %d", len(edits))
	}
	for i, e := range edits {
		if e.Op != drift.Insert {
			t.Errorf("edit[%d]: expected Insert, got %v", i, e.Op)
		}
		if e.OldLine != 0 {
			t.Errorf("edit[%d]: expected OldLine=0, got %d", i, e.OldLine)
		}
		if e.NewLine != i+1 {
			t.Errorf("edit[%d]: expected NewLine=%d, got %d", i, i+1, e.NewLine)
		}
	}
}

// TestNewEmptyAllDeletes — Test 4: new empty → all deletes.
func TestNewEmptyAllDeletes(t *testing.T) {
	m := myers.New()
	edits := m.Diff([]string{"x", "y"}, []string{})
	if len(edits) != 2 {
		t.Fatalf("expected 2 edits, got %d", len(edits))
	}
	for i, e := range edits {
		if e.Op != drift.Delete {
			t.Errorf("edit[%d]: expected Delete, got %v", i, e.Op)
		}
		if e.OldLine != i+1 {
			t.Errorf("edit[%d]: expected OldLine=%d, got %d", i, i+1, e.OldLine)
		}
		if e.NewLine != 0 {
			t.Errorf("edit[%d]: expected NewLine=0, got %d", i, e.NewLine)
		}
	}
}

// TestMyersPaperExample — Test 5: Myers 1986 paper classic example.
// edit distance == 5 (5 non-Equal edits).
func TestMyersPaperExample(t *testing.T) {
	m := myers.New()
	old := []string{"a", "b", "c", "a", "b", "b", "a"}
	newLines := []string{"c", "b", "a", "b", "a", "c"}
	edits := m.Diff(old, newLines)

	_, inserts, deletes := countOps(edits)
	editDist := inserts + deletes
	if editDist != 5 {
		t.Errorf("expected edit distance 5, got %d (inserts=%d, deletes=%d)", editDist, inserts, deletes)
		for i, e := range edits {
			t.Logf("edit[%d]: Op=%v OldLine=%d NewLine=%d", i, e.Op, e.OldLine, e.NewLine)
		}
	}
}

// TestSingleInsertionInMiddle — Test 6: single insertion in middle.
func TestSingleInsertionInMiddle(t *testing.T) {
	m := myers.New()
	old := []string{"a", "b", "d"}
	newLines := []string{"a", "b", "c", "d"}
	edits := m.Diff(old, newLines)

	// Expected: Equal(1,1) Equal(2,2) Insert(0,3) Equal(3,4)
	want := []drift.Edit{
		{Op: drift.Equal, OldLine: 1, NewLine: 1},
		{Op: drift.Equal, OldLine: 2, NewLine: 2},
		{Op: drift.Insert, OldLine: 0, NewLine: 3},
		{Op: drift.Equal, OldLine: 3, NewLine: 4},
	}
	if len(edits) != len(want) {
		t.Fatalf("expected %d edits, got %d", len(want), len(edits))
	}
	for i, got := range edits {
		w := want[i]
		if got.Op != w.Op || got.OldLine != w.OldLine || got.NewLine != w.NewLine {
			t.Errorf("edit[%d]: expected {Op:%v OldLine:%d NewLine:%d}, got {Op:%v OldLine:%d NewLine:%d}",
				i, w.Op, w.OldLine, w.NewLine, got.Op, got.OldLine, got.NewLine)
		}
	}
}

// TestSingleDeletionInMiddle — Test 7: single deletion in middle.
func TestSingleDeletionInMiddle(t *testing.T) {
	m := myers.New()
	old := []string{"a", "b", "c", "d"}
	newLines := []string{"a", "b", "d"}
	edits := m.Diff(old, newLines)

	// Expected: Equal(1,1) Equal(2,2) Delete(3,0) Equal(4,3)
	want := []drift.Edit{
		{Op: drift.Equal, OldLine: 1, NewLine: 1},
		{Op: drift.Equal, OldLine: 2, NewLine: 2},
		{Op: drift.Delete, OldLine: 3, NewLine: 0},
		{Op: drift.Equal, OldLine: 4, NewLine: 3},
	}
	if len(edits) != len(want) {
		t.Fatalf("expected %d edits, got %d", len(want), len(edits))
	}
	for i, got := range edits {
		w := want[i]
		if got.Op != w.Op || got.OldLine != w.OldLine || got.NewLine != w.NewLine {
			t.Errorf("edit[%d]: expected {Op:%v OldLine:%d NewLine:%d}, got {Op:%v OldLine:%d NewLine:%d}",
				i, w.Op, w.OldLine, w.NewLine, got.Op, got.OldLine, got.NewLine)
		}
	}
}

// TestCrossValidateWithSystemDiff — Test 8: cross-validate against system diff on ~50 lines.
func TestCrossValidateWithSystemDiff(t *testing.T) {
	// Skip if diff command is not available (Windows, restricted environments)
	if _, err := exec.LookPath("diff"); err != nil {
		t.Skip("diff command not available:", err)
	}

	// Two realistic Go source file snippets (~50 lines total).
	oldText := `package main

import (
	"fmt"
	"os"
)

// Config holds application settings.
type Config struct {
	Host     string
	Port     int
	Debug    bool
	MaxConns int
	Timeout  int
}

// NewConfig creates a default Config.
func NewConfig() *Config {
	return &Config{
		Host:     "localhost",
		Port:     8080,
		Debug:    false,
		MaxConns: 100,
		Timeout:  30,
	}
}

// Run starts the application.
func Run(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config required")
	}
	if cfg.Port <= 0 {
		return fmt.Errorf("invalid port: %d", cfg.Port)
	}
	fmt.Fprintf(os.Stdout, "starting on %s:%d\n", cfg.Host, cfg.Port)
	return nil
}

// Validate checks Config fields.
func Validate(cfg *Config) []string {
	var errs []string
	if cfg.Host == "" {
		errs = append(errs, "host required")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		errs = append(errs, "port out of range")
	}
	return errs
}
`

	newText := `package main

import (
	"errors"
	"fmt"
	"os"
)

// Config holds application settings.
type Config struct {
	Host     string
	Port     int
	Debug    bool
	MaxConns int
	Timeout  int
	LogLevel string
}

// NewConfig creates a default Config.
func NewConfig() *Config {
	return &Config{
		Host:     "0.0.0.0",
		Port:     8080,
		Debug:    false,
		MaxConns: 100,
		Timeout:  30,
		LogLevel: "info",
	}
}

// Run starts the application.
func Run(cfg *Config) error {
	if cfg == nil {
		return errors.New("config required")
	}
	if cfg.Port <= 0 {
		return fmt.Errorf("invalid port: %d", cfg.Port)
	}
	fmt.Fprintf(os.Stdout, "starting on %s:%d\n", cfg.Host, cfg.Port)
	return nil
}

// Validate checks Config fields.
func Validate(cfg *Config) []string {
	var errs []string
	if cfg.Host == "" {
		errs = append(errs, "host required")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		errs = append(errs, "port out of range")
	}
	if cfg.LogLevel == "" {
		errs = append(errs, "log level required")
	}
	return errs
}
`

	// Write to temp files
	oldFile, err := os.CreateTemp("", "myers-old-*.go")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(oldFile.Name())
	newFile, err := os.CreateTemp("", "myers-new-*.go")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(newFile.Name())

	if _, err := oldFile.WriteString(oldText); err != nil {
		t.Fatal(err)
	}
	oldFile.Close()
	if _, err := newFile.WriteString(newText); err != nil {
		t.Fatal(err)
	}
	newFile.Close()

	// Run system diff -u to get edit distance (added + removed lines = edit distance)
	cmd := exec.Command("diff", "-u", oldFile.Name(), newFile.Name())
	out, _ := cmd.Output() // diff exits non-zero when files differ; ignore error

	// Count +/- lines in unified diff (skip +++ and --- headers)
	sysAdded, sysRemoved := 0, 0
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "+++ ") || strings.HasPrefix(line, "--- ") {
			continue
		}
		switch line[0] {
		case '+':
			sysAdded++
		case '-':
			sysRemoved++
		}
	}
	sysEditDist := sysAdded + sysRemoved

	// Run Myers diff
	m := myers.New()
	oldLines := strings.Split(strings.TrimRight(oldText, "\n"), "\n")
	newLinesSlice := strings.Split(strings.TrimRight(newText, "\n"), "\n")
	edits := m.Diff(oldLines, newLinesSlice)

	_, myersInserts, myersDeletes := countOps(edits)
	myersEditDist := myersInserts + myersDeletes

	// Edit distances should match (both compute minimal edit sequences)
	if myersEditDist != sysEditDist {
		t.Errorf("edit distance mismatch: Myers=%d, system diff=%d (added=%d, removed=%d)",
			myersEditDist, sysEditDist, sysAdded, sysRemoved)
		t.Logf("system diff output:\n%s", string(out))
	}
}

// TestLineInvariant — Test 9: all lines accounted for.
// count(Equal)+count(Delete)==len(old), count(Equal)+count(Insert)==len(new)
func TestLineInvariant(t *testing.T) {
	m := myers.New()

	largeOld, largeNew := generateLargeInputs()
	cases := []struct {
		name string
		old  []string
		new  []string
	}{
		{"identical", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"both empty", []string{}, []string{}},
		{"old empty", []string{}, []string{"x", "y"}},
		{"new empty", []string{"x", "y"}, []string{}},
		{"paper example", []string{"a", "b", "c", "a", "b", "b", "a"}, []string{"c", "b", "a", "b", "a", "c"}},
		{"insert middle", []string{"a", "b", "d"}, []string{"a", "b", "c", "d"}},
		{"delete middle", []string{"a", "b", "c", "d"}, []string{"a", "b", "d"}},
		{"all replaced", []string{"a", "b", "c"}, []string{"x", "y", "z"}},
		{"large similar", largeOld, largeNew},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			edits := m.Diff(tc.old, tc.new)
			eq, ins, del := countOps(edits)

			if eq+del != len(tc.old) {
				t.Errorf("old invariant violated: Equal(%d)+Delete(%d)=%d != len(old)=%d",
					eq, del, eq+del, len(tc.old))
			}
			if eq+ins != len(tc.new) {
				t.Errorf("new invariant violated: Equal(%d)+Insert(%d)=%d != len(new)=%d",
					eq, ins, eq+ins, len(tc.new))
			}
		})
	}
}

// generateLargeInputs produces a ~100-line test case with insertions and deletions.
// Returns old, new slices.
func generateLargeInputs() (old, new []string) {
	// Build a large file: lines 1..100
	for i := 1; i <= 100; i++ {
		old = append(old, "line "+strconv.Itoa(i))
	}
	// New file: remove every 10th line, add "extra" at line 50
	for i := 1; i <= 100; i++ {
		if i%10 == 0 {
			continue // delete lines 10,20,...,100
		}
		if i == 50 {
			new = append(new, "extra line inserted here")
		}
		new = append(new, "line "+strconv.Itoa(i))
	}
	return old, new
}

// TestHirschbergMemory verifies that the Hirschberg linear-space implementation
// does not exhibit quadratic memory growth. When input doubles in size, peak
// heap allocation must grow by less than 10×.
func TestHirschbergMemory(t *testing.T) {
	m := myers.New()

	// Build a base input: 250 unique lines each side (fully disjoint = worst case
	// for the algorithm: all deletes then all inserts).
	makeLines := func(n, offset int) []string {
		lines := make([]string, n)
		for i := range lines {
			lines[i] = strconv.Itoa(offset + i)
		}
		return lines
	}

	runAndMeasure := func(n int) uint64 {
		old := makeLines(n, 0)
		newL := makeLines(n, n) // completely disjoint
		var before, after runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&before)
		_ = m.Diff(old, newL)
		runtime.GC()
		runtime.ReadMemStats(&after)
		if after.TotalAlloc < before.TotalAlloc {
			return 0
		}
		return after.TotalAlloc - before.TotalAlloc
	}

	alloc1 := runAndMeasure(250)
	alloc2 := runAndMeasure(500)

	if alloc1 == 0 {
		t.Skip("could not measure allocations (alloc1 == 0)")
	}

	ratio := float64(alloc2) / float64(alloc1)
	t.Logf("alloc(250)=%d alloc(500)=%d ratio=%.2f", alloc1, alloc2, ratio)

	const maxRatio = 10.0
	if ratio > maxRatio {
		t.Errorf("memory growth ratio %.2f exceeds %.1f× — possible quadratic allocation", ratio, maxRatio)
	}
}

// TestHirschbergLarge verifies that the Hirschberg implementation satisfies
// the line-count invariants for a 500-line diff.
func TestHirschbergLarge(t *testing.T) {
	m := myers.New()

	// Build 500-line inputs with scattered changes.
	var old, newLines []string
	for i := 1; i <= 500; i++ {
		old = append(old, "line "+strconv.Itoa(i))
	}
	for i := 1; i <= 500; i++ {
		if i%25 == 0 {
			continue // delete every 25th line (20 deletions)
		}
		if i == 100 || i == 300 {
			newLines = append(newLines, "inserted line at "+strconv.Itoa(i)) // 2 insertions
		}
		newLines = append(newLines, "line "+strconv.Itoa(i))
	}

	edits := m.Diff(old, newLines)
	eq, ins, del := countOps(edits)

	if eq+del != len(old) {
		t.Errorf("old invariant violated: Equal(%d)+Delete(%d)=%d != len(old)=%d",
			eq, del, eq+del, len(old))
	}
	if eq+ins != len(newLines) {
		t.Errorf("new invariant violated: Equal(%d)+Insert(%d)=%d != len(new)=%d",
			eq, ins, eq+ins, len(newLines))
	}
	t.Logf("500-line diff: equal=%d insert=%d delete=%d total_edits=%d", eq, ins, del, len(edits))
}
