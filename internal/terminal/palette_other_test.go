//go:build !unix

package terminal_test

import (
	"errors"
	"testing"

	"github.com/tylercrawford/drift/internal/terminal"
)

func TestQueryANSIPalette_notSupported(t *testing.T) {
	_, err := terminal.QueryANSIPalette()
	if !errors.Is(err, terminal.ErrNotSupported) {
		t.Fatalf("got %v want ErrNotSupported", err)
	}
}
