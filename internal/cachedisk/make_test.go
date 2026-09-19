package cachedisk_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/cachedisk"
	"github.com/The127/miso/internal/disk"
)

func TestANewCacheDiskIsExt4(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.img")

	// act
	err := cachedisk.Make(path, 64<<20)

	// assert
	require.NoError(t, err)
	file, err := os.Open(path)
	require.NoError(t, err)

	defer func() { _ = file.Close() }()

	kind, err := disk.FileSystem(file)
	require.NoError(t, err)
	assert.Equal(t, "ext4", kind)
}

func TestANewCacheDiskHasTheSizeAskedFor(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.img")

	// act
	err := cachedisk.Make(path, 96<<20)

	// assert
	require.NoError(t, err)
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, int64(96<<20), info.Size())
}

func TestACacheDiskThatIsThereIsNeverMadeAgain(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.img")
	require.NoError(t, cachedisk.Make(path, 64<<20))
	before := fileSystemID(t, path)

	// act
	err := cachedisk.Make(path, 64<<20)

	// assert
	assert.ErrorIs(t, err, fs.ErrExist)
	assert.Equal(t, before, fileSystemID(t, path))
}

func TestAFailedCacheDiskLeavesNothingAtThePath(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.img")

	// act
	// too small for mkfs, which fails after it made the file
	err := cachedisk.Make(path, 8<<10)

	// assert
	require.Error(t, err)
	assert.NoFileExists(t, path)
}

// fileSystemID reads the UUID of the ext4 at a path, which every mkfs picks
// anew.
func fileSystemID(t *testing.T, path string) []byte {
	t.Helper()

	file, err := os.Open(path)
	require.NoError(t, err)

	defer func() { _ = file.Close() }()

	id := make([]byte, 16)
	// s_uuid in the superblock, which starts 1024 bytes in
	_, err = file.ReadAt(id, 1024+0x68)
	require.NoError(t, err)

	return id
}
