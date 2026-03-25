package render

import (
	"bytes"
	"testing"
)

func TestTerminalWidth_PipeFallback(t *testing.T) {
	t.Setenv("COLUMNS", "")

	w := &bytes.Buffer{}
	got := TerminalWidth(w)
	if got != 80 {
		t.Errorf("TerminalWidth(bytes.Buffer) = %d; want 80 (pipe fallback default)", got)
	}
}

func TestTerminalWidth_NilWriter(t *testing.T) {
	t.Setenv("COLUMNS", "")

	got := TerminalWidth(nil)
	if got != 80 {
		t.Errorf("TerminalWidth(nil) = %d; want 80", got)
	}
}

func TestTerminalWidth_COLUMNSEnvVar(t *testing.T) {
	t.Setenv("COLUMNS", "120")

	w := &bytes.Buffer{}
	got := TerminalWidth(w)
	if got != 120 {
		t.Errorf("TerminalWidth with COLUMNS=120: got %d; want 120", got)
	}
}

func TestTerminalWidth_COLUMNSEnvVar_200(t *testing.T) {
	t.Setenv("COLUMNS", "200")

	w := &bytes.Buffer{}
	got := TerminalWidth(w)
	if got != 200 {
		t.Errorf("TerminalWidth with COLUMNS=200: got %d; want 200", got)
	}
}

func TestTerminalWidth_COLUMNSInvalid_NotNumeric(t *testing.T) {
	t.Setenv("COLUMNS", "wide")

	w := &bytes.Buffer{}
	got := TerminalWidth(w)
	if got != 80 {
		t.Errorf("TerminalWidth with COLUMNS=wide (invalid): got %d; want 80 fallback", got)
	}
}

func TestTerminalWidth_COLUMNSInvalid_Zero(t *testing.T) {
	t.Setenv("COLUMNS", "0")

	w := &bytes.Buffer{}
	got := TerminalWidth(w)
	if got != 80 {
		t.Errorf("TerminalWidth with COLUMNS=0 (non-positive): got %d; want 80 fallback", got)
	}
}

func TestTerminalWidth_COLUMNSInvalid_Negative(t *testing.T) {
	t.Setenv("COLUMNS", "-1")

	w := &bytes.Buffer{}
	got := TerminalWidth(w)
	if got != 80 {
		t.Errorf("TerminalWidth with COLUMNS=-1 (negative): got %d; want 80 fallback", got)
	}
}
