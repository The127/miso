package protocol

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
)

// Runner does the work the host asks for on the agent's side.
type Runner interface {
	Run(ctx context.Context, run Run, out io.Writer) (code int, err error)
	Import(ctx context.Context, request Import, out io.Writer) error
	Copy(ctx context.Context, request Copy, entries Entries, out io.Writer) error
}

// Serve answers one request of the host with what the runner did.
func (c *Conn) Serve(runner Runner) error {
	message, err := c.Receive()
	if errors.Is(err, ErrAnotherAgent) {
		return c.Send(Failed{Reason: err.Error()})
	}

	if err != nil {
		return err
	}

	run, isRun := message.(Run)
	request, isImport := message.(Import)
	copying, isCopy := message.(Copy)
	if !isRun && !isImport && !isCopy {
		return c.Send(Failed{Reason: fmt.Sprintf("%T is not a request", message)})
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// the host sends nothing more in an exchange, so whatever ends this read
	// is the host going away
	go func() {
		_, _ = c.Receive()

		cancel()
	}()

	if isImport {
		if err := runner.Import(ctx, request, outputs{c}); err != nil {
			return c.Send(Failed{Reason: err.Error()})
		}

		return c.Send(Done{})
	}

	if isCopy {
		if err := runner.Copy(ctx, copying, nil, outputs{c}); err != nil {
			return c.Send(Failed{Reason: err.Error()})
		}

		return c.Send(Done{})
	}

	code, err := runner.Run(ctx, run, outputs{c})
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

// outputPiece is the most bytes one Output carries. JSON grows them by a
// third, and the message must stay under MaxMessage.
const outputPiece = MaxMessage / 4

func (o outputs) Write(p []byte) (int, error) {
	sent := 0
	for piece := range slices.Chunk(p, outputPiece) {
		if err := o.conn.Send(Output{Bytes: piece}); err != nil {
			return sent, err
		}

		sent += len(piece)
	}

	return sent, nil
}
