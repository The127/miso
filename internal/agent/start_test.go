//go:build vmtest

package agent_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/protocol"
)

func TestAStartedAgentHasSweptWhatAStoppedOneLeft(t *testing.T) {
	// arrange
	dir := t.TempDir()
	require.NoError(t, agent.MountCache("miso-cache", dir))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "layers", "work-left"), 0o700))
	require.NoError(t, syscall.Unmount(dir, 0))

	// act
	_, err := agent.Start("miso-cache", dir)
	t.Cleanup(func() { assert.NoError(t, syscall.Unmount(dir, 0)) })

	// assert
	require.NoError(t, err)
	assert.NoDirExists(t, filepath.Join(dir, "layers", "work-left"))
}

func TestAnAgentStartsOnACacheDiskThatHoldsNoLayersYet(t *testing.T) {
	// arrange
	dir := t.TempDir()
	require.NoError(t, agent.MountCache("miso-cache", dir))
	require.NoError(t, os.RemoveAll(filepath.Join(dir, "layers")))
	require.NoError(t, syscall.Unmount(dir, 0))

	// act
	_, err := agent.Start("miso-cache", dir)
	t.Cleanup(func() { assert.NoError(t, syscall.Unmount(dir, 0)) })

	// assert
	require.NoError(t, err)
	assert.DirExists(t, filepath.Join(dir, "layers"))
}

func TestAStartedAgentKeepsAnImportedBaseOnTheCacheDisk(t *testing.T) {
	// arrange
	digest := os.Getenv("MISO_VMTEST_BASE_DIGEST")
	require.NotEmpty(t, digest, "MISO_VMTEST_BASE names no base image")
	dir := t.TempDir()
	worker, err := agent.Start("miso-cache", dir)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, syscall.Unmount(dir, 0)) })
	require.NotNil(t, worker)

	// act
	err = worker.Import(context.Background(), protocol.Import{Key: "started", Digest: digest}, io.Discard)

	// assert
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(dir, "layers", "started", "etc", "os-release"))
}
