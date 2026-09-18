package protocol

import (
	"encoding/json"
	"io"
)

// Conn is the host's or the agent's end of one exchange.
type Conn struct {
	decoder *json.Decoder
	encoder *json.Encoder
}

// New reads what the other side sends from r and writes to it through w.
func New(r io.Reader, w io.Writer) *Conn {
	return &Conn{decoder: json.NewDecoder(r), encoder: json.NewEncoder(w)}
}

// Send writes a message to the other side.
func (c *Conn) Send(message Message) error {
	var e envelope
	switch m := message.(type) {
	case Run:
		e.Run = &m
	case Exited:
		e.Exited = &m
	}

	return c.encoder.Encode(e)
}

// Receive reads what the other side sent next.
func (c *Conn) Receive() (Message, error) {
	var e envelope
	if err := c.decoder.Decode(&e); err != nil {
		return nil, err
	}

	if e.Exited != nil {
		return *e.Exited, nil
	}

	return *e.Run, nil
}
