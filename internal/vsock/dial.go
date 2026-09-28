package vsock

import (
	"io"

	mdvsock "github.com/mdlayher/vsock"
)

// Conn is a connection to a port of a machine.
type Conn interface {
	io.ReadWriteCloser

	// the other side reads to its end, and can still answer
	CloseWrite() error
}

// Dial connects to a port of the machine with a context ID.
func Dial(cid, port uint32) (Conn, error) {
	conn, err := mdvsock.Dial(cid, port, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
