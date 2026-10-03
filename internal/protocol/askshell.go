package protocol

import (
	"errors"
	"io"
)

// inputPiece is the most bytes one Input carries.
const inputPiece = 4096

// AskShell has the agent run an interactive shell, sends it what in reads
// and writes what the shell prints to out until the shell ends, and answers
// the code it exited with. Reading in goes on until it ends, so the caller
// closes it, and the connection serves no other request afterwards.
func (c *Conn) AskShell(shell Shell, in io.Reader, out io.Writer) (int, error) {
	if err := c.Send(shell); err != nil {
		return 0, err
	}

	go func() {
		piece := make([]byte, inputPiece)
		for {
			n, err := in.Read(piece)
			if n > 0 && c.Send(Input{Bytes: piece[:n]}) != nil {
				return
			}

			if err != nil {
				return
			}
		}
	}()

	err := c.answered(shell, nil, nil, out)
	if exited, ok := errors.AsType[exitError](err); ok {
		return exited.code, nil
	}

	return 0, err
}
