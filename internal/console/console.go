package console

import "io"

// Console is a machine's serial console. It keeps what it read past the
// last match and how the console ended.
type Console struct {
	serial io.Reader
	seen   []byte
	end    error
}

// New returns the console that the reader carries.
func New(serial io.Reader) *Console {
	return &Console{serial: serial}
}
