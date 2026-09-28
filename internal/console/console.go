package console

import (
	"io"
	"sync"
)

// Console is a machine's serial console. It reads what the machine shows
// and types on its keyboard. It keeps what it read past the last match and
// how the console ended. It serves one caller at a time.
type Console struct {
	mu   sync.Mutex
	seen []byte
	end  error
	more chan struct{}

	keyboard io.Writer
}

// New returns the console whose output the reader carries and whose input
// the writer takes. It reads until the reader fails or ends.
func New(screen io.Reader, keyboard io.Writer) *Console {
	c := &Console{more: make(chan struct{}, 1), keyboard: keyboard}
	go c.listen(screen)

	return c
}
