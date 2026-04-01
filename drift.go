package drift

import (
	"strings"

	"github.com/tbcrawford/drift/internal/algo/histogram"
	"github.com/tbcrawford/drift/internal/algo/myers"
	"github.com/tbcrawford/drift/internal/algo/patience"
	"github.com/tbcrawford/drift/internal/hunk"
)

// Diff computes the line-level diff between old and new using the configured
// algorithm and returns a structured DiffResult.
//
// Line endings are normalized: Windows \r\n is treated as \n.
// When both inputs are identical, DiffResult.IsEqual is true and Hunks is empty.
//
// An error is returned when the configuration is invalid, for example when
// WithContext is called with a negative value.
func Diff(old, new string, opts ...Option) (DiffResult, error) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}

	if err := cfg.validate(); err != nil {
		return DiffResult{}, err
	}

	// Normalize line endings: \r\n → \n
	old = strings.ReplaceAll(old, "\r\n", "\n")
	new = strings.ReplaceAll(new, "\r\n", "\n")

	// Fast path: identical inputs
	if old == new {
		return DiffResult{IsEqual: true}, nil
	}

	// Split into lines, trimming trailing empty string from final newline
	oldLines := splitLines(old)
	newLines := splitLines(new)

	// Dispatch to algorithm
	var differ algoInterface
	switch cfg.diff.algorithm {
	case Patience:
		differ = patience.New()
	case Histogram:
		differ = histogram.New()
	case Auto:
		differ = selectAuto(oldLines, newLines)
	default: // Myers (and invalid values)
		differ = myers.New()
	}

	edits := differ.Diff(oldLines, newLines)
	hunks := hunk.Build(edits, oldLines, newLines, cfg.diff.contextLines)

	return DiffResult{
		Hunks:   hunks,
		IsEqual: len(hunks) == 0,
	}, nil
}

// splitLines splits text by \n and removes the trailing empty element
// produced when the text ends with a newline.
func splitLines(text string) []string {
	lines := strings.Split(text, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// algoInterface is the internal interface for algorithm implementations.
// Matches internal/algo.Differ but defined here to avoid an import cycle:
// internal/algo/myers imports the root package (for drift.Edit), so the root
// package must not import internal/algo.
type algoInterface interface {
	Diff(oldLines, newLines []string) []Edit
}

// selectAuto chooses between Myers and Histogram based on file characteristics.
// It uses Histogram for small files (≤ 2000 total lines) where no old-side line
// appears more than 32 times, and Myers otherwise. The O(N) frequency scan is
// skipped entirely for files that exceed the size threshold.
func selectAuto(old, new []string) algoInterface {
	const (
		smallFileThreshold = 2000 // total lines (old + new)
		maxFreqThreshold   = 32   // max occurrences of any single old-side line
	)
	if len(old)+len(new) > smallFileThreshold {
		return myers.New()
	}
	freq := make(map[string]int, len(old))
	for _, l := range old {
		freq[l]++
	}
	for _, count := range freq {
		if count > maxFreqThreshold {
			return myers.New()
		}
	}
	return histogram.New()
}
