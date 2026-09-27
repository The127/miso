package builder

import (
	"errors"
	"io"
	"time"

	"github.com/The127/miso/internal/protocol"
)

// redial is how long Ask waits before it dials an agent again that does not
// listen yet.
const redial = 100 * time.Millisecond

// settle is how long Ask waits for the VM to count as stopped after a step
// failed, so the error can say why the VM stopped.
const settle = time.Second

// Ask asks the agent each request in order, on a connection of its own,
// because the agent answers one request per connection. What the agent
// writes goes to out.
func Ask(vm VM, dial func() (io.ReadWriteCloser, error), agent string, requests []protocol.Message, out io.Writer) error {
	for _, request := range requests {
		conn, err := connect(vm, dial)
		if err != nil {
			return err
		}

		err = protocol.New(agent, conn, conn).Ask(request, out)
		_ = conn.Close()
		if err != nil {
			return failed(vm, err)
		}
	}

	return nil
}

// connect dials until the agent listens, which it does only once its VM has
// booted, or until the VM stops.
func connect(vm VM, dial func() (io.ReadWriteCloser, error)) (io.ReadWriteCloser, error) {
	for {
		conn, err := dial()
		if err == nil {
			return conn, nil
		}

		select {
		case <-vm.Done():
			return nil, stopped(vm, "before its agent listened")
		case <-time.After(redial):
		}
	}
}

// failed is why a step failed: what the agent answered, or else why its VM
// stopped.
func failed(vm VM, err error) error {
	// the agent answered, so the VM still runs
	if errors.Is(err, protocol.ErrCommandFailed) || errors.Is(err, protocol.ErrAgentFailed) {
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
