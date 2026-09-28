package vsockns_test

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/vsockns"
)

// cookie names the network namespace a socket lives in.
func cookie(t *testing.T, socket *os.File) uint64 {
	t.Helper()

	raw, err := socket.SyscallConn()
	require.NoError(t, err)

	var found uint64
	var failed error
	require.NoError(t, raw.Control(func(fd uintptr) {
		found, failed = unix.GetsockoptUint64(int(fd), unix.SOL_SOCKET, unix.SO_NETNS_COOKIE)
	}))
	require.NoError(t, failed)

	return found
}

// ours is a socket miso makes in its own network namespace.
func ours(t *testing.T) *os.File {
	t.Helper()

	fd, err := unix.Socket(unix.AF_UNIX, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
	require.NoError(t, err)
	socket := os.NewFile(uintptr(fd), "ours")
	t.Cleanup(func() { _ = socket.Close() })

	return socket
}

func TestAListenerOfTheNamespaceIsNotInMisosOwn(t *testing.T) {
	// arrange
	namespace, err := vsockns.Open()
	if errors.Is(err, vsockns.ErrNotPrivate) {
		t.Skip("this host cannot keep vsock private")
	}

	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, namespace.Close()) })

	// act
	listener, err := namespace.Listen(unix.VMADDR_PORT_ANY)

	// assert
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })
	assert.NotEqual(t, cookie(t, ours(t)), cookie(t, listener))
}
