//go:build vmtest

package agent_test

import (
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/agent"
)

func TestAStagesMountLetsNoFileActAsDeviceSetuidOrProgram(t *testing.T) {
	// arrange
	root := t.TempDir()
	lowers := []string{t.TempDir(), t.TempDir()}

	// act
	err := agent.MountReadOnly(root, lowers)

	// assert
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, syscall.Unmount(root, syscall.MNT_DETACH)) })

	var mounted unix.Statfs_t
	require.NoError(t, unix.Statfs(root, &mounted))
	assert.NotZero(t, mounted.Flags&unix.ST_NODEV, "nodev")
	assert.NotZero(t, mounted.Flags&unix.ST_NOSUID, "nosuid")
	assert.NotZero(t, mounted.Flags&unix.ST_NOEXEC, "noexec")
}
