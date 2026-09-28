package console

import (
	"io"
	"sync"
)

// Console is a machine's serial console. It keeps what it read past the
// last match and how the console ended. It serves one caller at a time.
type Console struct {
	mu   sync.Mutex
	seen []byte
	end  error
	more chan struct{}
}

// New returns the console that the reader carries. It reads until the
// reader fails or ends.
func New(serial io.Reader) *Console {
	c := &Console{more: make(chan struct{}, 1)}
	go c.listen(serial)

	return c
}
