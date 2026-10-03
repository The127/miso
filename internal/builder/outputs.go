package builder

import (
	"errors"
	"io"
	"os"
	"path"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/protocol"
)

// Output is a file a disk is fetched into. Close delivers it, Discard
// throws away one whose fetch failed.
type Output interface {
	protocol.DiskFile
	Close() error
	Discard() error
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
	if err := conn.AskFetch(request, file, out); err != nil {
		return errors.Join(err, file.Discard())
	}

	return file.Close()
}

// OutputsIn are outputs that are files in a directory. A disk with checks
// is booted through check before it gets its name.
func OutputsIn(dir string, check Check) Outputs {
	return func(request build.Request) (Output, error) {
		name := request.Output
		if needsCheck(request) {
			name = unchecked(name)
		}

		var file *os.File

		err := inRoot(dir, func(root *os.Root) error {
			if err := root.MkdirAll(path.Dir(name), 0o750); err != nil {
				return err
			}

			// the holes of a disk are never written, so an older file under
			// the name would show through them
			var err error

			file, err = root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)

			return err
		})
		if err != nil {
			return nil, err
		}

		fetched := plain{File: file, dir: dir, name: name}

		if needsCheck(request) {
			return checked{plain: fetched, request: request, check: check}, nil
		}

		if request.Listed {
			return summed{fetched}, nil
		}

		return fetched, nil
	}
}

// plain is the file of a disk without checks.
type plain struct {
	*os.File

	dir  string
	name string
}

// Discard throws away what was fetched, which is no whole disk.
func (p plain) Discard() error {
	if err := p.Close(); err != nil {
		return err
	}

	return inRoot(p.dir, func(root *os.Root) error { return root.Remove(p.name) })
}
