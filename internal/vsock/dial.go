package vsock

import (
	"context"
	"io"
	"os"

	mdsocket "github.com/mdlayher/socket"
	"golang.org/x/sys/unix"
)

// Conn is a connection to a port of a machine.
type Conn interface {
	io.ReadWriteCloser

	// CloseWrite the other side reads to its end, and can still answer
	CloseWrite() error
}

// DialOn connects a socket that the maker makes elsewhere to a port of the
// machine with a context ID.
func DialOn(maker func() (*os.File, error), cid, port uint32) (Conn, error) {
	socket, err := maker()
	if err != nil {
		return nil, err
	}

	// the connection has a copy of the socket of its own
	defer func() { _ = socket.Close() }()

	conn, err := mdsocket.FileConn(socket, "vsock")
	if err != nil {
		return nil, err
	}

	if _, err := conn.Connect(context.Background(), &unix.SockaddrVM{CID: cid, Port: port}); err != nil {
		_ = conn.Close()

		return nil, err
	}

	return handed{conn}, nil
}

// handed is a connection dialed on a socket made elsewhere.
type handed struct {
	*mdsocket.Conn
}

func (h handed) CloseWrite() error {
	return h.Shutdown(unix.SHUT_WR)
}
