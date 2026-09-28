package console

import "io"

// Console is a machine's serial console. It keeps what it read past the
// last match.
type Console struct {
	serial io.Reader
	seen   []byte
}

// New returns the console that the reader carries.
func New(serial io.Reader) *Console {
	return &Console{serial: serial}
}
