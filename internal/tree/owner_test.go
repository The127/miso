//go:build vmtest

package tree_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/tree"
)

func TestFilesDirectoriesAndLinksKeepTheirOwners(t *testing.T) {
	// arrange
	source := t.TempDir()
	target := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(source, "f"), nil, 0o600))
	require.NoError(t, os.Chown(filepath.Join(source, "f"), 1234, 5678))
	require.NoError(t, os.Mkdir(filepath.Join(source, "d"), 0o700))
	require.NoError(t, os.Chown(filepath.Join(source, "d"), 2345, 6789))
	require.NoError(t, os.Symlink("f", filepath.Join(source, "l")))
	require.NoError(t, os.Lchown(filepath.Join(source, "l"), 4321, 8765))

	// act
	err := tree.Copy(source, target)

	// assert
	require.NoError(t, err)
	assert.Equal(t, [2]uint32{1234, 5678}, owner(t, filepath.Join(target, "f")))
	assert.Equal(t, [2]uint32{2345, 6789}, owner(t, filepath.Join(target, "d")))
	assert.Equal(t, [2]uint32{4321, 8765}, owner(t, filepath.Join(target, "l")))
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
