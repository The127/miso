package protocol

import (
	"errors"
	"io"
)

// ErrNegativeSize is an entry or piece that says less than nothing follows it.
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

// Content is what follows the entry or piece received last. The next
// Receive skips what is left of it.
func (c *Conn) Content() io.Reader {
	return c.content
}

// sized reads the content of one entry. Unlike io.LimitReader it does not
// end quietly when the wire ends early, which would pass a cut file as
// whole.
type sized struct {
	r    io.Reader
	left int64
}

func (s *sized) Read(p []byte) (int, error) {
	if s.left == 0 {
		return 0, io.EOF
	}

	if int64(len(p)) > s.left {
		p = p[:s.left]
	}

	n, err := s.r.Read(p)
	s.left -= int64(n)

	if errors.Is(err, io.EOF) && s.left > 0 {
		return n, io.ErrUnexpectedEOF
	}

	return n, err
}
