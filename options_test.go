package drift

import (
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
