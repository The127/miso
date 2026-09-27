package builder

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/The127/miso/internal/protocol"
)

// redial is how long Ask waits before it dials an agent again that does not
// listen yet.
const redial = 100 * time.Millisecond

// Ask asks the agent each request in order, on a connection of its own,
// because the agent answers one request per connection. What the agent
// writes goes to out. It dials until the agent listens, which it does
// only once its VM has booted, or until the VM stops.
func Ask(vm VM, dial func() (io.ReadWriteCloser, error), agent string, requests []protocol.Message, out io.Writer) error {
	for _, request := range requests {
		conn, err := dial()
		for err != nil {
			select {
			case <-vm.Done():
				if vm.Err() == nil {
					return errors.New("the builder VM stopped before its agent listened")
				}

				return fmt.Errorf("the builder VM stopped before its agent listened: %w", vm.Err())
			case <-time.After(redial):
			}

			conn, err = dial()
		}

		err = protocol.New(agent, conn, conn).Ask(request, out)
		_ = conn.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
