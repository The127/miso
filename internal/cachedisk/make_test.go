package cachedisk_test

import (
	"io/fs"
	"os"
	"os/exec"
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

func TestACacheDiskKilledHalfwayIsNeverAtThePath(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.img")
	// writes a little of the disk, then kills the miso that started it
	fakeMkfs(t, "printf half > \"$2\"\nkill -9 $PPID\n")
	self, err := os.Executable()
	require.NoError(t, err)
	child := exec.Command(self, path) //nolint:gosec // the test runs its own binary
	child.Args[0] = maker

	// act
	err = child.Run()

	// assert
	require.Error(t, err)
	assert.NoFileExists(t, path)
}

func TestAFailedCacheDiskSaysWhatMkfsSaid(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.img")
	fakeMkfs(t, "echo no room for a file system >&2\nexit 1\n")

	// act
	err := cachedisk.Make(path, 64<<20)

	// assert
	assert.ErrorContains(t, err, "no room for a file system")
}

func TestACacheDiskWithoutMkfsNamesWhereToGetIt(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.img")
	t.Setenv("PATH", t.TempDir())
	cachedisk.SearchAlso(t)

	// act
	err := cachedisk.Make(path, 64<<20)

	// assert
	assert.ErrorContains(t, err, "e2fsprogs")
}

func TestACacheDiskFindsMkfsWhereThePathDoesNotLook(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.img")
	t.Setenv("PATH", t.TempDir())
	// as /usr/sbin, which a user's PATH on Debian leaves out
	sbin := t.TempDir()
	mkfsIn(t, sbin, "printf made > \"$2\"\n")
	cachedisk.SearchAlso(t, sbin)

	// act
	err := cachedisk.Make(path, 64<<20)

	// assert
	require.NoError(t, err)
	assert.FileExists(t, path)
}

func TestANewCacheDiskRemovesWhatAKilledOneLeft(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.img")
	left := path + ".making-123"
	require.NoError(t, os.WriteFile(left, []byte("half"), 0o600))

	// act
	err := cachedisk.Make(path, 64<<20)

	// assert
	require.NoError(t, err)
	assert.NoFileExists(t, left)
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

// fakeMkfs puts an mkfs.ext4 first on the PATH that runs a shell script.
func fakeMkfs(t *testing.T, script string) {
	t.Helper()

	bin := t.TempDir()
	mkfsIn(t, bin, script)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// mkfsIn puts an mkfs.ext4 into a directory that runs a shell script, which
// finds the disk in $2.
func mkfsIn(t *testing.T, dir, script string) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "mkfs.ext4"), []byte("#!/bin/sh\n"+script), 0o700)) //nolint:gosec // the script must be executable
}
