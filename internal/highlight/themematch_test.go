package highlight_test

import (
	"image/color"
	"testing"

	"github.com/tylercrawford/drift/internal/highlight"
)

func TestBestMatchTheme(t *testing.T) {
	t.Run("nil and empty return monokai", func(t *testing.T) {
		if got := highlight.BestMatchTheme(nil); got != "monokai" {
			t.Errorf("nil palette: got %q want monokai", got)
		}
		if got := highlight.BestMatchTheme([]color.RGBA{}); got != "monokai" {
			t.Errorf("empty palette: got %q want monokai", got)
		}
	})

	t.Run("fixed small palette deterministic winner", func(t *testing.T) {
		// Black, red, green, blue — recorded once via BestMatchTheme against chroma v2.23.1
		palette := []color.RGBA{
			{R: 0, G: 0, B: 0, A: 255},
			{R: 255, G: 0, B: 0, A: 255},
			{R: 0, G: 255, B: 0, A: 255},
			{R: 0, G: 0, B: 255, A: 255},
		}
		const want = "igor"
		if got := highlight.BestMatchTheme(palette); got != want {
			t.Errorf("BestMatchTheme(...) = %q, want %q", got, want)
		}
	})
}

func TestBestMatchTheme_deterministic(t *testing.T) {
	palette := []color.RGBA{
		{R: 10, G: 20, B: 30, A: 255},
		{R: 100, G: 110, B: 120, A: 255},
	}
	a := highlight.BestMatchTheme(palette)
	b := highlight.BestMatchTheme(palette)
	if a != b {
		t.Fatalf("non-deterministic: %q vs %q", a, b)
	}
}
