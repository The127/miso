package qemu

import (
	"fmt"
	"math/rand/v2"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

// vhostVsockSetGuestCID is VHOST_VSOCK_SET_GUEST_CID of linux/vhost.h,
// _IOW(0xAF, 0x60, __u64), which x/sys does not carry.
const vhostVsockSetGuestCID = 0x4008af60

// openVsock opens the host's vhost-vsock device, on which the host claims a
// CID and which QEMU then gets to keep it.
func openVsock(path string) (*os.File, error) {
	device, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("the host's vsock device needs the vhost_vsock module and access for this user: %w", err)
	}

	return device, nil
}

// holdCID opens the device and claims a CID on it, held for as long as the
// device stays open.
func holdCID(path string) (*os.File, uint32, error) {
	device, err := openVsock(path)
	if err != nil {
		return nil, 0, err
	}

	cid, err := claim(func(cid uint32) error { return takeCID(device, cid) }, rand.Uint32)
	if err != nil {
		_ = device.Close()

		return nil, 0, err
	}

	return device, cid, nil
}

// takeCID holds the CID on the open device for as long as the device stays
// open, the kernel refuses it to every other open device meanwhile.
func takeCID(device *os.File, cid uint32) error {
	// the kernel copies a __u64 from the pointer, a uint32 would hand it four
	// bytes of whatever follows
	wide := uint64(cid)

	//nolint:gosec // the ioctl reads a __u64 through a pointer, x/sys has no helper for it
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, device.Fd(), vhostVsockSetGuestCID, uintptr(unsafe.Pointer(&wide)))
	if errno != 0 {
		return errno
	}

	return nil
}
