package builder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/protocol"
)

// redial is how long Ask waits before it dials an agent again that does not
// listen yet.
const redial = 100 * time.Millisecond

// settle is how long Ask waits for the VM to count as stopped after a step
// failed, so the error can say why the VM stopped.
const settle = time.Second

// patience is how long Ask waits for the agent to listen at first, far more
// than a builder VM takes to boot, for one that hangs without stopping.
var patience = time.Minute

// slower is how many times longer a VM without KVM takes to boot, at most.
const slower = 10

// patienceFor is how long Ask waits for the agent of the VM to listen.
func patienceFor(vm VM) time.Duration {
	if vm.WithoutKVM() {
		return slower * patience
	}

	return patience
}

// Ask asks the agent each request in order, on a connection of its own,
// because the agent answers one request per connection. What the agent
// writes goes to out, a copy takes what it carries from files and a fetch
// writes into the file outputs creates for it. Once ctx is done the running
// step is cancelled.
func Ask(ctx context.Context, vm VM, dial Dial, agent string, requests []build.Request, files Files, outputs Outputs, out io.Writer) error {
	booted := time.After(patienceFor(vm))
	for _, request := range requests {
		var file Output
		if _, isFetch := request.Message.(protocol.Fetch); isFetch {
			var err error
			file, err = outputOf(request, outputs)
			if err != nil {
				return request.Failed(err)
			}

			if file == nil {
				continue
			}
		}

		conn, err := connect(ctx, vm, dial, booted)
		if err != nil && file != nil {
			return errors.Join(err, file.Discard())
		}

		if err != nil {
			return err
		}

		// once the agent listened, a step may take as long as it takes
		booted = nil

		err = ask(ctx, conn, agent, request, files, file, out)
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err != nil {
			return request.Failed(failed(vm, err))
		}
	}

	return nil
}

// ask asks the agent one request on a connection and closes it. Once ctx is
// done the connection closes early, which cancels the step in the agent.
func ask(ctx context.Context, conn io.ReadWriteCloser, agent string, request build.Request, files Files, file Output, out io.Writer) error {
	closeOnCancel := context.AfterFunc(ctx, func() { _ = conn.Close() })

	var err error
	copying, isCopy := request.Message.(protocol.Copy)
	fetching, isFetch := request.Message.(protocol.Fetch)
	switch {
	// a stage's files are in the builder, so the host has none to send
	case isCopy && copying.Stage == "":
		err = protocol.New(agent, conn, conn).AskCopy(copying, hosted(files(copying)), out)
	case isFetch:
		err = fetch(protocol.New(agent, conn, conn), fetching, file, out)
	default:
		err = protocol.New(agent, conn, conn).Ask(request.Message, out)
	}

	closeOnCancel()
	_ = conn.Close()

	return err
}

// connect dials until the agent listens, which it does only once its VM has
// booted, or until the VM stops, ctx is done or the VM took too long to boot.
func connect(ctx context.Context, vm VM, dial Dial, booted <-chan time.Time) (io.ReadWriteCloser, error) {
	for {
		conn, err := dial()
		if err == nil {
			return conn, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-vm.Done():
			// a cancel kills QEMU too, and then the cancel is what to tell
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			return nil, stopped(vm, "before its agent listened")
		case <-booted:
			return nil, fmt.Errorf("the agent did not listen within %s", patienceFor(vm))
		case <-time.After(redial):
		}
	}
}

// failed is why a step failed: what the agent answered, or else why its VM
// stopped.
func failed(vm VM, err error) error {
	// the agent sent a message, so the VM still runs
	if errors.Is(err, protocol.ErrCommandFailed) || errors.Is(err, protocol.ErrAgentFailed) || errors.Is(err, protocol.ErrNoAnswer) {
		return err
	}

	if errors.As(err, &hostError{}) {
		return err
	}

	// a killed QEMU breaks the connection a moment before it counts as
	// stopped
	select {
	case <-vm.Done():
		return stopped(vm, "during a step")
	case <-time.After(settle):
		return err
	}
}
