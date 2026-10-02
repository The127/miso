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
	Disk(ctx context.Context, request Disk, out io.Writer) error
	Rootfs(ctx context.Context, request Rootfs, out io.Writer) error
	BootPart(ctx context.Context, request BootPart, out io.Writer) error
	Fetch(ctx context.Context, request Fetch, pieces Pieces, out io.Writer) error
	Prune(ctx context.Context, request Prune, out io.Writer) error
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

	switch request := message.(type) {
	case Copy:
		return c.serveCopy(runner, request)
	case Disk:
		return c.serveDisk(runner, request)
	case Fetch:
		return c.serveFetch(runner, request)
	case Import:
		return c.serveImport(runner, request)
	case Prune:
		return c.servePrune(runner, request)
	case Rootfs:
		return c.serveRootfs(runner, request)
	case BootPart:
		return c.serveBootPart(runner, request)
	case Run:
		return c.serveRun(runner, request)
	default:
		return c.Send(Failed{Reason: fmt.Sprintf("%T is not a request", message)})
	}
}

// serveCopy has the runner copy what the host sends. A copy reads its
// entries from the host, so no one else may read.
func (c *Conn) serveCopy(runner Runner, request Copy) error {
	entries := &reader{conn: c}
	err := runner.Copy(context.Background(), request, entries, outputs{c})
	if err != nil {
		// the host sends every entry before it reads an answer, so it
		// would wait on us while we wait on it. Unasked, it sends none
		for entries.asked {
			if _, _, err := entries.Next(); err != nil {
				break
			}
		}
	}

	return c.answer(err)
}

func (c *Conn) serveDisk(runner Runner, request Disk) error {
	return c.watching(func(ctx context.Context) error {
		return runner.Disk(ctx, request, outputs{c})
	})
}

func (c *Conn) serveBootPart(runner Runner, request BootPart) error {
	return c.watching(func(ctx context.Context) error {
		return runner.BootPart(ctx, request, outputs{c})
	})
}

func (c *Conn) serveFetch(runner Runner, request Fetch) error {
	return c.watching(func(ctx context.Context) error {
		return runner.Fetch(ctx, request, sending{c}, outputs{c})
	})
}

func (c *Conn) serveImport(runner Runner, request Import) error {
	return c.watching(func(ctx context.Context) error {
		return runner.Import(ctx, request, outputs{c})
	})
}

func (c *Conn) servePrune(runner Runner, request Prune) error {
	return c.watching(func(ctx context.Context) error {
		return runner.Prune(ctx, request, outputs{c})
	})
}

func (c *Conn) serveRootfs(runner Runner, request Rootfs) error {
	return c.watching(func(ctx context.Context) error {
		return runner.Rootfs(ctx, request, outputs{c})
	})
}

func (c *Conn) serveRun(runner Runner, request Run) error {
	ctx, cancel := c.watched()
	defer cancel()

	code, err := runner.Run(ctx, request, outputs{c})
	if err == nil && code != 0 {
		return c.Send(Exited{Code: code})
	}

	return c.answer(err)
}

// watching does the work of a request until the host goes away, and tells
// the host how it ended.
func (c *Conn) watching(work func(ctx context.Context) error) error {
	ctx, cancel := c.watched()
	defer cancel()

	return c.answer(work(ctx))
}

// watched is cancelled once the host goes away. The host sends nothing
// more in an exchange, so whatever ends this read is the host leaving.
func (c *Conn) watched() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		_, _ = c.Receive()

		cancel()
	}()

	return ctx, cancel
}

// answer tells the host how the request ended.
func (c *Conn) answer(err error) error {
	if err != nil {
		return c.Send(Failed{Reason: err.Error()})
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
