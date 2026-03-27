// Package worddiff computes intra-line alignment between two strings for diff
// highlighting. It runs the same Myers SES used for line diffs over per-character
// or per-word tokens, then merges adjacent same-status tokens into Segments.
package worddiff

import (
	"regexp"

	"github.com/tylercrawford/drift/internal/algo/myers"
	"github.com/tylercrawford/drift/internal/edittype"
)

var tokenRE = regexp.MustCompile(`\S+|\s+`)

// Token is a slice of the original string with byte offsets.
type Token struct {
	Text       string
	Start, End int
}

// Segment is a half-open byte range [Start, End) in the original string and
// whether that region is part of an insertion/deletion relative to the other side.
type Segment struct {
	Start, End int
	Changed    bool
}

// tokenize splits s into word and whitespace tokens with byte offsets.
func tokenize(s string) []Token {
	var out []Token
	for _, idx := range tokenRE.FindAllStringIndex(s, -1) {
		start, end := idx[0], idx[1]
		out = append(out, Token{Text: s[start:end], Start: start, End: end})
	}
	return out
}

// tokenizeChars splits s into one Token per Unicode code point with byte offsets.
func tokenizeChars(s string) []Token {
	var out []Token
	i := 0
	for _, r := range s {
		size := len(string(r))
		out = append(out, Token{Text: string(r), Start: i, End: i + size})
		i += size
	}
	return out
}

func stringsFromTokens(toks []Token) []string {
	s := make([]string, len(toks))
	for i := range toks {
		s[i] = toks[i].Text
	}
	return s
}

// PairSegments returns contiguous segments on old and new for muted vs changed styling.
// Tokens are words and whitespace runs. For character-level granularity use [PairCharSegments].
// When old == new, both sides get a single unchanged segment covering the full string.
func PairSegments(old, new string) (oldSegs, newSegs []Segment) {
	return pairWithTokenizer(old, new, tokenize)
}

// PairCharSegments is like [PairSegments] but tokenizes at Unicode code-point granularity,
// producing the smallest possible changed spans — individual characters.
func PairCharSegments(old, new string) (oldSegs, newSegs []Segment) {
	return pairWithTokenizer(old, new, tokenizeChars)
}

func pairWithTokenizer(old, new string, tok func(string) []Token) (oldSegs, newSegs []Segment) {
	if old == new {
		if old == "" {
			return nil, nil
		}
		s := Segment{Start: 0, End: len(old), Changed: false}
		return []Segment{s}, []Segment{s}
	}

	ot := tok(old)
	nt := tok(new)
	if len(ot) == 0 && len(nt) == 0 {
		return nil, nil
	}

	os := stringsFromTokens(ot)
	ns := stringsFromTokens(nt)
	edits := myers.New().Diff(os, ns)

	oldChanged := make([]bool, len(ot))
	newChanged := make([]bool, len(nt))

	for _, e := range edits {
		switch e.Op {
		case edittype.Equal:
			// oldChanged and newChanged are already false from make(); no assignment needed.
		case edittype.Delete:
			if e.OldLine > 0 && e.OldLine <= len(ot) {
				oldChanged[e.OldLine-1] = true
			}
		case edittype.Insert:
			if e.NewLine > 0 && e.NewLine <= len(nt) {
				newChanged[e.NewLine-1] = true
			}
		}
	}

	return mergeTokenMarks(ot, oldChanged), mergeTokenMarks(nt, newChanged)
}

func mergeTokenMarks(toks []Token, changed []bool) []Segment {
	if len(toks) == 0 {
		return nil
	}
	var out []Segment
	cur := Segment{Start: toks[0].Start, End: toks[0].End, Changed: changed[0]}
	for i := 1; i < len(toks); i++ {
		if changed[i] == cur.Changed && toks[i].Start == cur.End {
			cur.End = toks[i].End
			continue
		}
		out = append(out, cur)
		cur = Segment{Start: toks[i].Start, End: toks[i].End, Changed: changed[i]}
	}
	out = append(out, cur)
	return out
}
