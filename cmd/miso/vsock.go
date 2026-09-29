package main

import (
	"errors"
	"os"

	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/vsockns"
)

// privateVsock hands the driver the vsock device of a namespace of miso's
// own, and gives the sockets the host dials the VM on and how to end the
// namespace. On a host that cannot keep vsock private the driver is handed
// why instead, and reaches the VM over its virtio port.
func privateVsock(driver *qemu.Driver) (func() (*os.File, error), func(), error) {
	namespace, err := vsockns.Open()
	if errors.Is(err, vsockns.ErrNotPrivate) {
		driver.OpenVsock = func() (*os.File, error) { return nil, err }

		return nil, func() {}, nil
	}

	if err != nil {
		return nil, nil, err
	}

	driver.OpenVsock = namespace.Device

	return namespace.Socket, func() { _ = namespace.Close() }, nil
}
