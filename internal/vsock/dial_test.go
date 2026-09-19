//go:build vmtest

package vsock_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func TestADialedConnectionIsNotInheritedByAProcess(t *testing.T) {
	// act
	_, dialed := dialing(t, 1027)

	// assert
	flags, err := unix.FcntlInt(dialed.Fd(), unix.F_GETFD, 0)
	require.NoError(t, err)
	assert.NotZero(t, flags&unix.FD_CLOEXEC)
}
