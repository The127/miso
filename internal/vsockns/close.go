package vsockns

import "golang.org/x/sys/unix"

// Close ends the helper, and with it the namespace. Only the first Close
// does so, the number of the channel may belong to another file after it.
func (n *Namespace) Close() error {
	n.closing.Do(func() {
		_ = unix.Close(n.conn)
		n.closed = n.helper.Wait()
	})

	return n.closed
}
