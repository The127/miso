//go:build vmtest

package vsock_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/vsock"
)

func TestAConnectionCarriesWhatTheOtherSideWrites(t *testing.T) {
	// arrange
	listener, err := vsock.Listen(1024)
	require.NoError(t, err)
	dialed, err := vsock.Dial(unix.VMADDR_CID_LOCAL, 1024)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, dialed.Close()) })
	_, err = io.WriteString(dialed, "hello")
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
