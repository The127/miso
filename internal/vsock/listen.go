package vsock

import (
	"io"
	"os"

	"golang.org/x/sys/unix"
)

// Listener takes the connections made to a port of this machine.
type Listener struct {
	fd int
}

// Listen takes the connections made to a port of this machine.
func Listen(port uint32) (*Listener, error) {
	fd, err := unix.Socket(unix.AF_VSOCK, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}

	if err := unix.Bind(fd, &unix.SockaddrVM{CID: unix.VMADDR_CID_ANY, Port: port}); err != nil {
		return nil, err
	}

	if err := unix.Listen(fd, unix.SOMAXCONN); err != nil {
		return nil, err
	}

	return &Listener{fd: fd}, nil
}

// Accept waits for the next connection.
func (l *Listener) Accept() (io.ReadWriteCloser, error) {
	fd, _, err := unix.Accept4(l.fd, unix.SOCK_CLOEXEC)
	if err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(fd), "vsock"), nil
}
