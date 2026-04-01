package drift_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tbcrawford/drift"
	"github.com/tbcrawford/drift/internal/testhelpers"
)

// integrationCanonicalLines splits text into lines the same way drift internals do.
func integrationCanonicalLines(text string) []string {
	lines := strings.Split(text, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return []string{}
	}
	return lines
}

const (
	integrationOld = "func A() {\n    doThing1()\n}\n\nfunc B() {\n    doThing2()\n}\n"
	integrationNew = "func A() {\n    doThing1()\n    doThing3()\n}\n\nfunc B() {\n    doThing2()\n}\n"
)

func TestWithAlgorithm_Patience_RoundTrip(t *testing.T) {
	result, err := drift.Diff(integrationOld, integrationNew, drift.WithAlgorithm(drift.Patience))
	if err != nil {
		t.Fatalf("Diff returned unexpected error: %v", err)
	}
	if result.IsEqual {
		t.Fatal("expected IsEqual=false for differing inputs")
	}

	got := testhelpers.Apply(result, integrationCanonicalLines(integrationOld))
	gotText := strings.Join(got, "\n")
	wantText := strings.Join(integrationCanonicalLines(integrationNew), "\n")
	if gotText != wantText {
		t.Fatalf("round-trip failed with Patience:\n  want: %q\n  got:  %q", wantText, gotText)
	}
}

func TestWithAlgorithm_Histogram_RoundTrip(t *testing.T) {
	result, err := drift.Diff(integrationOld, integrationNew, drift.WithAlgorithm(drift.Histogram))
	if err != nil {
		t.Fatalf("Diff returned unexpected error: %v", err)
	}
	if result.IsEqual {
		t.Fatal("expected IsEqual=false for differing inputs")
	}

	got := testhelpers.Apply(result, integrationCanonicalLines(integrationOld))
	gotText := strings.Join(got, "\n")
	wantText := strings.Join(integrationCanonicalLines(integrationNew), "\n")
	if gotText != wantText {
		t.Fatalf("round-trip failed with Histogram:\n  want: %q\n  got:  %q", wantText, gotText)
	}
}

// TestDefault_Algorithm_RoundTrip verifies that calling Diff without an explicit
// algorithm (which uses Auto by default) produces correct round-trip output.
func TestDefault_Algorithm_RoundTrip(t *testing.T) {
	result, err := drift.Diff(integrationOld, integrationNew)
	if err != nil {
		t.Fatalf("Diff returned unexpected error: %v", err)
	}
	if result.IsEqual {
		t.Fatal("expected IsEqual=false for differing inputs")
	}

	got := testhelpers.Apply(result, integrationCanonicalLines(integrationOld))
	gotText := strings.Join(got, "\n")
	wantText := strings.Join(integrationCanonicalLines(integrationNew), "\n")
	if gotText != wantText {
		t.Fatalf("round-trip failed with Auto (default):\n  want: %q\n  got:  %q", wantText, gotText)
	}
}

func TestAllAlgorithmsCorrect(t *testing.T) {
	algos := []drift.Algorithm{drift.Myers, drift.Patience, drift.Histogram, drift.Auto}
	algoNames := []string{"Myers", "Patience", "Histogram", "Auto"}

	cases := []struct{ name, old, new string }{
		{"simple insert", "a\nb\nd\n", "a\nb\nc\nd\n"},
		{"simple delete", "a\nb\nc\nd\n", "a\nb\nd\n"},
		{"all replaced", "x\ny\nz\n", "a\nb\nc\n"},
		{"identical", "same\nlines\n", "same\nlines\n"},
		{"empty to nonempty", "", "a\nb\n"},
	}

	for _, tc := range cases {
		for i, algo := range algos {
			t.Run(tc.name+"/"+algoNames[i], func(t *testing.T) {
				result, err := drift.Diff(tc.old, tc.new, drift.WithAlgorithm(algo))
				if err != nil {
					t.Fatalf("Diff returned unexpected error: %v", err)
				}

				got := testhelpers.Apply(result, integrationCanonicalLines(tc.old))
				gotText := strings.Join(got, "\n")
				wantText := strings.Join(integrationCanonicalLines(tc.new), "\n")
				if gotText != wantText {
					t.Fatalf("round-trip invariant failed:\n  algo: %s\n  old: %q\n  new: %q\n  want: %q\n  got:  %q",
						algoNames[i], tc.old, tc.new, wantText, gotText)
				}
			})
		}
	}
}

// TestWithAlgorithm_Auto_RoundTrip verifies that Auto satisfies the round-trip
// invariant: apply(diff(a, b), a) == b for a representative code-like input.
func TestWithAlgorithm_Auto_RoundTrip(t *testing.T) {
	result, err := drift.Diff(integrationOld, integrationNew, drift.WithAlgorithm(drift.Auto))
	if err != nil {
		t.Fatalf("Diff returned unexpected error: %v", err)
	}
	if result.IsEqual {
		t.Fatal("expected IsEqual=false for differing inputs")
	}
	got := testhelpers.Apply(result, integrationCanonicalLines(integrationOld))
	gotText := strings.Join(got, "\n")
	wantText := strings.Join(integrationCanonicalLines(integrationNew), "\n")
	if gotText != wantText {
		t.Fatalf("round-trip failed with Auto:\n  want: %q\n  got:  %q", wantText, gotText)
	}
}

// TestAuto_SelectsHistogram_SmallCleanFile verifies that Auto chooses Histogram
// (and thus produces the same result as explicit Histogram) for a small file
// with no high-frequency lines.
func TestAuto_SelectsHistogram_SmallCleanFile(t *testing.T) {
	// integrationOld/New are small and use unique-ish lines — Auto should pick Histogram.
	auto, err := drift.Diff(integrationOld, integrationNew, drift.WithAlgorithm(drift.Auto))
	if err != nil {
		t.Fatalf("Diff(Auto) error: %v", err)
	}
	hist, err := drift.Diff(integrationOld, integrationNew, drift.WithAlgorithm(drift.Histogram))
	if err != nil {
		t.Fatalf("Diff(Histogram) error: %v", err)
	}
	// Both must produce correct output (round-trip).
	autoGot := testhelpers.Apply(auto, integrationCanonicalLines(integrationOld))
	histGot := testhelpers.Apply(hist, integrationCanonicalLines(integrationOld))
	want := integrationCanonicalLines(integrationNew)
	if strings.Join(autoGot, "\n") != strings.Join(want, "\n") {
		t.Fatalf("Auto round-trip failed on small clean file")
	}
	if strings.Join(histGot, "\n") != strings.Join(want, "\n") {
		t.Fatalf("Histogram round-trip failed on small clean file")
	}
}

// TestAuto_SelectsMyers_HighFrequency verifies that Auto falls back to Myers
// when a line in the old text appears more than 32 times.
func TestAuto_SelectsMyers_HighFrequency(t *testing.T) {
	// Build a file where "}" appears 33 times — above the maxFreqThreshold.
	repeated := strings.Repeat("}\n", 33)
	old := repeated + "unique_old_line\n"
	new := repeated + "unique_new_line\n"

	result, err := drift.Diff(old, new, drift.WithAlgorithm(drift.Auto))
	if err != nil {
		t.Fatalf("Diff(Auto) error: %v", err)
	}
	// Round-trip must hold regardless of which algorithm was chosen.
	got := testhelpers.Apply(result, integrationCanonicalLines(old))
	if strings.Join(got, "\n") != strings.Join(integrationCanonicalLines(new), "\n") {
		t.Fatalf("Auto round-trip failed on high-frequency input")
	}
}

// TestAuto_SelectsMyers_LargeFile verifies that Auto falls back to Myers for
// files exceeding the 2000-total-line threshold.
func TestAuto_SelectsMyers_LargeFile(t *testing.T) {
	// Build old = 1001 lines, new = 1001 lines: total = 2002 > 2000 threshold.
	var sb strings.Builder
	for i := range 1001 {
		fmt.Fprintf(&sb, "line %04d\n", i)
	}
	old := sb.String()
	sb.Reset()
	for i := range 1001 {
		if i == 500 {
			fmt.Fprintf(&sb, "changed line\n")
		} else {
			fmt.Fprintf(&sb, "line %04d\n", i)
		}
	}
	new := sb.String()

	result, err := drift.Diff(old, new, drift.WithAlgorithm(drift.Auto))
	if err != nil {
		t.Fatalf("Diff(Auto) error: %v", err)
	}
	got := testhelpers.Apply(result, integrationCanonicalLines(old))
	if strings.Join(got, "\n") != strings.Join(integrationCanonicalLines(new), "\n") {
		t.Fatalf("Auto round-trip failed on large file")
	}
}
