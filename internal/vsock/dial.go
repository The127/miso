package vsock

import (
	"io"

	mdvsock "github.com/mdlayher/vsock"
)

// Dial connects to a port of the machine with a context ID.
func Dial(cid, port uint32) (io.ReadWriteCloser, error) {
	conn, err := mdvsock.Dial(cid, port, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
