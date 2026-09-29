package vsockns_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/vsockns"
	"github.com/The127/miso/internal/vsockns/vsocknstest"
)

// reusing puts a file of its own at the number, as the next file opened
// after a close gets it.
func reusing(t *testing.T, number int) {
	t.Helper()

	file, err := os.Open(os.DevNull)
	require.NoError(t, err)
	t.Cleanup(func() { _ = file.Close() })

	if int(file.Fd()) != number {
		require.NoError(t, unix.Dup3(int(file.Fd()), number, unix.O_CLOEXEC))
		t.Cleanup(func() { _ = unix.Close(number) })
	}
}

func TestANamespaceClosedTwiceLeavesTheFileNowAtItsNumber(t *testing.T) {
	// arrange
	namespace := vsocknstest.Private(t)
	number := vsockns.Conn(namespace)
	require.NoError(t, namespace.Close())
	reusing(t, number)

	// act
	_ = namespace.Close()

	// assert
	_, err := unix.FcntlInt(uintptr(number), unix.F_GETFD, 0)
	assert.NoError(t, err)
}
