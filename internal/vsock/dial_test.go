//go:build vmtest

package vsock_test

import (
	"io"
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
