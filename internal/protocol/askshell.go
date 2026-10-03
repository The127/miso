package protocol

import (
	"errors"
	"io"
)

// inputPiece is the most bytes one Input carries.
const inputPiece = 4096

// AskShell has the agent run an interactive shell. It sends the shell what
// the terminal's In reads and the sizes the terminal changes to, writes what
// the shell prints to out until the shell ends, and answers the code it
// exited with. Reading In goes on until it ends, so the caller closes it,
// and the connection serves no other request afterwards.
func (c *Conn) AskShell(shell Shell, term Terminal, out io.Writer) (int, error) {
	if err := c.Send(shell); err != nil {
		return 0, err
	}

	go func() {
		piece := make([]byte, inputPiece)
		for {
			n, err := term.In.Read(piece)
			if n > 0 && c.Send(Input{Bytes: piece[:n]}) != nil {
				return
			}

			if err != nil {
				return
			}
		}
	}()

	// the read of In cannot be stopped, so only this loop can be told to end
	over := make(chan struct{})
	defer close(over)

	go func() {
		for {
			select {
			case size := <-term.Resized:
				if c.Send(size) != nil {
					return
				}
			case <-over:
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
