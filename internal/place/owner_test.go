//go:build vmtest

package place_test

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/place"
)

func TestAFileIsRootsEvenInADirectoryThatHandsOnItsGroup(t *testing.T) {
	// arrange
	root := t.TempDir()
	srv := filepath.Join(root, "srv")
	require.NoError(t, os.Mkdir(srv, 0o700))
	require.NoError(t, os.Chown(srv, 1000, 1000))
	require.NoError(t, os.Chmod(srv, 0o2775|os.ModeSetgid))

	// act
	err := place.Open(root).File("/srv/motd", 0o644, strings.NewReader("hello\n"))

	// assert
	require.NoError(t, err)
	assert.Equal(t, [2]uint32{0, 0}, owner(t, filepath.Join(srv, "motd")))
}

// owner is the user and group that own a path, not following a link.
func owner(t *testing.T, name string) [2]uint32 {
	t.Helper()

	info, err := os.Lstat(name)
	require.NoError(t, err)

	stat, ok := info.Sys().(*syscall.Stat_t)
	require.True(t, ok)

	return [2]uint32{stat.Uid, stat.Gid}
}
