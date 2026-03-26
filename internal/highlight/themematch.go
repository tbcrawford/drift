package highlight

import (
	"image/color"
	"math"
	"sort"

	chroma "github.com/alecthomas/chroma/v2"
	chromastyles "github.com/alecthomas/chroma/v2/styles"
)

// defaultBestMatchTheme is returned when the palette is empty or no theme could be scored.
const defaultBestMatchTheme = "monokai"

func euclideanDist(a, b color.RGBA) float64 {
	dr := float64(a.R) - float64(b.R)
	dg := float64(a.G) - float64(b.G)
	db := float64(a.B) - float64(b.B)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// BestMatchTheme returns the Chroma theme name whose sampled token colors best match
// the terminal palette (sum of nearest-neighbor Euclidean distances per token).
// Lower total distance is better. Tie-break: first name in alphabetical order.
// An empty or nil palette yields "monokai".
func BestMatchTheme(palette []color.RGBA) string {
	if len(palette) == 0 {
		return defaultBestMatchTheme
	}

	sampleTokens := []chroma.TokenType{
		chroma.Keyword,
		chroma.LiteralString,
		chroma.Comment,
		chroma.NameBuiltin,
		chroma.LiteralNumber,
	}

	names := chromastyles.Names()
	sort.Strings(names)

	bestName := ""
	bestScore := math.MaxFloat64
	scored := false

	for _, name := range names {
		style := chromastyles.Get(name)
		if style == nil {
			continue
		}

		var themeScore float64
		tokenCount := 0

		for _, tok := range sampleTokens {
			entry := style.Get(tok)
			if !entry.Colour.IsSet() {
				continue
			}
			tokenColor := color.RGBA{
				R: entry.Colour.Red(),
				G: entry.Colour.Green(),
				B: entry.Colour.Blue(),
				A: 255,
			}
			minDist := math.MaxFloat64
			for _, slot := range palette {
				if d := euclideanDist(tokenColor, slot); d < minDist {
					minDist = d
				}
			}
			themeScore += minDist
			tokenCount++
		}

		if tokenCount == 0 {
			continue
		}

		if !scored || themeScore < bestScore {
			bestScore = themeScore
			bestName = name
			scored = true
		}
	}

	if !scored {
		return defaultBestMatchTheme
	}
	return bestName
}
