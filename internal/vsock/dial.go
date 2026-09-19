package vsock

import (
	"os"

	"golang.org/x/sys/unix"
)

// Dial connects to a port of the machine with a context ID.
func Dial(cid, port uint32) (*os.File, error) {
	fd, err := unix.Socket(unix.AF_VSOCK, unix.SOCK_STREAM, 0)
	if err != nil {
		return nil, err
	}

	if err := unix.Connect(fd, &unix.SockaddrVM{CID: cid, Port: port}); err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(fd), "vsock"), nil
}
