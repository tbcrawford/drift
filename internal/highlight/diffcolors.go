package highlight

import (
	"math"

	"github.com/alecthomas/chroma/v2"
)

// DiffLineBackgroundColour returns the background for a full diff line (add/remove).
//
// It mirrors terrasort’s chromaDiffLineRGBA pipeline: [chroma.GenericInserted] /
// [chroma.GenericDeleted] entries, preferring Background, then blending Colour toward
// the terminal base (dark (18,18,22) vs light (255,255,255)) at mix 0.78 when only
// foreground is set, then the same fallback hexes as terrasort [fallbackDiffRGBA].
func DiffLineBackgroundColour(style *chroma.Style, isDark, del bool) chroma.Colour {
	if style == nil {
		return fallbackDiffChroma(isDark, del)
	}
	var tt chroma.TokenType
	if del {
		tt = chroma.GenericDeleted
	} else {
		tt = chroma.GenericInserted
	}
	e := style.Get(tt)
	c := diffEntryChromaColour(e, isDark)
	if !c.IsSet() {
		return fallbackDiffChroma(isDark, del)
	}
	return c
}

func diffEntryChromaColour(e chroma.StyleEntry, isDark bool) chroma.Colour {
	if e.Background.IsSet() {
		return e.Background
	}
	if e.Colour.IsSet() {
		return blendChromaTowardTerminalBase(e.Colour, isDark)
	}
	return chroma.Colour(0)
}

func blendChromaTowardTerminalBase(c chroma.Colour, isDark bool) chroma.Colour {
	var br, bg, bb uint8
	if isDark {
		br, bg, bb = 18, 18, 22
	} else {
		br, bg, bb = 255, 255, 255
	}
	const mix = 0.78
	r := float64(c.Red())*(1-mix) + float64(br)*mix
	g := float64(c.Green())*(1-mix) + float64(bg)*mix
	b := float64(c.Blue())*(1-mix) + float64(bb)*mix
	return chroma.NewColour(
		uint8(clampFloat(r)),
		uint8(clampFloat(g)),
		uint8(clampFloat(b)),
	)
}

func fallbackDiffChroma(isDark, del bool) chroma.Colour {
	if isDark {
		if del {
			return chroma.MustParseColour("#3a2228")
		}
		return chroma.MustParseColour("#243520")
	}
	if del {
		return chroma.MustParseColour("#ffeaea")
	}
	return chroma.MustParseColour("#e6f7e6")
}

// GutterBackgroundHex returns a neutral #RRGGBB background for gutter cells on
// context (unchanged) lines — old vs new column use slightly different grays (ANSI
// 240/238 dark, 254/255 light). Delete/insert rows use the same neutrals for
// line-number columns; semantic add/remove backgrounds apply to the full code line
// via [DiffLineBackgroundColour] / [DiffLineStyle].
func GutterBackgroundHex(isDark, oldSide bool) string {
	if isDark {
		if oldSide {
			// ANSI 240
			return "#585858"
		}
		// ANSI 238
		return "#444444"
	}
	if oldSide {
		// ANSI 254
		return "#e4e4e4"
	}
	// ANSI 255
	return "#eeeeee"
}

// LineFallbackFromTerminalRGB blends the terminal background toward semantic green/red
// when no Chroma theme is available (same formula as terrasort's
// lineFallbackFromTerminalRGB). Returns a Chroma colour for use with Lip Gloss or TTY.
func LineFallbackFromTerminalRGB(r, g, b uint8, isDark, del bool) chroma.Colour {
	tr, tg, tb := redGreenTarget(del)
	alpha := 0.18
	if !isDark {
		alpha = 0.12
	}
	return chroma.NewColour(
		uint8(clampFloat(float64(r)*(1-alpha)+float64(tr)*alpha)),
		uint8(clampFloat(float64(g)*(1-alpha)+float64(tg)*alpha)),
		uint8(clampFloat(float64(b)*(1-alpha)+float64(tb)*alpha)),
	)
}

func redGreenTarget(del bool) (tr, tg, tb uint8) {
	if del {
		return 255, 0, 0
	}
	return 0, 255, 0
}

func clampFloat(v float64) float64 {
	return math.Max(0, math.Min(255, v))
}
