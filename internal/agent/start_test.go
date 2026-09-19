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
