//go:build vmtest

package vsock_test

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/vsock"
)

// dialing is a listener on a port of this machine with a connection made
// to it that is not accepted yet.
func dialing(t *testing.T, port uint32) (*vsock.Listener, *os.File) {
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
	flags, err := unix.FcntlInt(accepted.(*os.File).Fd(), unix.F_GETFD, 0)
	require.NoError(t, err)
	assert.NotZero(t, flags&unix.FD_CLOEXEC)
}

func TestAListenerIsNotInheritedByAProcess(t *testing.T) {
	// act
	listener, err := vsock.Listen(1026)

	// assert
	require.NoError(t, err)
	flags, err := unix.FcntlInt(uintptr(vsock.FD(listener)), unix.F_GETFD, 0)
	require.NoError(t, err)
	assert.NotZero(t, flags&unix.FD_CLOEXEC)
}
