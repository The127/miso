package protocol

import (
	"fmt"
	"io"
)

// Runner does the work of a run on the agent's side.
type Runner interface {
	Run(run Run, out io.Writer) (code int, err error)
}

// Serve answers one request of the host with what the runner did.
func (c *Conn) Serve(runner Runner) error {
	message, err := c.Receive()
	if err != nil {
		return err
	}

	run, ok := message.(Run)
	if !ok {
		return c.Send(Failed{Reason: fmt.Sprintf("%T is not a request", message)})
	}

	code, err := runner.Run(run, outputs{c})
	if err != nil {
		return c.Send(Failed{Reason: err.Error()})
	}

	if code != 0 {
		return c.Send(Exited{Code: code})
	}

	return c.Send(Done{})
}

// outputs sends what is written to it to the host.
type outputs struct {
	conn *Conn
}

func (o outputs) Write(p []byte) (int, error) {
	if err := o.conn.Send(Output{Bytes: p}); err != nil {
		return 0, err
	}

	return len(p), nil
}
