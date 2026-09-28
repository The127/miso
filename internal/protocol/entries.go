package protocol

import (
	"errors"
	"fmt"
	"io"
)

// Entries are the entries of a copy one after the other, each with its
// content, and io.EOF after the last.
type Entries interface {
	Next() (Entry, io.Reader, error)
}

// reader reads the entries of one copy from the host.
type reader struct {
	conn *Conn

	// whether the entries of the copy were asked for
	asked bool

	// whether the host said all entries are sent
	sent bool
}

// Next asks the host for the entries of the copy the first time, and then
// reads them one by one until it says all are sent.
func (r *reader) Next() (Entry, io.Reader, error) {
	// nothing follows Sent, a read would wait forever
	if r.sent {
		return Entry{}, nil, io.EOF
	}

	if !r.asked {
		r.asked = true

		if err := r.conn.Send(Send{}); err != nil {
			return Entry{}, nil, err
		}
	}

	message, err := r.conn.Receive()
	if errors.Is(err, io.EOF) {
		// io.EOF would tell the runner that it has every entry
		return Entry{}, nil, fmt.Errorf("host gone before all entries were sent: %w", io.ErrUnexpectedEOF)
	}

	if err != nil {
		return Entry{}, nil, err
	}

	switch message := message.(type) {
	case Entry:
		return message, r.conn.content, nil
	case Sent:
		r.sent = true

		return Entry{}, nil, io.EOF
	}

	return Entry{}, nil, fmt.Errorf("%T is not an entry", message)
}
