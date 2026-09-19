//go:build vmtest

package agent_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/layer"
)

// onCache is a directory the cache disk was mounted on, as change left it
// before it was unmounted again.
func onCache(t *testing.T, change func(dir string)) string {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, agent.MountCache("miso-cache", dir))
	change(dir)
	require.NoError(t, syscall.Unmount(dir, 0))

	return dir
}

func unmountAtEnd(t *testing.T, dir string) {
	t.Helper()

	t.Cleanup(func() { assert.NoError(t, syscall.Unmount(dir, 0)) })
}

func TestALayerOnTheCacheIsThereAfterTheCacheIsMountedAgain(t *testing.T) {
	// arrange
	dir := onCache(t, func(dir string) {
		work, err := layer.Open(dir).Begin("abc")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(work.Dir(), "hello"), []byte("hi"), 0o600))
		require.NoError(t, work.Finish())
	})

	// act
	err := agent.MountCache("miso-cache", dir)
	unmountAtEnd(t, dir)

	// assert
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(dir, "abc", "hello"))
}

func TestACacheDiskWithoutAFileSystemFailsNamingItsSerial(t *testing.T) {
	// arrange
	dir := t.TempDir()

	// act
	err := agent.MountCache("miso-test-blank", dir)

	// assert
	assert.ErrorContains(t, err, "miso-test-blank")
}
