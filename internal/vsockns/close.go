package vsockns

import (
	"errors"

	"golang.org/x/sys/unix"
)

// ErrClosed is a question to a namespace after its Close.
var ErrClosed = errors.New("the vsock namespace is closed")

// Close ends the helper, and with it the namespace. Only the first Close
// does so, the number of the channel may belong to another file after it.
func (n *Namespace) Close() error {
	n.closing.Do(func() {
		// a question still waiting for its answer would keep the channel
		// from Close
		_ = unix.Shutdown(n.conn, unix.SHUT_RDWR)

		n.asking.Lock()
		n.gone = true
		_ = unix.Close(n.conn)
		n.asking.Unlock()

		n.closed = n.helper.Wait()
	})

	return n.closed
}
