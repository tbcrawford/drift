package highlight

import (
	"fmt"
	"math"

	"github.com/alecthomas/chroma/v2"
)

// DiffLineBackgroundColour returns the background for a full diff line (add/remove).
//
// Pipeline (terrasort-aligned):
//  1. diffEntryChromaColour: Background if explicitly set, else blend Colour at mix 0.78
//     toward terminal base — this is terrasort's exact chromaDiffLineRGBA algorithm.
//  2. If the colour returned in step 1 equals the theme's own base background the tint
//     would be invisible (monokai embeds #272822 on every token). Fall through to step 3.
//  3. Blend the theme's base background toward semantic red/green at α=0.18 dark / 0.12
//     light — mirrors terrasort's lineFallbackFromTerminalRGB, using the Chroma base as
//     a proxy for the terminal background.
//  4. Final fallback: hardcoded constants matching terrasort's fallbackDiffRGBA.
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
	baseBackground := style.Get(chroma.Background).Background

	c := diffEntryChromaColour(e, isDark)

	// If the chroma entry yielded a real colour that is distinct from the theme's own
	// background, use it directly (e.g. github-dark → #490202 / #0f5323).
	if c.IsSet() && c != baseBackground {
		return c
	}

	// The colour is either unset or equals the theme's own base BG (monokai case).
	// Synthesise by blending the base BG toward semantic red/green — mirrors terrasort's
	// lineFallbackFromTerminalRGB using the Chroma base as a terminal-BG proxy.
	if baseBackground.IsSet() {
		return LineFallbackFromTerminalRGB(
			baseBackground.Red(), baseBackground.Green(), baseBackground.Blue(),
			isDark, del,
		)
	}

	return fallbackDiffChroma(isDark, del)
}

// diffEntryChromaColour is terrasort's exact algorithm: use Background if set,
// otherwise blend Colour toward terminal base at mix 0.78.
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

// GutterBackgroundHex returns a neutral #RRGGBB for old vs new column when a
// fallback background is needed (e.g. [WordSpanBackgroundColour] unset). Context
// line gutters in the renderer use foreground only (no fill). ANSI 240/238 dark,
// 254/255 light.
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

// GutterDimForegroundHex is the muted gray for context line numbers and for the old/new
// column separator (│). It matches terrasort's UXTheme.DimFg (see terrasort
// internal/highlight/uxtheme.go).
func GutterDimForegroundHex(isDark bool) string {
	if isDark {
		return "#919aa1"
	}
	return "#64646e"
}

// GutterHighlightForegroundHex is the gutter number color on delete/insert rows (on top of
// semantic backgrounds). It matches terrasort's UXTheme.GutterHighlightFg.
func GutterHighlightForegroundHex(isDark bool) string {
	if isDark {
		return "#d1d7e0"
	}
	return "#14141e"
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

// ChromeAccentColor returns a "#RRGGBB" hex color suitable for DeltaTheme chrome
// decoration (Δ glyph, filename, rules, hunk header box). It queries the Chroma
// style's Keyword token color (usually blue or cyan) as the accent. When the style
// is nil or has no usable keyword color, a reasonable fallback is returned:
// #5f87ff (bright blue) for dark terminals and #0050d0 (blue) for light terminals.
func ChromeAccentColor(style *chroma.Style, isDark bool) string {
	if style != nil {
		// Try Keyword first (typically blue in most themes).
		for _, tt := range []chroma.TokenType{chroma.Keyword, chroma.NameFunction, chroma.LiteralString} {
			e := style.Get(tt)
			if e.Colour.IsSet() {
				r, g, b := e.Colour.Red(), e.Colour.Green(), e.Colour.Blue()
				return fmt.Sprintf("#%02x%02x%02x", r, g, b)
			}
		}
	}
	// Fallback: bright blue for dark terminals, deep blue for light terminals.
	if isDark {
		return "#5f87ff"
	}
	return "#0050d0"
}

// blendColourTowardRGB linearly blends c toward (tr,tg,tb); alpha is the weight on the target.
func blendColourTowardRGB(c chroma.Colour, tr, tg, tb uint8, alpha float64) chroma.Colour {
	r := float64(c.Red())*(1-alpha) + float64(tr)*alpha
	g := float64(c.Green())*(1-alpha) + float64(tg)*alpha
	b := float64(c.Blue())*(1-alpha) + float64(tb)*alpha
	return chroma.NewColour(
		uint8(clampFloat(r)),
		uint8(clampFloat(g)),
		uint8(clampFloat(b)),
	)
}

// WordSpanBackgroundColour returns a **brighter** semantic red (delete) or green (insert)
// background for intra-line changed spans. It starts from [DiffLineBackgroundColour] and
// blends toward pure red/green so changed tokens read stronger than the full-line plane.
func WordSpanBackgroundColour(style *chroma.Style, isDark, del bool) chroma.Colour {
	base := DiffLineBackgroundColour(style, isDark, del)
	if !base.IsSet() {
		return chroma.Colour(0)
	}
	if del {
		return blendColourTowardRGB(base, 255, 0, 0, 0.32)
	}
	return blendColourTowardRGB(base, 0, 255, 0, 0.32)
}
