package highlight

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/tylercrawford/drift/internal/edittype"
)

// DiffLineStyle returns the background colour for insert/delete full lines. It matches
// terrasort’s full-line AddBg/RemoveBg: [DiffLineBackgroundColour] (Chroma
// GenericInserted/GenericDeleted with the same fallbacks and fg→bg blend as terrasort’s
// chromaDiffLineRGBA). Word-diff highlights use [WordSpanBackgroundColour], blended from
// the same base toward pure red/green.
//
// The second return is false when op is Equal or when no usable background can be derived.
func DiffLineStyle(style *chroma.Style, op edittype.Op, isDark bool) (chroma.Colour, bool) {
	if style == nil || op == edittype.Equal {
		return chroma.Colour(0), false
	}

	var del bool
	switch op {
	case edittype.Delete:
		del = true
	case edittype.Insert:
		del = false
	default:
		return chroma.Colour(0), false
	}

	bg := DiffLineBackgroundColour(style, isDark, del)
	if !bg.IsSet() {
		return chroma.Colour(0), false
	}

	return bg, true
}
