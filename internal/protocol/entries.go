package protocol

import (
	"fmt"
	"io"
)

// Entries are the entries of a copy one after the other, each with its
// content, and io.EOF after the last.
type Entries interface {
	Next() (Entry, io.Reader, error)
}

// Next asks the host for the entries of the copy the first time, and then
// reads them one by one until it says all are sent.
func (c *Conn) Next() (Entry, io.Reader, error) {
	// nothing follows Sent, a read would wait forever
	if c.sent {
		return Entry{}, nil, io.EOF
	}

	if !c.asked {
		c.asked = true

		if err := c.Send(Send{}); err != nil {
			return Entry{}, nil, err
		}
	}

	message, err := c.Receive()
	if err != nil {
		return Entry{}, nil, err
	}

	switch message := message.(type) {
	case Entry:
		return message, c.content, nil
	case Sent:
		c.sent = true

		return Entry{}, nil, io.EOF
	}

	return Entry{}, nil, fmt.Errorf("%T is not an entry", message)
}
