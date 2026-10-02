package cachedisk_test

import (
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/cachedisk"
)

// fileSystemBytes is how large the ext4 in a file says it is, from its
// superblock, which is 1024 bytes in.
func fileSystemBytes(t *testing.T, path string) int64 {
	t.Helper()

	file, err := os.Open(path)
	require.NoError(t, err)

	defer func() { _ = file.Close() }()

	superblock := make([]byte, 1024)
	_, err = file.ReadAt(superblock, 1024)
	require.NoError(t, err)

	blocks := int64(binary.LittleEndian.Uint32(superblock[4:]))
	blockSize := int64(1024) << binary.LittleEndian.Uint32(superblock[24:])

	return blocks * blockSize
}

func TestAGrownCacheDiskIsAsLargeAsAskedForInTheFileAndInItsFileSystem(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.img")
	require.NoError(t, cachedisk.Make(path, 64<<20))

	// act
	err := cachedisk.Grow(path, 128<<20)

	// assert
	require.NoError(t, err)
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, int64(128<<20), info.Size())
	assert.Equal(t, int64(128<<20), fileSystemBytes(t, path))
}

func TestACacheDiskIsNeverMadeSmallerAndStaysAsItWas(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.img")
	require.NoError(t, cachedisk.Make(path, 64<<20))

	// act
	err := cachedisk.Grow(path, 32<<20)

	// assert
	assert.ErrorContains(t, err, "only grows")
	info, statErr := os.Stat(path)
	require.NoError(t, statErr)
	assert.Equal(t, int64(64<<20), info.Size())
	assert.Equal(t, int64(64<<20), fileSystemBytes(t, path))
}

func TestACacheDiskOfTheSizeAskedForIsLeftAlone(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.img")
	require.NoError(t, cachedisk.Make(path, 64<<20))

	// act
	err := cachedisk.Grow(path, 64<<20)

	// assert
	assert.ErrorContains(t, err, "only grows")
}

func TestACacheDiskThatIsNotThereCannotBeGrown(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.img")

	// act
	err := cachedisk.Grow(path, 64<<20)

	// assert
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestACacheDiskWithoutResize2fsNamesWhereToGetItAndStaysAsItWas(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.img")
	require.NoError(t, cachedisk.Make(path, 64<<20))
	// e2fsck is there, so the disk would be checked and grown first if the
	// lookup of resize2fs came after
	t.Setenv("PATH", toolsOf(t, ""))
	cachedisk.SearchAlso(t)

	// act
	err := cachedisk.Grow(path, 128<<20)

	// assert
	assert.ErrorContains(t, err, "e2fsprogs")
	info, statErr := os.Stat(path)
	require.NoError(t, statErr)
	assert.Equal(t, int64(64<<20), info.Size())
}

func TestAGrownCacheDiskKeepsItsFiles(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.img")
	require.NoError(t, cachedisk.Make(path, 64<<20))
	marker := filepath.Join(t.TempDir(), "marker")
	require.NoError(t, os.WriteFile(marker, []byte("a layer\n"), 0o600))
	debugfs(t, path, "write "+marker+" marker")

	// act
	err := cachedisk.Grow(path, 128<<20)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "a layer\n", debugfs(t, path, "cat marker"))
}

func TestACacheDiskThatWasNotCheckedAfterItsLastMountIsStillGrown(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.img")
	require.NoError(t, cachedisk.Make(path, 64<<20))
	// a build killed while it had the disk mounted leaves it like this, and
	// resize2fs refuses it until it was checked
	debugfs(t, path, "set_super_value state 0")

	// act
	err := cachedisk.Grow(path, 128<<20)

	// assert
	require.NoError(t, err)
	assert.Equal(t, int64(128<<20), fileSystemBytes(t, path))
}

// debugfs runs a command of debugfs on a file system in a file, which may
// write to it, and answers what it printed.
func debugfs(t *testing.T, path, command string) string {
	t.Helper()

	said, err := exec.Command("debugfs", "-w", "-R", command, path).Output() //nolint:gosec // the path is a temporary file of the test
	require.NoError(t, err)

	return string(said)
}

// toolsOf is a directory for the PATH that holds the real e2fsck and the
// given script as resize2fs, so a test decides what the resize does.
func toolsOf(t *testing.T, resize string) string {
	t.Helper()

	check, err := exec.LookPath("e2fsck")
	require.NoError(t, err)

	dir := t.TempDir()
	require.NoError(t, os.Symlink(check, filepath.Join(dir, "e2fsck")))

	if resize != "" {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "resize2fs"), []byte("#!/bin/sh\n"+resize+"\n"), 0o700)) //nolint:gosec // a script of the test, which has to run
	}

	return dir
}

func TestACacheDiskWhoseResizeFailsIsAsLargeAsItWasSoThatAnotherTryResizesAgain(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.img")
	require.NoError(t, cachedisk.Make(path, 64<<20))
	t.Setenv("PATH", toolsOf(t, "exit 1"))
	cachedisk.SearchAlso(t)

	// act
	err := cachedisk.Grow(path, 128<<20)

	// assert
	require.Error(t, err)
	info, statErr := os.Stat(path)
	require.NoError(t, statErr)
	assert.Equal(t, int64(64<<20), info.Size())
}
