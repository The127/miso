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

func TestALayerOnTheCacheIsThereAfterTheCacheIsMountedAgain(t *testing.T) {
	// arrange
	dir := t.TempDir()
	require.NoError(t, agent.MountCache("miso-cache", dir))
	store := layer.Open(dir)
	work, err := store.Begin("abc")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(work.Dir(), "hello"), []byte("hi"), 0o600))
	require.NoError(t, work.Finish())
	require.NoError(t, syscall.Unmount(dir, 0))

	// act
	err = agent.MountCache("miso-cache", dir)
	t.Cleanup(func() { assert.NoError(t, syscall.Unmount(dir, 0)) })

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
