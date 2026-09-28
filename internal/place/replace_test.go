//go:build vmtest

package place_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/place"
)

func TestAFileReplacesALinkInsteadOfWritingThroughIt(t *testing.T) {
	// arrange
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "real"), []byte("kept\n"), 0o600))
	require.NoError(t, os.Symlink("/real", filepath.Join(root, "motd")))

	// act
	err := place.Open(root).File("/motd", 0o644, strings.NewReader("hello\n"))

	// assert
	require.NoError(t, err)
	info, err := os.Lstat(filepath.Join(root, "motd"))
	require.NoError(t, err)
	assert.True(t, info.Mode().IsRegular())
	kept, err := os.ReadFile(filepath.Join(root, "real"))
	require.NoError(t, err)
	assert.Equal(t, "kept\n", string(kept))
}

func TestAFileOntoADirectoryFailsNamingThePathInTheImage(t *testing.T) {
	// arrange
	root := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, "etc"), 0o700))

	// act
	err := place.Open(root).File("/etc", 0o644, strings.NewReader("hello\n"))

	// assert
	require.ErrorIs(t, err, unix.EISDIR)
	assert.ErrorContains(t, err, " /etc:")
	assert.DirExists(t, filepath.Join(root, "etc"))
}
