//go:build vmtest

package place_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/place"
)

func TestANewDirectoryHasTheModeOfItsEntry(t *testing.T) {
	// arrange
	root := t.TempDir()
	old := syscall.Umask(0o077)
	t.Cleanup(func() { syscall.Umask(old) })

	// act
	err := place.Open(root).Directory("/srv", 0o1775)

	// assert
	require.NoError(t, err)
	info, err := os.Lstat(filepath.Join(root, "srv"))
	require.NoError(t, err)
	assert.Equal(t, os.ModeDir|os.ModeSticky|0o775, info.Mode())
}

func TestADirectoryThatIsThereKeepsItsModeAndOwner(t *testing.T) {
	// arrange
	root := t.TempDir()
	etc := filepath.Join(root, "etc")
	require.NoError(t, os.Mkdir(etc, 0o700))
	require.NoError(t, os.Chown(etc, 1000, 1000))

	// act
	err := place.Open(root).Directory("/etc", 0o775)

	// assert
	require.NoError(t, err)
	info, err := os.Lstat(etc)
	require.NoError(t, err)
	assert.Equal(t, os.ModeDir|0o700, info.Mode())
	assert.Equal(t, [2]uint32{1000, 1000}, owner(t, etc))
}

func TestADirectoryOntoALinkToADirectoryGoesIntoItsTarget(t *testing.T) {
	// arrange
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "usr", "bin"), 0o700))
	require.NoError(t, os.Symlink("usr/bin", filepath.Join(root, "bin")))
	image := place.Open(root)

	// act
	err := image.Directory("/bin", 0o755)
	require.NoError(t, err)
	err = image.Link("/bin/sh", "dash")

	// assert
	require.NoError(t, err)
	info, err := os.Lstat(filepath.Join(root, "bin"))
	require.NoError(t, err)
	assert.Equal(t, os.ModeSymlink, info.Mode().Type())
	target, err := os.Readlink(filepath.Join(root, "usr", "bin", "sh"))
	require.NoError(t, err)
	assert.Equal(t, "dash", target)
}
