package vsock

import (
	"io"

	mdvsock "github.com/mdlayher/vsock"
	"golang.org/x/sys/unix"
)

// Listener takes the connections made to a port of this machine.
type Listener struct {
	listener *mdvsock.Listener
}

// Listen takes the connections made to a port of this machine.
func Listen(port uint32) (*Listener, error) {
	// any context ID, so a connection arrives whichever the machine has
	listener, err := mdvsock.ListenContextID(unix.VMADDR_CID_ANY, port, nil)
	if err != nil {
		return nil, err
	}

	return &Listener{listener: listener}, nil
}

// Accept waits for the next connection.
func (l *Listener) Accept() (io.ReadWriteCloser, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Port is the port the listener takes connections on, the one the system
// chose for a listener on any port.
func (l *Listener) Port() uint32 {
	// a vsock listener's address is always a vsock address
	return l.listener.Addr().(*mdvsock.Addr).Port
}
