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

// outputOf is the file a fetch request is written into, or none when the
// disk stays in the cache, as a build without an output target does in
// Docker.
func outputOf(request build.Request, outputs Outputs) (Output, error) {
	if outputs == nil {
		return nil, nil
	}

	return outputs(request)
}

// fetch has the agent send the disk of a request into the file of its
// output.
func fetch(conn *protocol.Conn, request protocol.Fetch, file Output, out io.Writer) error {
	return errors.Join(conn.AskFetch(request, file, out), file.Close())
}

// OutputsIn are outputs that are files in a directory. A disk with checks
// is booted through check before it gets its name.
func OutputsIn(dir string, check Check) Outputs {
	return func(request build.Request) (Output, error) {
		root, err := os.OpenRoot(dir)
		if err != nil {
			return nil, err
		}

		defer func() { _ = root.Close() }()

		name := request.Output
		if len(request.Checks) > 0 {
			name = unchecked(name)
		}

		// the holes of a disk are never written, so an older file under the
		// name would show through them
		file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return nil, err
		}

		if len(request.Checks) > 0 {
			return checked{File: file, dir: dir, name: name, request: request, check: check}, nil
		}

		return file, nil
	}
}
