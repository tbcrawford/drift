package highlight_test

import (
	"image/color"
	"testing"

	"github.com/tbcrawford/drift/internal/highlight"
)

func TestParseOSC4Responses_twoSlots(t *testing.T) {
	// Slot 0 black, slot 7 gray — sparse slice length 8
	raw := []byte("\x1b]4;0;rgb:0000/0000/0000\x07\x1b]4;7;rgb:8080/8080/8080\x07")
	palette := highlight.ParseOSC4Responses(raw)
	if palette == nil {
		t.Fatal("expected non-nil palette")
	}
	if len(palette) != 8 {
		t.Fatalf("len=%d want 8", len(palette))
	}
	if palette[0] != (color.RGBA{R: 0, G: 0, B: 0, A: 255}) {
		t.Errorf("palette[0] = %+v", palette[0])
	}
	if palette[7] != (color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 255}) {
		t.Errorf("palette[7] = %+v", palette[7])
	}
}

func TestParseOSC4Responses_nilEmpty(t *testing.T) {
	if highlight.ParseOSC4Responses(nil) != nil {
		t.Error("nil input want nil")
	}
	if highlight.ParseOSC4Responses([]byte{}) != nil {
		t.Error("empty input want nil")
	}
}
