//go:build !unix

package terminal

import (
	"errors"
	"image/color"
)

// ErrNotSupported is returned by QueryANSIPalette on platforms where OSC 4
// queries are not implemented (e.g. Windows).
var ErrNotSupported = errors.New("OSC 4 palette query not supported on this platform")

// QueryANSIPalette is a stub on non-Unix platforms.
func QueryANSIPalette() ([]color.RGBA, error) {
	return nil, ErrNotSupported
}
