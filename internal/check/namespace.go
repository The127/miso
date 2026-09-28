package check

import (
	"os"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/vsock"
)

// Namespace keeps the vsock of a boot to miso: the VM runs on its device and
// the checks talk over its sockets.
type Namespace interface {
	Device() (*os.File, error)
	Socket() (*os.File, error)
	Listen(port uint32) (*os.File, error)
}

// listen takes the notices of the boot on any port.
func (b Boot) listen() (*vsock.Listener, error) {
	if b.Namespace == nil {
		return vsock.Listen(unix.VMADDR_PORT_ANY)
	}

	socket, err := b.Namespace.Listen(unix.VMADDR_PORT_ANY)
	if err != nil {
		return nil, err
	}

	defer func() { _ = socket.Close() }()

	return vsock.Listening(socket)
}

// dial connects to a port of the VM.
func (b Boot) dial(cid, port uint32) (vsock.Conn, error) {
	if b.Namespace == nil {
		return vsock.Dial(cid, port)
	}

	socket, err := b.Namespace.Socket()
	if err != nil {
		return nil, err
	}

	defer func() { _ = socket.Close() }()

	return vsock.DialOn(socket, cid, port)
}
