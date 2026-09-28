package vsockns_test

import (
	"math/rand/v2"
	"os"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

// vhostVsockSetGuestCID is VHOST_VSOCK_SET_GUEST_CID of linux/vhost.h,
// _IOW(0xAF, 0x60, __u64), which x/sys does not carry.
const vhostVsockSetGuestCID = 0x4008af60

// claimed is a CID a VM on the device would have.
func claimed(t *testing.T, device *os.File) uint32 {
	t.Helper()

	// above the CIDs a host hands out by itself
	//nolint:gosec // a CID is no secret
	cid := 1<<30 + rand.Uint32N(1<<30)
	wide := uint64(cid)

	//nolint:gosec // the ioctl reads a __u64 through a pointer, x/sys has no helper for it
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, device.Fd(), vhostVsockSetGuestCID, uintptr(unsafe.Pointer(&wide)))
	require.Zero(t, errno)

	return cid
}

func TestADeviceOfTheNamespaceHasItsVMsOutOfMisosReach(t *testing.T) {
	// arrange
	namespace := opened(t)

	// act
	device, err := namespace.Device()

	// assert
	require.NoError(t, err)
	t.Cleanup(func() { _ = device.Close() })
	cid := claimed(t, device)
	socket, err := unix.Socket(unix.AF_VSOCK, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
	require.NoError(t, err)
	t.Cleanup(func() { _ = unix.Close(socket) })
	err = unix.Connect(socket, &unix.SockaddrVM{CID: cid, Port: 1024})
	assert.ErrorIs(t, err, unix.ENODEV)
}
