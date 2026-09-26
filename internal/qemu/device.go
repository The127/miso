package qemu

import (
	"fmt"
	"os"
)

// openVsock opens the host's vhost-vsock device, on which the host claims a
// CID and which QEMU then gets to keep it.
func openVsock(path string) (*os.File, error) {
	device, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("the host's vsock device needs the vhost_vsock module and access for this user: %w", err)
	}

	return device, nil
}
