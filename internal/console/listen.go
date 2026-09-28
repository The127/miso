package console

import "io"

// listen reads all the time, not only while someone expects, so a machine
// writing to its console is never held up by a full pipe.
func (c *Console) listen(screen io.Reader) {
	chunk := make([]byte, 4096)
	for {
		n, err := screen.Read(chunk)

		c.mu.Lock()
		c.seen = append(c.seen, chunk[:n]...)
		c.end = err
		c.mu.Unlock()

		select {
		case c.more <- struct{}{}:
		default:
		}

		if err != nil {
			return
		}
	}
}
