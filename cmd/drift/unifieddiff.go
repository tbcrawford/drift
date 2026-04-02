package main

import (
	"bufio"
	"io"
	"strings"
)

// parsedFileDiff holds the reconstructed old/new content for one file
// extracted from a multi-file unified diff stream (e.g. git diff output).
type parsedFileDiff struct {
	// OldName is the display path from "--- a/..." (e.g. "a/cmd/drift/main.go").
	// Empty for newly added files.
	OldName string
	// NewName is the display path from "+++ b/..." (e.g. "b/cmd/drift/main.go").
	// Empty for deleted files.
	NewName string
	// Name is the canonical filename for display headers (stripped of a/ b/ prefix).
	// Derived from NewName when present, OldName otherwise.
	Name string
	// OldContent is the reconstructed full old file content from context + delete lines.
	OldContent string
	// NewContent is the reconstructed full new file content from context + add lines.
	NewContent string
	// IsBinary is true when the diff header indicates binary file changes.
	// When true, OldContent and NewContent are empty.
	IsBinary bool
}

// parseUnifiedDiff reads a multi-file unified diff (as produced by git diff,
// git show, git log -p, etc.) from r and returns one parsedFileDiff per changed file.
// Returns nil (not an error) when the input contains no diffs (empty or identical).
func parseUnifiedDiff(r io.Reader) ([]parsedFileDiff, error) {
	type state int
	const (
		stateHeader state = iota // looking for "diff --git" line
		stateMeta                // inside header block (index, mode, similarity lines)
		stateHunk                // processing +/- lines within a @@ hunk
	)

	var (
		files   []parsedFileDiff
		current *parsedFileDiff
		oldBuf  strings.Builder
		newBuf  strings.Builder
		cur     state = stateHeader
		// isNewFile / isDeletedFile track whether we should skip reconstructing
		// the null side for added/deleted files.
		isNewFile     bool
		isDeletedFile bool
	)

	// flush finalizes the current parsedFileDiff and appends it to files.
	flush := func() {
		if current == nil {
			return
		}
		if !isNewFile {
			current.OldContent = oldBuf.String()
		}
		if !isDeletedFile {
			current.NewContent = newBuf.String()
		}
		// Derive Name from NewName when available, OldName otherwise.
		if current.NewName != "" && current.NewName != "/dev/null" {
			current.Name = stripABPrefix(current.NewName)
		} else if current.OldName != "" && current.OldName != "/dev/null" {
			current.Name = stripABPrefix(current.OldName)
		}
		files = append(files, *current)
		current = nil
		oldBuf.Reset()
		newBuf.Reset()
		isNewFile = false
		isDeletedFile = false
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		// New file block: always starts a fresh parsedFileDiff.
		case strings.HasPrefix(line, "diff --git "):
			flush()
			current = &parsedFileDiff{}
			cur = stateMeta

		// --- /dev/null or --- a/... (header state or hunk state)
		case cur == stateMeta && strings.HasPrefix(line, "--- "):
			path := line[4:]
			if path == "/dev/null" {
				isNewFile = true
				current.OldName = ""
			} else {
				current.OldName = path
				isNewFile = false
			}

		// +++ /dev/null or +++ b/... (header state or hunk state)
		case cur == stateMeta && strings.HasPrefix(line, "+++ "):
			path := line[4:]
			if path == "/dev/null" {
				isDeletedFile = true
				current.NewName = ""
			} else {
				current.NewName = path
				isDeletedFile = false
			}

		// Binary file indicator
		case cur == stateMeta && strings.HasPrefix(line, "Binary files "):
			if current != nil {
				current.IsBinary = true
				// Extract name from "Binary files a/X and b/X differ"
				// The name will be derived from existing OldName/NewName if set;
				// if not, parse from this line.
				if current.OldName == "" && current.NewName == "" {
					// "Binary files a/X and b/Y differ"
					rest := strings.TrimPrefix(line, "Binary files ")
					if andIdx := strings.Index(rest, " and "); andIdx >= 0 {
						current.OldName = rest[:andIdx]
						after := rest[andIdx+5:]
						if spIdx := strings.Index(after, " differ"); spIdx >= 0 {
							current.NewName = after[:spIdx]
						}
					}
				}
			}

		// Hunk header @@ — transition to stateHunk
		case strings.HasPrefix(line, "@@ "):
			if cur != stateHunk {
				cur = stateHunk
			}
			// Hunk headers are not added to content.

		// Content lines within a hunk
		case cur == stateHunk:
			switch {
			case len(line) == 0:
				// Rare: empty line in hunk; treat as context (both sides).
				oldBuf.WriteByte('\n')
				newBuf.WriteByte('\n')
			case line[0] == ' ':
				// Context line — appears in both old and new.
				content := line[1:] + "\n"
				if !isNewFile {
					oldBuf.WriteString(content)
				}
				if !isDeletedFile {
					newBuf.WriteString(content)
				}
			case line[0] == '-':
				// Delete line — old only.
				if !isNewFile {
					oldBuf.WriteString(line[1:] + "\n")
				}
			case line[0] == '+':
				// Add line — new only.
				if !isDeletedFile {
					newBuf.WriteString(line[1:] + "\n")
				}
			case line[0] == '\\':
				// "\ No newline at end of file" — git metadata, skip.
				// Strip the trailing \n we may have added to the previous line.
				if oldBuf.Len() > 0 {
					s := oldBuf.String()
					if s[len(s)-1] == '\n' {
						oldBuf.Reset()
						oldBuf.WriteString(s[:len(s)-1])
					}
				}
				if newBuf.Len() > 0 {
					s := newBuf.String()
					if s[len(s)-1] == '\n' {
						newBuf.Reset()
						newBuf.WriteString(s[:len(s)-1])
					}
				}
			}

		// In stateMeta: skip index, mode, similarity, rename from/to lines, etc.
		default:
			// no-op; unrecognized meta lines are ignored
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	flush()
	return files, nil
}

// stripABPrefix removes the leading "a/" or "b/" prefix that git adds to diff paths.
// If the path does not have such a prefix, it is returned unchanged.
func stripABPrefix(path string) string {
	if strings.HasPrefix(path, "a/") || strings.HasPrefix(path, "b/") {
		return path[2:]
	}
	return path
}
