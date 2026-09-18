package tree_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/tree"
)

func TestAFileKeepsItsContentAndMode(t *testing.T) {
	// arrange
	source := t.TempDir()
	target := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(source, "hello"), []byte("hi"), 0o600))
	require.NoError(t, os.Chmod(filepath.Join(source, "hello"), 0o400))

	// act
	err := tree.Copy(source, target)

	// assert
	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(target, "hello"))
	require.NoError(t, err)
	assert.Equal(t, "hi", string(got))
	info, err := os.Lstat(filepath.Join(target, "hello"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o400), info.Mode())
}

func TestALinkStaysALinkToWhatItNames(t *testing.T) {
	// arrange
	source := t.TempDir()
	target := t.TempDir()
	require.NoError(t, os.Symlink("/nowhere/at/all", filepath.Join(source, "link")))

	// act
	err := tree.Copy(source, target)

	// assert
	require.NoError(t, err)
	got, err := os.Readlink(filepath.Join(target, "link"))
	require.NoError(t, err)
	assert.Equal(t, "/nowhere/at/all", got)
}

func TestTwoNamesOfOneFileStayOneFile(t *testing.T) {
	// arrange
	source := t.TempDir()
	target := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(source, "a"), []byte("same"), 0o600))
	require.NoError(t, os.Link(filepath.Join(source, "a"), filepath.Join(source, "b")))

	// act
	err := tree.Copy(source, target)

	// assert
	require.NoError(t, err)
	a, err := os.Lstat(filepath.Join(target, "a"))
	require.NoError(t, err)
	b, err := os.Lstat(filepath.Join(target, "b"))
	require.NoError(t, err)
	assert.True(t, os.SameFile(a, b))
}

func TestFilesAndDirectoriesKeepTheirTime(t *testing.T) {
	// arrange
	source := t.TempDir()
	target := t.TempDir()
	then := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	require.NoError(t, os.Mkdir(filepath.Join(source, "a"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(source, "a/f"), nil, 0o600))
	require.NoError(t, os.Chtimes(filepath.Join(source, "a/f"), then, then))
	require.NoError(t, os.Chtimes(filepath.Join(source, "a"), then, then))

	// act
	err := tree.Copy(source, target)

	// assert
	require.NoError(t, err)
	file, err := os.Lstat(filepath.Join(target, "a/f"))
	require.NoError(t, err)
	assert.True(t, then.Equal(file.ModTime()), file.ModTime())
	dir, err := os.Lstat(filepath.Join(target, "a"))
	require.NoError(t, err)
	assert.True(t, then.Equal(dir.ModTime()), dir.ModTime())
}

func TestALinkKeepsItsOwnTime(t *testing.T) {
	// arrange
	source := t.TempDir()
	target := t.TempDir()
	then := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	require.NoError(t, os.Symlink("/nowhere/at/all", filepath.Join(source, "link")))
	times := []unix.Timespec{unix.NsecToTimespec(then.UnixNano()), unix.NsecToTimespec(then.UnixNano())}
	require.NoError(t, unix.UtimesNanoAt(unix.AT_FDCWD, filepath.Join(source, "link"), times, unix.AT_SYMLINK_NOFOLLOW))

	// act
	err := tree.Copy(source, target)

	// assert
	require.NoError(t, err)
	link, err := os.Lstat(filepath.Join(target, "link"))
	require.NoError(t, err)
	assert.True(t, then.Equal(link.ModTime()), link.ModTime())
}

func TestADirectoryKeepsWhatIsInItAndItsMode(t *testing.T) {
	// arrange
	source := t.TempDir()
	target := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(source, "a/b"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(source, "a/b/deep"), []byte("down"), 0o600))
	require.NoError(t, os.Chmod(filepath.Join(source, "a"), 0o500))
	t.Cleanup(func() {
		_ = os.Chmod(filepath.Join(source, "a"), 0o700)
		_ = os.Chmod(filepath.Join(target, "a"), 0o700)
	})

	// act
	err := tree.Copy(source, target)

	// assert
	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(target, "a/b/deep"))
	require.NoError(t, err)
	assert.Equal(t, "down", string(got))
	info, err := os.Lstat(filepath.Join(target, "a"))
	require.NoError(t, err)
	assert.Equal(t, os.ModeDir|0o500, info.Mode())
}
