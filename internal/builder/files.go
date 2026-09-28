package builder

import (
	"io"

	"github.com/The127/miso/internal/protocol"
)

// Files are the files a copy carries from the build context, handed in
// because the builder reads no files of the host.
type Files func(protocol.Copy) protocol.Files

// hostError is a copy's files failing on the host. The VM plays no part in
// it, so there is no reason of the VM to wait for.
type hostError struct {
	err error
}

func (h hostError) Error() string { return h.err.Error() }

func (h hostError) Unwrap() error { return h.err }

// hosted are the files with every error that is not sending them marked as
// the host's own.
func hosted(files protocol.Files) protocol.Files {
	return func(send func(protocol.Entry, io.Reader) error) error {
		var sendErr error
		err := files(func(entry protocol.Entry, content io.Reader) error {
			sendErr = send(entry, content)

			return sendErr
		})
		if err != nil && sendErr == nil {
			return hostError{err}
		}

		return err
	}
}
