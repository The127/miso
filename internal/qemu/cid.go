package qemu

import (
	"errors"

	"golang.org/x/sys/unix"
)

// claim takes a CID for the machine on the host's vhost-vsock device.
func claim(take func(cid uint32) error, random func() uint32) (uint32, error) {
	for {
		cid := random()

		err := take(cid)
		if errors.Is(err, unix.EADDRINUSE) {
			continue
		}

		return cid, err
	}
}
