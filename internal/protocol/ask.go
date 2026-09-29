package protocol

import (
	"errors"
	"fmt"
	"io"
)

// ErrNoAnswer is a message the agent sent that does not answer the request.
var ErrNoAnswer = errors.New("no answer")

// Ask has the agent carry out a request and writes what the request writes
// to out until the agent tells how it ended.
func (c *Conn) Ask(request Message, out io.Writer) error {
	return c.ask(request, nil, nil, out)
}

func (c *Conn) ask(request Message, files Files, disk DiskFile, out io.Writer) error {
	if err := c.Send(request); err != nil {
		return err
	}

	for {
		message, err := c.Receive()
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("the agent stopped before it answered: %w", io.ErrUnexpectedEOF)
		}

		if err != nil {
			return err
		}

		switch m := message.(type) {
		case Output:
			if _, err := out.Write(m.Bytes); err != nil {
				return err
			}
		case Send:
			if files == nil {
				return noAnswer(m, request)
			}

			if err := files(c.SendEntry); err != nil {
				return err
			}

			if err := c.Send(Sent{}); err != nil {
				return err
			}
		case Piece:
			if disk == nil {
				return noAnswer(m, request)
			}

			if _, err := io.Copy(io.NewOffsetWriter(disk, m.Offset), c.Content()); err != nil {
				return err
			}
		case Length:
			if err := disk.Truncate(m.Size); err != nil {
				return err
			}
		case Done:
			return nil
		case Exited:
			return fmt.Errorf("%w: exit code %d", ErrCommandFailed, m.Code)
		case Failed:
			return fmt.Errorf("%w: %s", ErrAgentFailed, m.Reason)
		default:
			return noAnswer(m, request)
		}
	}
}

// noAnswer is the error for a message the request does not expect.
func noAnswer(message, request Message) error {
	return fmt.Errorf("%T is %w to %T", message, ErrNoAnswer, request)
}
