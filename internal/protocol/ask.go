package protocol

import (
	"errors"
	"fmt"
	"io"
)

// Ask has the agent carry out a request and writes what the request writes
// to out until the agent tells how it ended.
func (c *Conn) Ask(request Message, out io.Writer) error {
	if err := c.Send(request); err != nil {
		return err
	}

	for {
		message, err := c.Receive()
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("the agent stopped before the step ended: %w", io.ErrUnexpectedEOF)
		}

		if err != nil {
			return err
		}

		switch m := message.(type) {
		case Output:
			if _, err := out.Write(m.Bytes); err != nil {
				return err
			}
		case Done:
			return nil
		case Exited:
			return fmt.Errorf("%w: exit code %d", ErrCommandFailed, m.Code)
		case Failed:
			return fmt.Errorf("%w: %s", ErrAgentFailed, m.Reason)
		}
	}
}
