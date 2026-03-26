//go:build unix

package terminal

import (
	"fmt"
	"image/color"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/tylercrawford/drift/internal/highlight"
)

// QueryANSIPalette queries the terminal for ANSI palette slots 0–15 via OSC 4,
// using /dev/tty for raw I/O. It returns a sparse palette slice suitable for
// highlight.BestMatchTheme. On failure or timeout (500ms), returns a non-nil error.
func QueryANSIPalette() ([]color.RGBA, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open /dev/tty: %w", err)
	}
	defer func() { _ = tty.Close() }()

	oldState, err := term.MakeRaw(int(tty.Fd()))
	if err != nil {
		return nil, fmt.Errorf("set raw mode: %w", err)
	}
	defer func() { _ = term.Restore(int(tty.Fd()), oldState) }()

	var query strings.Builder
	for n := 0; n < 16; n++ {
		fmt.Fprintf(&query, "\033]4;%d;?\007", n)
	}

	if _, err := tty.WriteString(query.String()); err != nil {
		return nil, fmt.Errorf("write OSC 4 query: %w", err)
	}

	ch := make(chan []byte, 1)
	go func() {
		var buf []byte
		tmp := make([]byte, 4096)
		for {
			n, err := tty.Read(tmp)
			if n > 0 {
				buf = append(buf, tmp[:n]...)
			}
			if err != nil || n == 0 {
				break
			}
			if strings.Count(string(buf), "\007") >= 16 {
				break
			}
		}
		ch <- buf
	}()

	select {
	case raw := <-ch:
		if len(raw) == 0 {
			return nil, fmt.Errorf("no OSC 4 response")
		}
		palette := highlight.ParseOSC4Responses(raw)
		if palette == nil {
			return nil, fmt.Errorf("no valid OSC 4 palette in response")
		}
		return palette, nil
	case <-time.After(500 * time.Millisecond):
		return nil, fmt.Errorf("OSC 4 query timed out")
	}
}
