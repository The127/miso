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

func TestAQuestionTheHelperDoesNotKnowIsAnsweredAsFailed(t *testing.T) {
	// arrange
	namespace := vsocknstest.Private(t)

	// act
	_, err := vsockns.Ask(namespace, "shutdown")

	// assert
	assert.ErrorContains(t, err, "unknown question shutdown")
}

// channel is the helper's end of a connected pair of the kind miso hands
// its helper.
func channel(t *testing.T) int {
	t.Helper()

	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_SEQPACKET|unix.SOCK_CLOEXEC, 0)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = unix.Close(fds[0])
		_ = unix.Close(fds[1])
	})

	return fds[0]
}

func TestAProgramNamedLikeTheHelperWithoutItsChannelIsNoHelper(t *testing.T) {
	// arrange
	file, err := os.Open(os.DevNull)
	require.NoError(t, err)
	t.Cleanup(func() { _ = file.Close() })

	// act
	helper := vsockns.IsHelper([]string{"miso-vsockns", "/etc/passwd", "/dev/vhost-vsock"}, int(file.Fd()))

	// assert
	assert.False(t, helper)
}

func TestAProgramNamedLikeTheHelperWithTooFewArgumentsIsNoHelper(t *testing.T) {
	// arrange
	helperEnd := channel(t)

	// act
	helper := vsockns.IsHelper([]string{"miso-vsockns"}, helperEnd)

	// assert
	assert.False(t, helper)
}

func TestTheHelperMisoStartsIsAHelper(t *testing.T) {
	// arrange
	helperEnd := channel(t)

	// act
	helper := vsockns.IsHelper([]string{"miso-vsockns", "/proc/sys/net/vsock/child_ns_mode", "/dev/vhost-vsock"}, helperEnd)

	// assert
	assert.True(t, helper)
}
