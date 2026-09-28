//go:build vmtest

package vsock_test

import (
	"io"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/vsock"
)

func TestADialedConnectionIsNotInheritedByAProcess(t *testing.T) {
	// act
	_, dialed := dialing(t, 1027)

	// assert
	conn, isConn := dialed.(syscall.Conn)
	require.True(t, isConn)
	assert.True(t, closedOnExec(t, conn))
}

func TestAFailedDialGivesNoConnection(t *testing.T) {
	// arrange
	var conn io.ReadWriteCloser
	var err error

	// act
	conn, err = vsock.Dial(unix.VMADDR_CID_LOCAL, 1028)

	// assert
	require.Error(t, err)
	// assert.Nil would take an interface around a nil pointer for nil
	assert.True(t, conn == nil)
}

func TestAConnectionClosedForWritingStillHearsTheAnswer(t *testing.T) {
	// arrange
	listener, dialed := dialing(t, 1028)
	accepted, err := listener.Accept()
	require.NoError(t, err)
	_, err = io.WriteString(dialed, "command -v sh\n")
	require.NoError(t, err)

	// act
	err = dialed.CloseWrite()

	// assert
	require.NoError(t, err)
	asked, err := io.ReadAll(accepted)
	require.NoError(t, err)
	assert.Equal(t, "command -v sh\n", string(asked))
	_, err = io.WriteString(accepted, "/usr/bin/sh\n")
	require.NoError(t, err)
	require.NoError(t, accepted.Close())
	answer, err := io.ReadAll(dialed)
	require.NoError(t, err)
	assert.Equal(t, "/usr/bin/sh\n", string(answer))
}

// unconnected is a socket of this machine that is neither bound nor
// connected, made outside the package as a namespace makes it.
func unconnected(t *testing.T) *os.File {
	t.Helper()

	fd, err := unix.Socket(unix.AF_VSOCK, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
	require.NoError(t, err)
	socket := os.NewFile(uintptr(fd), "vsock")
	t.Cleanup(func() { assert.NoError(t, socket.Close()) })

	return socket
}

func TestAConnectionDialedOnAHandedSocketCarriesWhatItWrites(t *testing.T) {
	// arrange
	listener, err := vsock.Listen(unix.VMADDR_PORT_ANY)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, listener.Close()) })

	// act
	dialed, err := vsock.DialOn(unconnected(t), unix.VMADDR_CID_LOCAL, listener.Port())

	// assert
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, dialed.Close()) })
	_, err = io.WriteString(dialed, "hello")
	require.NoError(t, err)
	accepted, err := listener.Accept()
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, accepted.Close()) })
	got := make([]byte, 5)
	_, err = io.ReadFull(accepted, got)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(got))
}

func TestAConnectionDialedOnAHandedSocketClosedForWritingStillHearsTheAnswer(t *testing.T) {
	// arrange
	listener, err := vsock.Listen(unix.VMADDR_PORT_ANY)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, listener.Close()) })
	dialed, err := vsock.DialOn(unconnected(t), unix.VMADDR_CID_LOCAL, listener.Port())
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, dialed.Close()) })
	accepted, err := listener.Accept()
	require.NoError(t, err)
	_, err = io.WriteString(dialed, "command -v sh\n")
	require.NoError(t, err)

	// act
	err = dialed.CloseWrite()

	// assert
	require.NoError(t, err)
	asked, err := io.ReadAll(accepted)
	require.NoError(t, err)
	assert.Equal(t, "command -v sh\n", string(asked))
	_, err = io.WriteString(accepted, "/usr/bin/sh\n")
	require.NoError(t, err)
	require.NoError(t, accepted.Close())
	answer, err := io.ReadAll(dialed)
	require.NoError(t, err)
	assert.Equal(t, "/usr/bin/sh\n", string(answer))
}
