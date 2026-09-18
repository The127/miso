package tree_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
