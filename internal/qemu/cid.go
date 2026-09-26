package qemu

import (
	"errors"
	"fmt"

	"golang.org/x/sys/unix"
)

// tries bounds the search for a free CID. Of about four billion, ten taken
// in a row means the host is out of them or the device answers wrong.
const tries = 10

// claim takes a CID for the machine on the host's vhost-vsock device.
func claim(take func(cid uint32) error, random func() uint32) (uint32, error) {
	var err error
	for range tries {
		cid := draw(random)

		err = take(cid)
		if errors.Is(err, unix.EADDRINUSE) {
			continue
		}

		return cid, err
	}

	return 0, fmt.Errorf("no free CID after %d tries: %w", tries, err)
}

// draw is a random CID a machine may have: above the host's and never the
// one that means any, which the kernel refuses.
func draw(random func() uint32) uint32 {
	for {
		cid := random()
		if cid > unix.VMADDR_CID_HOST && cid != unix.VMADDR_CID_ANY {
			return cid
		}
	}
}
