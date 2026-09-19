//go:build vmtest

package sandbox_test

import (
	"context"
	"io"
	"os"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/sandbox"
)

// open are the file descriptors this process holds.
func open(t *testing.T) []int {
	t.Helper()

	entries, err := os.ReadDir("/proc/self/fd")
	require.NoError(t, err)
	var fds []int
	for _, entry := range entries {
		fd, err := strconv.Atoi(entry.Name())
		require.NoError(t, err)
		fds = append(fds, fd)
	}

	return fds
}

// roomFor leaves the process room for a number of file descriptors and no
// more, so the one after them cannot be made.
func roomFor(t *testing.T, more int) {
	t.Helper()

	var limit unix.Rlimit
	require.NoError(t, unix.Getrlimit(unix.RLIMIT_NOFILE, &limit))
	t.Cleanup(func() { assert.NoError(t, unix.Setrlimit(unix.RLIMIT_NOFILE, &limit)) })

	lowered := limit
	lowered.Cur = uint64(slices.Max(open(t)) + 1 + more) //nolint:gosec // a test counts a handful of files
	require.NoError(t, unix.Setrlimit(unix.RLIMIT_NOFILE, &lowered))
}

func TestARunThatCannotBeSetUpLeavesNoPipesBehind(t *testing.T) {
	// arrange
	before := open(t)
	// the run's first pipe fits, its second does not
	roomFor(t, 2)

	// act
	_, err := sandbox.Run(context.Background(), t.TempDir(), protocol.Run{Command: "true"}, io.Discard)

	// assert
	require.Error(t, err)
	assert.Equal(t, before, open(t))
}
