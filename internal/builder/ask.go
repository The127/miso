package builder

import (
	"io"

	"github.com/The127/miso/internal/protocol"
)

// Ask asks the agent each request in order, on a connection of its own,
// because the agent answers one request per connection.
func Ask(dial func() (io.ReadWriteCloser, error), agent string, requests []protocol.Message) error {
	for _, request := range requests {
		conn, err := dial()
		if err != nil {
			return err
		}

		err = protocol.New(agent, conn, conn).Ask(request, io.Discard)
		_ = conn.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
