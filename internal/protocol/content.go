package protocol

import (
	"errors"
	"io"
)

// ErrNegativeSize is an entry that says less than nothing follows it.
var ErrNegativeSize = errors.New("negative size")

// SendEntry writes an entry to the other side and its content right after
// it, as raw bytes.
func (c *Conn) SendEntry(entry Entry, content io.Reader) error {
	if err := c.Send(entry); err != nil {
		return err
	}

	_, err := io.CopyN(c.w, content, entry.Size)

	return err
}

// Content is what follows the entry received last. The next Receive skips
// what is left of it.
func (c *Conn) Content() io.Reader {
	return c.content
}
