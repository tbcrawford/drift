package highlight

import (
	"testing"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
)

// blendClose asserts two colours are within ±2 per channel (float rounding).
func blendClose(t *testing.T, label string, got, want chroma.Colour) {
	t.Helper()
	diff := func(a, b uint8) int {
		if a > b {
			return int(a - b)
		}
		return int(b - a)
	}
	if diff(got.Red(), want.Red()) > 2 || diff(got.Green(), want.Green()) > 2 || diff(got.Blue(), want.Blue()) > 2 {
		t.Errorf("%s: got #%02x%02x%02x want ~#%02x%02x%02x",
			label, got.Red(), got.Green(), got.Blue(), want.Red(), want.Green(), want.Blue())
	}
}

// TestBlendChromaTowardTerminalBase_dark verifies dark-terminal blending.
// base dark=(18,18,22); mix=0.78; src=#ff0000:
//
//	r = 255*0.22 + 18*0.78 ≈ 56+14 = 70
//	g = 0*0.22  + 18*0.78 ≈ 14
//	b = 0*0.22  + 22*0.78 ≈ 17
func TestBlendChromaTowardTerminalBase_dark(t *testing.T) {
	c := chroma.MustParseColour("#ff0000")
	got := blendChromaTowardTerminalBase(c, true)
	blendClose(t, "dark red", got, chroma.NewColour(70, 14, 17))
}

// TestBlendChromaTowardTerminalBase_light verifies light-terminal blending.
// base light=(255,255,255); mix=0.78; src=#ff0000:
//
//	r = 255
//	g = 0*0.22 + 255*0.78 ≈ 199
//	b = 0*0.22 + 255*0.78 ≈ 199
func TestBlendChromaTowardTerminalBase_light(t *testing.T) {
	c := chroma.MustParseColour("#ff0000")
	got := blendChromaTowardTerminalBase(c, false)
	blendClose(t, "light red", got, chroma.NewColour(255, 199, 199))
}

// TestFallbackDiffChroma covers all four constant values.
func TestFallbackDiffChroma_allVariants(t *testing.T) {
	cases := []struct {
		isDark, del bool
		want        string
	}{
		{true, true, "#3a2228"},
		{true, false, "#243520"},
		{false, true, "#ffeaea"},
		{false, false, "#e6f7e6"},
	}
	for _, c := range cases {
		got := fallbackDiffChroma(c.isDark, c.del)
		want := chroma.MustParseColour(c.want)
		if got != want {
			t.Errorf("fallbackDiffChroma(isDark=%v,del=%v): got %s want %s", c.isDark, c.del, got, want)
		}
	}
}

// TestDiffEntryChromaColour_usesBackground verifies Background field is preferred.
func TestDiffEntryChromaColour_prefersBackground(t *testing.T) {
	bg := chroma.MustParseColour("#aabbcc")
	e := chroma.StyleEntry{Background: bg, Colour: chroma.MustParseColour("#112233")}
	got := diffEntryChromaColour(e, true)
	if got != bg {
		t.Errorf("expected Background=%s, got %s", bg, got)
	}
}

// TestDiffEntryChromaColour_blendsColourWhenNoBackground verifies Colour blend fallback.
func TestDiffEntryChromaColour_blendsColour(t *testing.T) {
	c := chroma.MustParseColour("#ff0000")
	e := chroma.StyleEntry{Colour: c}
	got := diffEntryChromaColour(e, true)
	if !got.IsSet() {
		t.Error("expected set colour from Colour blend path")
	}
	if got == c {
		t.Error("blended result should differ from source colour")
	}
}

// TestDiffEntryChromaColour_zeroWhenBothUnset verifies empty entry returns zero.
func TestDiffEntryChromaColour_zeroWhenBothUnset(t *testing.T) {
	got := diffEntryChromaColour(chroma.StyleEntry{}, true)
	if got.IsSet() {
		t.Errorf("expected zero (unset) colour, got %s", got)
	}
}

func TestGutterBackgroundHex_matchesTerrasortPalette(t *testing.T) {
	t.Parallel()
	cases := []struct {
		isDark, oldSide bool
		want            string
	}{
		{true, true, "#585858"},
		{true, false, "#444444"},
		{false, true, "#e4e4e4"},
		{false, false, "#eeeeee"},
	}
	for _, c := range cases {
		got := GutterBackgroundHex(c.isDark, c.oldSide)
		if got != c.want {
			t.Errorf("GutterBackgroundHex(%v,%v) = %q, want %q", c.isDark, c.oldSide, got, c.want)
		}
	}
}

func TestGutterForegroundHex_matchesTerrasortUXTheme(t *testing.T) {
	t.Parallel()
	if g, w := GutterDimForegroundHex(true), "#919aa1"; g != w {
		t.Errorf("GutterDimForegroundHex(true) = %q, want %q", g, w)
	}
	if g, w := GutterDimForegroundHex(false), "#64646e"; g != w {
		t.Errorf("GutterDimForegroundHex(false) = %q, want %q", g, w)
	}
	if g, w := GutterHighlightForegroundHex(true), "#d1d7e0"; g != w {
		t.Errorf("GutterHighlightForegroundHex(true) = %q, want %q", g, w)
	}
	if g, w := GutterHighlightForegroundHex(false), "#14141e"; g != w {
		t.Errorf("GutterHighlightForegroundHex(false) = %q, want %q", g, w)
	}
}

func TestDiffLineBackgroundColour_githubStyle_nonZero(t *testing.T) {
	t.Parallel()
	style := styles.Get("github")
	if style == nil {
		t.Fatal("github style")
	}
	for _, del := range []bool{true, false} {
		c := DiffLineBackgroundColour(style, false, del)
		if !c.IsSet() {
			t.Fatalf("del=%v: expected set colour", del)
		}
	}
}

func TestTerrasortParity_DiffLineBias_MonokaiAndGithubDark(t *testing.T) {
	t.Parallel()
	// github-dark: terrasort pipeline yields clearly greener insert vs redder delete.
	gh := styles.Get("github-dark")
	if gh == nil {
		t.Fatal("github-dark style")
	}
	ins := DiffLineBackgroundColour(gh, true, false)
	del := DiffLineBackgroundColour(gh, true, true)
	insGB := int(ins.Green()) - int(ins.Red())
	delGB := int(del.Green()) - int(del.Red())
	if insGB <= delGB {
		t.Fatalf("insert should be greener than delete (G-R ins=%d del=%d)", insGB, delGB)
	}
	delRG := int(del.Red()) - int(del.Green())
	insRG := int(ins.Red()) - int(ins.Green())
	if delRG <= insRG {
		t.Fatalf("delete should be redder than insert (R-G del=%d ins=%d)", delRG, insRG)
	}

	// monokai: some Chroma themes give gd/gi entries that collapse to the same RGB
	// after blending; still require usable colours (parity with terrasort for those styles).
	mk := styles.Get("monokai")
	if mk == nil {
		t.Fatal("monokai style")
	}
	for _, del := range []bool{true, false} {
		c := DiffLineBackgroundColour(mk, true, del)
		if !c.IsSet() {
			t.Fatalf("monokai del=%v: expected set colour", del)
		}
	}
}

func TestWordSpanBrighterThanDiffLinePlane_githubDark(t *testing.T) {
	t.Parallel()
	style := styles.Get("github-dark")
	if style == nil {
		t.Fatal("github-dark style")
	}
	lineDel := DiffLineBackgroundColour(style, true, true)
	wordDel := WordSpanBackgroundColour(style, true, true)
	if wordDel.Red() <= lineDel.Red() {
		t.Fatalf("delete: word span R=%d should exceed full-line R=%d", wordDel.Red(), lineDel.Red())
	}
	lineIns := DiffLineBackgroundColour(style, true, false)
	wordIns := WordSpanBackgroundColour(style, true, false)
	if wordIns.Green() <= lineIns.Green() {
		t.Fatalf("insert: word span G=%d should exceed full-line G=%d", wordIns.Green(), lineIns.Green())
	}
}

func TestLineFallbackFromTerminalRGB_nearBlack(t *testing.T) {
	t.Parallel()
	c := LineFallbackFromTerminalRGB(13, 17, 23, true, false)
	if !c.IsSet() {
		t.Fatal("expected colour")
	}
	// Green tint: G should exceed R for an insert on dark bg.
	if c.Green() <= c.Red() {
		t.Fatalf("expected green bias, got R=%d G=%d B=%d", c.Red(), c.Green(), c.Blue())
	}
}
