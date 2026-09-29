package check

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/vsock"
)

// ErrNoNamespace is a boot without a vsock namespace, whose VM would share
// the host's vsock with everything else on it.
var ErrNoNamespace = errors.New("a boot needs a vsock namespace")

// Namespace keeps the vsock of a boot to miso: the VM runs on its device and
// the checks talk over its sockets.
type Namespace interface {
	Device() (*os.File, error)
	Socket() (*os.File, error)
	Listen(port uint32) (*os.File, error)
}

// listen takes the notices of the boot on any port.
func (b Boot) listen() (*vsock.Listener, error) {
	socket, err := b.Namespace.Listen(unix.VMADDR_PORT_ANY)
	if err != nil {
		return nil, err
	}

	defer func() { _ = socket.Close() }()

	return vsock.Listening(socket)
}

// dial connects to a port of the VM.
func (b Boot) dial(cid, port uint32) (vsock.Conn, error) {
	socket, err := b.Namespace.Socket()
	if err != nil {
		return nil, err
	}

	defer func() { _ = socket.Close() }()

	return vsock.DialOn(socket, cid, port)
}
