package highlight

import (
	"testing"

	"github.com/alecthomas/chroma/v2/styles"
)

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
