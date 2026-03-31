package drift

import (
	"strings"
	"testing"
)

// TestWithLang verifies that WithLang correctly sets config.render.lang.
func TestWithLang(t *testing.T) {
	cfg := defaultConfig()
	if cfg.render.lang != "" {
		t.Fatalf("default config.render.lang = %q; want empty string", cfg.render.lang)
	}
	WithLang("go")(cfg)
	if cfg.render.lang != "go" {
		t.Errorf("config.render.lang after WithLang('go') = %q; want %q", cfg.render.lang, "go")
	}
}

// TestWithTheme verifies that WithTheme correctly sets config.render.theme.
func TestWithTheme(t *testing.T) {
	cfg := defaultConfig()
	if cfg.render.theme != "" {
		t.Fatalf("default config.render.theme = %q; want empty string", cfg.render.theme)
	}
	WithTheme("monokai")(cfg)
	if cfg.render.theme != "monokai" {
		t.Errorf("config.render.theme after WithTheme('monokai') = %q; want %q", cfg.render.theme, "monokai")
	}
}

// TestWithNoColor_SetsFlag verifies that WithNoColor sets config.render.noColor to true.
func TestWithNoColor_SetsFlag(t *testing.T) {
	cfg := defaultConfig()
	if cfg.render.noColor {
		t.Fatal("default config.render.noColor should be false")
	}
	WithNoColor()(cfg)
	if !cfg.render.noColor {
		t.Error("config.render.noColor after WithNoColor() = false; want true")
	}
}

// TestWithContextNegative verifies that passing WithContext(-1) to Diff() returns
// a non-nil error containing "non-negative".
func TestWithContextNegative(t *testing.T) {
	old := "line1\nline2\nline3"
	new := "line1\nlineX\nline3"
	_, err := Diff(old, new, WithContext(-1))
	if err == nil {
		t.Fatal("expected non-nil error for WithContext(-1), got nil")
	}
	if !strings.Contains(err.Error(), "non-negative") {
		t.Errorf("error message should mention 'non-negative', got: %s", err.Error())
	}
}

// TestWithContextZero verifies that passing WithContext(0) to Diff() succeeds
// and returns a non-equal diff result (zero context lines is a valid choice).
func TestWithContextZero(t *testing.T) {
	old := "line1\nline2\nline3"
	new := "line1\nlineX\nline3"
	result, err := Diff(old, new, WithContext(0))
	if err != nil {
		t.Fatalf("unexpected error for WithContext(0): %v", err)
	}
	if result.IsEqual {
		t.Error("expected non-equal diff result")
	}
}
