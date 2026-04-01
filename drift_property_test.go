package drift_test

import (
	"fmt"
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/tbcrawford/drift"
	"github.com/tbcrawford/drift/internal/testhelpers"
)

// TestProperty_RoundTrip verifies the fundamental diff invariant:
// Apply(Diff(old, new), canonicalLines(old)) == canonicalLines(new) for all inputs.
//
// This is the core correctness guarantee: given any two texts, applying the
// diff to the old text must perfectly reconstruct the new text.
//
// Note on normalization: drift.Diff internally splits text on "\n" and strips
// a trailing empty element (matching the behavior of files that end with \n).
// The invariant is therefore expressed over canonical (normalized) line slices,
// not raw generator output. We compare joined text to avoid slice-shape ambiguity
// between []string{} and []string{""} which both join to "".
func TestProperty_RoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate raw line slices (no \r or \n within a line).
		// SliceOfN(elem, -1, 50) means minLen=-1 (unconstrained), maxLen=50.
		oldRaw := rapid.SliceOfN(
			rapid.StringMatching(`[^\r\n]*`), -1, 50,
		).Draw(t, "oldLines")

		newRaw := rapid.SliceOfN(
			rapid.StringMatching(`[^\r\n]*`), -1, 50,
		).Draw(t, "newLines")

		// Join to text strings (canonical diff input).
		oldText := strings.Join(oldRaw, "\n")
		newText := strings.Join(newRaw, "\n")

		// Compute canonical line slices by re-splitting (mirrors drift internals).
		// This eliminates ambiguity: joinLines(["", ""]) == "\n" but
		// drift treats it as one empty line, not two. Working with canonical
		// slices ensures Apply's output is comparable.
		oldLines := canonicalLines(oldText)

		result, err := drift.Diff(oldText, newText)
		if err != nil {
			t.Fatalf("Diff returned unexpected error: %v", err)
		}

		got := testhelpers.Apply(result, oldLines)

		// Compare as joined text. Note: drift normalizes trailing newlines by
		// stripping a trailing empty element from splitLines. So the invariant
		// is: join(Apply(...)) == join(canonicalLines(newText)).
		// E.g., newText="\n" is normalized to [""], which joins back to "".
		// We compare against the canonical (normalized) form of newText.
		gotText := strings.Join(got, "\n")
		wantText := strings.Join(canonicalLines(newText), "\n")
		if gotText != wantText {
			t.Fatalf(
				"round-trip invariant failed:\n  oldText:  %q\n  newText:  %q\n  wantText: %q\n  gotText:  %q\n  hunks:    %s",
				oldText, newText, wantText, gotText, formatHunks(result),
			)
		}
	})
}

// canonicalLines splits text into lines using the same logic as drift.splitLines:
// split on "\n" and strip the trailing empty element produced by a trailing newline.
func canonicalLines(text string) []string {
	lines := strings.Split(text, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return []string{}
	}
	return lines
}

// TestProperty_IdenticalInputs verifies that diffing a text against itself
// always produces IsEqual=true and an empty Hunks slice.
func TestProperty_IdenticalInputs(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		lines := rapid.SliceOfN(
			rapid.StringMatching(`[^\r\n]*`), -1, 50,
		).Draw(t, "lines")

		text := strings.Join(lines, "\n")

		result, err := drift.Diff(text, text)
		if err != nil {
			t.Fatalf("Diff returned unexpected error: %v", err)
		}

		if !result.IsEqual {
			t.Fatalf("expected IsEqual=true for identical inputs, got false\n  text: %q", text)
		}
		if len(result.Hunks) != 0 {
			t.Fatalf("expected 0 hunks for identical inputs, got %d", len(result.Hunks))
		}
	})
}

// TestProperty_HunkAccounting verifies that hunk line counts do not exceed
// the actual sizes of the old and new files.
//
// Invariant: sum(h.OldLines) <= len(oldLines) and sum(h.NewLines) <= len(newLines).
func TestProperty_HunkAccounting(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		oldLines := rapid.SliceOfN(
			rapid.StringMatching(`[^\r\n]*`), -1, 50,
		).Draw(t, "oldLines")

		newLines := rapid.SliceOfN(
			rapid.StringMatching(`[^\r\n]*`), -1, 50,
		).Draw(t, "newLines")

		oldText := strings.Join(oldLines, "\n")
		newText := strings.Join(newLines, "\n")

		result, err := drift.Diff(oldText, newText)
		if err != nil {
			t.Fatalf("Diff returned unexpected error: %v", err)
		}

		var totalOld, totalNew int
		for _, h := range result.Hunks {
			totalOld += h.OldLines
			totalNew += h.NewLines
		}

		if totalOld > len(oldLines) {
			t.Fatalf(
				"hunk OldLines sum (%d) exceeds len(oldLines) (%d)\n  oldLines: %v\n  newLines: %v\n  hunks: %s",
				totalOld, len(oldLines), oldLines, newLines, formatHunks(result),
			)
		}
		if totalNew > len(newLines) {
			t.Fatalf(
				"hunk NewLines sum (%d) exceeds len(newLines) (%d)\n  oldLines: %v\n  newLines: %v\n  hunks: %s",
				totalNew, len(newLines), oldLines, newLines, formatHunks(result),
			)
		}
	})
}

// TestProperty_RoundTrip_Patience verifies the round-trip invariant using the Patience algorithm:
// Apply(Diff(old, new, WithAlgorithm(Patience)), canonicalLines(old)) == canonicalLines(new).
func TestProperty_RoundTrip_Patience(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		oldRaw := rapid.SliceOfN(
			rapid.StringMatching(`[^\r\n]*`), -1, 50,
		).Draw(t, "oldLines")

		newRaw := rapid.SliceOfN(
			rapid.StringMatching(`[^\r\n]*`), -1, 50,
		).Draw(t, "newLines")

		oldText := strings.Join(oldRaw, "\n")
		newText := strings.Join(newRaw, "\n")
		oldLines := canonicalLines(oldText)

		result, err := drift.Diff(oldText, newText, drift.WithAlgorithm(drift.Patience))
		if err != nil {
			t.Fatalf("Diff returned unexpected error: %v", err)
		}

		got := testhelpers.Apply(result, oldLines)

		gotText := strings.Join(got, "\n")
		wantText := strings.Join(canonicalLines(newText), "\n")
		if gotText != wantText {
			t.Fatalf(
				"round-trip invariant failed (Patience):\n  oldText:  %q\n  newText:  %q\n  wantText: %q\n  gotText:  %q\n  hunks:    %s",
				oldText, newText, wantText, gotText, formatHunks(result),
			)
		}
	})
}

// TestProperty_RoundTrip_Histogram verifies the round-trip invariant using the Histogram algorithm:
// Apply(Diff(old, new, WithAlgorithm(Histogram)), canonicalLines(old)) == canonicalLines(new).
func TestProperty_RoundTrip_Histogram(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		oldRaw := rapid.SliceOfN(
			rapid.StringMatching(`[^\r\n]*`), -1, 50,
		).Draw(t, "oldLines")

		newRaw := rapid.SliceOfN(
			rapid.StringMatching(`[^\r\n]*`), -1, 50,
		).Draw(t, "newLines")

		oldText := strings.Join(oldRaw, "\n")
		newText := strings.Join(newRaw, "\n")
		oldLines := canonicalLines(oldText)

		result, err := drift.Diff(oldText, newText, drift.WithAlgorithm(drift.Histogram))
		if err != nil {
			t.Fatalf("Diff returned unexpected error: %v", err)
		}

		got := testhelpers.Apply(result, oldLines)

		gotText := strings.Join(got, "\n")
		wantText := strings.Join(canonicalLines(newText), "\n")
		if gotText != wantText {
			t.Fatalf(
				"round-trip invariant failed (Histogram):\n  oldText:  %q\n  newText:  %q\n  wantText: %q\n  gotText:  %q\n  hunks:    %s",
				oldText, newText, wantText, gotText, formatHunks(result),
			)
		}
	})
}

// TestProperty_RoundTrip_Auto verifies the round-trip invariant using the Auto algorithm:
// Apply(Diff(old, new, WithAlgorithm(Auto)), canonicalLines(old)) == canonicalLines(new).
func TestProperty_RoundTrip_Auto(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		oldRaw := rapid.SliceOfN(
			rapid.StringMatching(`[^\r\n]*`), -1, 50,
		).Draw(t, "oldLines")

		newRaw := rapid.SliceOfN(
			rapid.StringMatching(`[^\r\n]*`), -1, 50,
		).Draw(t, "newLines")

		oldText := strings.Join(oldRaw, "\n")
		newText := strings.Join(newRaw, "\n")
		oldLines := canonicalLines(oldText)

		result, err := drift.Diff(oldText, newText, drift.WithAlgorithm(drift.Auto))
		if err != nil {
			t.Fatalf("Diff returned unexpected error: %v", err)
		}

		got := testhelpers.Apply(result, oldLines)

		gotText := strings.Join(got, "\n")
		wantText := strings.Join(canonicalLines(newText), "\n")
		if gotText != wantText {
			t.Fatalf(
				"round-trip invariant failed (Auto):\n  oldText:  %q\n  newText:  %q\n  wantText: %q\n  gotText:  %q\n  hunks:    %s",
				oldText, newText, wantText, gotText, formatHunks(result),
			)
		}
	})
}

// formatHunks returns a human-readable summary of the hunks for failure messages.
func formatHunks(result drift.DiffResult) string {
	if len(result.Hunks) == 0 {
		return "(no hunks)"
	}
	var sb strings.Builder
	for i, h := range result.Hunks {
		fmt.Fprintf(&sb, "\n  hunk[%d]: @@ -%d,%d +%d,%d @@",
			i, h.OldStart, h.OldLines, h.NewStart, h.NewLines)
		for _, l := range h.Lines {
			var prefix string
			switch l.Op {
			case drift.Equal:
				prefix = " "
			case drift.Insert:
				prefix = "+"
			case drift.Delete:
				prefix = "-"
			}
			fmt.Fprintf(&sb, "\n    %s%q", prefix, l.Content)
		}
	}
	return sb.String()
}
