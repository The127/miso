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

// Send writes a run to the other side.
func (c *Conn) Send(run Run) error {
	return c.encoder.Encode(run)
}

// Receive reads what the other side sent next.
func (c *Conn) Receive() (Run, error) {
	var run Run
	err := c.decoder.Decode(&run)

	return run, err
}
