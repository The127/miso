package builder

import (
	"errors"
	"io"

	"github.com/The127/miso/internal/protocol"
)

// Output is a file a disk is fetched into.
type Output interface {
	protocol.DiskFile
	Close() error
}

// Outputs create the file of an output under its name, handed in so the
// caller decides where outputs go.
type Outputs func(name string) (Output, error)

// fetch has the agent send the disk of a request into the file of its
// output.
func fetch(conn *protocol.Conn, request protocol.Fetch, name string, outputs Outputs, out io.Writer) error {
	file, err := outputs(name)
	if err != nil {
		return err
	}

	return errors.Join(conn.AskFetch(request, file, out), file.Close())
}
