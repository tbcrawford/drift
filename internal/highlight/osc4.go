package highlight

import (
	"image/color"
	"regexp"
	"strconv"
)

// osc4Re matches a single OSC 4 terminal palette response:
//
//	\033]4;{n};rgb:{rr}/{gg}/{bb}\007
//
// Capture groups: (1) slot index, (2) red hex4, (3) green hex4, (4) blue hex4.
var osc4Re = regexp.MustCompile(`\x1b\]4;(\d+);rgb:([0-9a-fA-F]{4})/([0-9a-fA-F]{4})/([0-9a-fA-F]{4})\x07`)

// ParseOSC4Responses parses raw OSC 4 terminal responses and returns a palette slice.
// Input: raw bytes containing zero or more OSC 4 responses concatenated.
// Output: []color.RGBA indexed by slot number (sparse — only found slots populated).
//
// Returns a slice of length=max(slot)+1, with zero values for missing slots.
// Returns nil if input is empty or no valid responses found.
func ParseOSC4Responses(raw []byte) []color.RGBA {
	if len(raw) == 0 {
		return nil
	}

	matches := osc4Re.FindAllStringSubmatch(string(raw), -1)
	if len(matches) == 0 {
		return nil
	}

	maxSlot := -1
	for _, m := range matches {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		if n > maxSlot {
			maxSlot = n
		}
	}
	if maxSlot < 0 {
		return nil
	}

	palette := make([]color.RGBA, maxSlot+1)

	for _, m := range matches {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}

		r, err := parseHex4(m[2])
		if err != nil {
			continue
		}
		g, err := parseHex4(m[3])
		if err != nil {
			continue
		}
		b, err := parseHex4(m[4])
		if err != nil {
			continue
		}

		palette[n] = color.RGBA{R: r, G: g, B: b, A: 255}
	}

	return palette
}

func parseHex4(s string) (uint8, error) {
	v, err := strconv.ParseUint(s, 16, 16)
	if err != nil {
		return 0, err
	}
	return uint8(v >> 8), nil
}
