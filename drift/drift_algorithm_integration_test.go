package drift_test

import (
	"strings"
	"testing"

	"github.com/tylercrawford/drift/drift"
	"github.com/tylercrawford/drift/internal/testhelpers"
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

func TestWithAlgorithm_Myers_StillDefault(t *testing.T) {
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
		t.Fatalf("round-trip failed with Myers (default):\n  want: %q\n  got:  %q", wantText, gotText)
	}
}

func TestAllAlgorithmsCorrect(t *testing.T) {
	algos := []drift.Algorithm{drift.Myers, drift.Patience, drift.Histogram}
	algoNames := []string{"Myers", "Patience", "Histogram"}

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
