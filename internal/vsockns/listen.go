package vsockns

import (
	"os"

	"golang.org/x/sys/unix"
)

// Listen is a vsock socket inside the namespace, listening on a port of
// it. It stays in the namespace wherever it is used.
func (n *Namespace) Listen(port uint32) (*os.File, error) {
	socket, err := n.Socket()
	if err != nil {
		return nil, err
	}

	raw, err := socket.SyscallConn()
	if err != nil {
		_ = socket.Close()

		return nil, err
	}

	var failed error
	err = raw.Control(func(fd uintptr) {
		failed = unix.Bind(int(fd), &unix.SockaddrVM{CID: unix.VMADDR_CID_ANY, Port: port})
		if failed == nil {
			failed = unix.Listen(int(fd), unix.SOMAXCONN)
		}
	})
	if err == nil {
		err = failed
	}

	if err != nil {
		_ = socket.Close()

		return nil, err
	}

	return socket, nil
}
