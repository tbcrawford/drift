package drift

import (
	"testing"
)

// TestWithLang verifies that WithLang correctly sets config.lang.
func TestWithLang(t *testing.T) {
	cfg := defaultConfig()
	if cfg.lang != "" {
		t.Fatalf("default config.lang = %q; want empty string", cfg.lang)
	}
	WithLang("go")(cfg)
	if cfg.lang != "go" {
		t.Errorf("config.lang after WithLang('go') = %q; want %q", cfg.lang, "go")
	}
}

// TestWithTheme verifies that WithTheme correctly sets config.theme.
func TestWithTheme(t *testing.T) {
	cfg := defaultConfig()
	if cfg.theme != "" {
		t.Fatalf("default config.theme = %q; want empty string", cfg.theme)
	}
	WithTheme("monokai")(cfg)
	if cfg.theme != "monokai" {
		t.Errorf("config.theme after WithTheme('monokai') = %q; want %q", cfg.theme, "monokai")
	}
}

// TestWithNoColor_SetsFlag verifies that WithNoColor sets config.noColor to true.
func TestWithNoColor_SetsFlag(t *testing.T) {
	cfg := defaultConfig()
	if cfg.noColor {
		t.Fatal("default config.noColor should be false")
	}
	WithNoColor()(cfg)
	if !cfg.noColor {
		t.Error("config.noColor after WithNoColor() = false; want true")
	}
}
