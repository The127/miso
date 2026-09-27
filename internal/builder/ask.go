package builder

import (
	"io"

	"github.com/The127/miso/internal/protocol"
)

// Ask asks the agent each request in order, on a connection of its own,
// because the agent answers one request per connection. What the agent
// writes goes to out.
func Ask(dial func() (io.ReadWriteCloser, error), agent string, requests []protocol.Message, out io.Writer) error {
	for _, request := range requests {
		conn, err := dial()
		if err != nil {
			return err
		}

		err = protocol.New(agent, conn, conn).Ask(request, out)
		_ = conn.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
