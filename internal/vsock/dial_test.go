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
	conn, err = vsock.DialOn(unconnected, unix.VMADDR_CID_LOCAL, 1028)

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
func unconnected() (*os.File, error) {
	fd, err := unix.Socket(unix.AF_VSOCK, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(fd), "vsock"), nil
}
