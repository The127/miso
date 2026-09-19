//go:build vmtest

package vsock_test

import (
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"

	mdvsock "github.com/mdlayher/vsock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/vsock"
)

// dialing is a listener on a port of this machine with a connection made
// to it that is not accepted yet.
func dialing(t *testing.T, port uint32) (*vsock.Listener, *mdvsock.Conn) {
	t.Helper()
	listener, err := vsock.Listen(port)
	require.NoError(t, err)
	dialed, err := vsock.Dial(unix.VMADDR_CID_LOCAL, port)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, dialed.Close()) })

	return listener, dialed
}

func TestAConnectionCarriesWhatTheOtherSideWrites(t *testing.T) {
	// arrange
	listener, dialed := dialing(t, 1024)
	_, err := io.WriteString(dialed, "hello")
	require.NoError(t, err)

	// act
	accepted, err := listener.Accept()
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, accepted.Close()) })
	got := make([]byte, 5)
	_, err = io.ReadFull(accepted, got)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "hello", string(got))
}

func TestAnAcceptedConnectionIsNotInheritedByAProcess(t *testing.T) {
	// arrange
	listener, _ := dialing(t, 1025)

	// act
	accepted, err := listener.Accept()

	// assert
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, accepted.Close()) })
	assert.True(t, closedOnExec(t, accepted.(syscall.Conn)))
}

func TestAListenerIsNotInheritedByAProcess(t *testing.T) {
	// arrange
	before := sockets(t)

	// act
	_, err := vsock.Listen(1026)

	// assert
	require.NoError(t, err)
	var added []int
	for _, fd := range sockets(t) {
		if !slices.Contains(before, fd) {
			added = append(added, fd)
		}
	}

	require.Len(t, added, 1)
	flags, err := unix.FcntlInt(uintptr(added[0]), unix.F_GETFD, 0)
	require.NoError(t, err)
	assert.NotZero(t, flags&unix.FD_CLOEXEC)
}

// sockets are the file descriptors of this process that are sockets. A
// listener hides its own, so a test finds it among them.
func sockets(t *testing.T) []int {
	t.Helper()
	entries, err := os.ReadDir("/proc/self/fd")
	require.NoError(t, err)
	var fds []int
	for _, entry := range entries {
		target, err := os.Readlink("/proc/self/fd/" + entry.Name())
		if err != nil || !strings.HasPrefix(target, "socket:") {
			continue
		}

		fd, err := strconv.Atoi(entry.Name())
		require.NoError(t, err)
		fds = append(fds, fd)
	}

	return fds
}

func closedOnExec(t *testing.T, conn syscall.Conn) bool {
	t.Helper()
	raw, err := conn.SyscallConn()
	require.NoError(t, err)
	var flags int
	require.NoError(t, raw.Control(func(fd uintptr) {
		flags, err = unix.FcntlInt(fd, unix.F_GETFD, 0)
	}))
	require.NoError(t, err)

	return flags&unix.FD_CLOEXEC != 0
}
