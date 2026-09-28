package protocol

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// MaxMessage is the most bytes a message may take on the wire. Bulk data
// goes as raw bytes after a message, never inside one.
const MaxMessage = 1 << 20

// ErrMessageTooLong is a message that would take more than MaxMessage.
var ErrMessageTooLong = errors.New("message too long")

// Conn is the host's or the agent's end of one exchange.
type Conn struct {
	agent string
	r     io.Reader
	w     io.Writer

	// what follows the entry received last, as far as it is not read yet
	content io.Reader

	// whether the entries of the copy were asked for
	asked bool

	// whether the host said all entries are sent
	sent bool
}

// New speaks as the given agent, reads what the other side sends from r and
// writes to it through w.
func New(agent string, r io.Reader, w io.Writer) *Conn {
	return &Conn{agent: agent, r: r, w: w}
}

// Send writes a message to the other side, its length first.
func (c *Conn) Send(message Message) error {
	e := envelope{Agent: c.agent}
	message.into(&e)

	body, err := json.Marshal(e)
	if err != nil {
		return err
	}

	size := len(body)
	if size > MaxMessage {
		return fmt.Errorf("%w: %d bytes", ErrMessageTooLong, size)
	}

	// one write, so a message never reaches the other side in pieces of two
	// writers
	frame := binary.BigEndian.AppendUint32(nil, uint32(size))
	_, err = c.w.Write(append(frame, body...))

	return err
}

// Receive reads what the other side sent next, and not a byte further, so
// that raw bytes may follow a message.
func (c *Conn) Receive() (Message, error) {
	// content left unread would be taken for the next message
	if c.content != nil {
		if _, err := io.Copy(io.Discard, c.content); err != nil {
			return nil, err
		}

		c.content = nil
	}

	var length [4]byte
	if _, err := io.ReadFull(c.r, length[:]); err != nil {
		return nil, err
	}

	// a broken length must not have us allocate up to 4 GiB
	size := binary.BigEndian.Uint32(length[:])
	if size > MaxMessage {
		return nil, fmt.Errorf("%w: %d bytes", ErrMessageTooLong, size)
	}

	body := make([]byte, size)
	if _, err := io.ReadFull(c.r, body); err != nil {
		return nil, err
	}

	var e envelope
	if err := json.Unmarshal(body, &e); err != nil {
		return nil, err
	}

	if e.Agent != c.agent {
		return nil, fmt.Errorf("%w: sent by %s, read by %s", ErrAnotherAgent, e.Agent, c.agent)
	}

	message := e.open()
	if message == nil {
		return nil, ErrUnknownMessage
	}

	if entry, isEntry := message.(Entry); isEntry {
		// the content would be read as nothing, and the stream no longer
		// say where the next message starts
		if entry.Size < 0 {
			return nil, fmt.Errorf("%s: %w: %d", entry.Path, ErrNegativeSize, entry.Size)
		}

		c.content = io.LimitReader(c.r, entry.Size)
	}

	return message, nil
}
