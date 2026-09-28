package vsockns_test

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/vsockns"
)

// netnsOf names the network namespace the namespace's sockets live in, the
// way /proc/self/ns/net names that of a process.
func netnsOf(t *testing.T, namespace *vsockns.Namespace) string {
	t.Helper()

	socket, err := namespace.Socket()
	require.NoError(t, err)
	t.Cleanup(func() { _ = socket.Close() })

	netns, err := unix.IoctlRetInt(int(socket.Fd()), unix.SIOCGSKNS)
	require.NoError(t, err)
	t.Cleanup(func() { _ = unix.Close(netns) })

	var stat unix.Stat_t
	require.NoError(t, unix.Fstat(netns, &stat))

	return fmt.Sprintf("net:[%d]", stat.Ino)
}

func TestAProgramRunInTheNamespaceRunsInIt(t *testing.T) {
	// arrange
	namespace := opened(t)
	read, written, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { _ = read.Close() })
	inside := netnsOf(t, namespace)

	// act
	err = namespace.Run([]string{"readlink", "/proc/self/ns/net"}, written)

	// assert
	require.NoError(t, err)
	require.NoError(t, written.Close())
	theirs, err := io.ReadAll(read)
	require.NoError(t, err)
	assert.Equal(t, inside, strings.TrimSpace(string(theirs)))
}

// inherited is a file miso holds without close-on-exec, the way it holds
// what its own caller left open for it.
func inherited(t *testing.T) {
	t.Helper()

	null, err := os.Open(os.DevNull)
	require.NoError(t, err)
	t.Cleanup(func() { _ = null.Close() })

	fd, err := unix.Dup(int(null.Fd()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = unix.Close(fd) })
}

func TestAProgramRunInTheNamespaceHoldsOnlyItsStandardFiles(t *testing.T) {
	// arrange
	inherited(t)
	namespace := opened(t)
	read, written, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { _ = read.Close() })

	// act
	err = namespace.Run([]string{"ls", "-l", "/proc/self/fd"}, written)

	// assert
	require.NoError(t, err)
	require.NoError(t, written.Close())
	listed, err := io.ReadAll(read)
	require.NoError(t, err)
	require.Contains(t, string(listed), " 1 -> pipe:")
	for _, line := range strings.Split(string(listed), "\n") {
		fd, target, isLink := strings.Cut(line, " -> ")
		if !isLink {
			continue
		}

		fields := strings.Fields(fd)
		number := fields[len(fields)-1]
		// ls holds the directory it lists
		if number != "0" && number != "1" && number != "2" {
			assert.True(t, strings.HasPrefix(target, "/proc/"), "fd %s -> %s", number, target)
		}
	}
}
