package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpListsAllFlags(t *testing.T) {
	t.Helper()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, flag := range []string{
		"--split",
		"--algorithm",
		"--lang",
		"--theme",
		"--no-color",
		"--context",
		"--from",
		"--to",
	} {
		if !strings.Contains(out, flag) {
			t.Errorf("help output missing flag %q\noutput:\n%s", flag, out)
		}
	}
}
