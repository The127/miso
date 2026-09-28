package protocol

import "io"

// SendEntry writes an entry to the other side and its content right after
// it, as raw bytes.
func (c *Conn) SendEntry(entry Entry, content io.Reader) error {
	if err := c.Send(entry); err != nil {
		return err
	}

	_, err := io.CopyN(c.w, content, entry.Size)

	return err
}

// Content is what follows an entry just received.
func (c *Conn) Content(entry Entry) io.Reader {
	return io.LimitReader(c.r, entry.Size)
}
