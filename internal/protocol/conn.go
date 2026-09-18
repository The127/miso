package protocol

import (
	"encoding/json"
	"fmt"
	"io"
)

// Conn is the host's or the agent's end of one exchange.
type Conn struct {
	agent   string
	decoder *json.Decoder
	encoder *json.Encoder
}

// New speaks as the given agent, reads what the other side sends from r and
// writes to it through w.
func New(agent string, r io.Reader, w io.Writer) *Conn {
	return &Conn{agent: agent, decoder: json.NewDecoder(r), encoder: json.NewEncoder(w)}
}

// Send writes a message to the other side.
func (c *Conn) Send(message Message) error {
	e := envelope{Agent: c.agent}
	message.into(&e)

	return c.encoder.Encode(e)
}

// Receive reads what the other side sent next.
func (c *Conn) Receive() (Message, error) {
	var e envelope
	if err := c.decoder.Decode(&e); err != nil {
		return nil, err
	}

	if e.Agent != c.agent {
		return nil, fmt.Errorf("%w: sent by %s, read by %s", ErrAnotherAgent, e.Agent, c.agent)
	}

	message := e.open()
	if message == nil {
		return nil, ErrUnknownMessage
	}

	return message, nil
}
