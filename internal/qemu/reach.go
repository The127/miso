package qemu

import (
	"fmt"
	"os"
)

// reach is how the host reaches the machine's agent: a CID on vsock, or a
// virtio port on a host without vsock.
type reach struct {
	args    []string
	machine *os.File
	cid     uint32
	host    *os.File

	// withoutVsock is why the host has no vsock for the machine
	withoutVsock error
}

// reachFor claims a CID on the host's vsock device, or else gives the
// machine a virtio port.
func reachFor(vsockDevice string) (reach, error) {
	device, cid, withoutVsock := holdCID(vsockDevice)
	if withoutVsock == nil {
		return reach{args: vsock(cid, deviceFD), machine: device, cid: cid}, nil
	}

	host, machine, err := socketPair()
	if err != nil {
		return reach{}, fmt.Errorf("%w, and no virtio port either: %w", withoutVsock, err)
	}

	return reach{args: port(deviceFD), machine: machine, host: host, withoutVsock: withoutVsock}, nil
}

// close lets go of the host's end, when QEMU never ran.
func (r reach) close() error {
	if r.host == nil {
		return nil
	}

	return r.host.Close()
}
