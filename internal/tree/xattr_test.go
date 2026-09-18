//go:build vmtest

package tree_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/tree"
)

// netRaw is a file capability of version 2 granting cap_net_raw, effective
// and permitted, as ping has it.
var netRaw = []byte{
	0x01, 0x00, 0x00, 0x02,
	0x00, 0x20, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
}

func TestExtendedAttributesAreKept(t *testing.T) {
	// arrange
	source := t.TempDir()
	target := t.TempDir()
	file := filepath.Join(source, "ping")
	require.NoError(t, os.WriteFile(file, nil, 0o600))
	require.NoError(t, os.Chown(file, 1234, 1234))
	require.NoError(t, unix.Lsetxattr(file, "user.note", []byte("file"), 0))
	require.NoError(t, unix.Lsetxattr(file, "security.capability", netRaw, 0))
	require.NoError(t, os.Mkdir(filepath.Join(source, "d"), 0o700))
	require.NoError(t, unix.Lsetxattr(filepath.Join(source, "d"), "user.note", []byte("dir"), 0))
	require.NoError(t, os.Symlink("ping", filepath.Join(source, "l")))
	require.NoError(t, unix.Lsetxattr(filepath.Join(source, "l"), "trusted.note", []byte("link"), 0))

	// act
	err := tree.Copy(source, target)

	// assert
	require.NoError(t, err)
	assert.Equal(t, netRaw, xattr(t, filepath.Join(target, "ping"), "security.capability"))
	assert.Equal(t, "file", string(xattr(t, filepath.Join(target, "ping"), "user.note")))
	assert.Equal(t, "dir", string(xattr(t, filepath.Join(target, "d"), "user.note")))
	assert.Equal(t, "link", string(xattr(t, filepath.Join(target, "l"), "trusted.note")))
}

func TestAnAttributeThatCannotBeKeptNamesItsFile(t *testing.T) {
	// arrange
	source := t.TempDir()
	target := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(source, "noted"), nil, 0o600))
	require.NoError(t, unix.Lsetxattr(filepath.Join(source, "noted"), "user.note", []byte("file"), 0))
	require.NoError(t, unix.Mount("ramfs", target, "ramfs", 0, ""))
	t.Cleanup(func() { _ = unix.Unmount(target, unix.MNT_DETACH) })

	// act
	err := tree.Copy(source, target)

	// assert
	assert.ErrorContains(t, err, "noted")
}

// xattr is the value of an extended attribute of a path, not following a
// link.
func xattr(t *testing.T, name, attribute string) []byte {
	t.Helper()

	value := make([]byte, 256)
	size, err := unix.Lgetxattr(name, attribute, value)
	require.NoError(t, err, attribute)

	return value[:size]
}
