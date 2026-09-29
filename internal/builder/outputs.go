package builder

import (
	"errors"
	"io"
	"os"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/protocol"
)

// Output is a file a disk is fetched into.
type Output interface {
	protocol.DiskFile
	Close() error
}

// Outputs create the file of the output a fetch request names, handed in so
// the caller decides where outputs go.
type Outputs func(request build.Request) (Output, error)

// fetch has the agent send the disk of a request into the file of its
// output.
func fetch(conn *protocol.Conn, request build.Request, fetching protocol.Fetch, outputs Outputs, out io.Writer) error {
	file, err := outputs(request)
	if err != nil {
		return err
	}

	return errors.Join(conn.AskFetch(fetching, file, out), file.Close())
}

// OutputsIn are outputs that are files in a directory.
func OutputsIn(dir string) Outputs {
	return func(request build.Request) (Output, error) {
		root, err := os.OpenRoot(dir)
		if err != nil {
			return nil, err
		}

		defer func() { _ = root.Close() }()

		// the holes of a disk are never written, so an older file under the
		// name would show through them
		file, err := root.OpenFile(request.Output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return nil, err
		}

		return file, nil
	}
}
