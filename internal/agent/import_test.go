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

func TestTheRootOfABaseDiskIsMountedReadOnly(t *testing.T) {
	// arrange
	target := t.TempDir()

	// act
	err := agent.MountRoot("miso-test-base", target)

	// assert
	require.NoError(t, err)
	t.Cleanup(func() { _ = syscall.Unmount(target, 0) })
	info, err := os.Lstat(filepath.Join(target, "etc/os-release"))
	require.NoError(t, err)
	assert.Equal(t, os.ModeSymlink, info.Mode().Type())
	mounted, err := os.OpenRoot(target)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, mounted.Close()) })
	release, err := mounted.ReadFile("etc/os-release")
	require.NoError(t, err)
	assert.Contains(t, string(release), "ID=debian")
	err = os.WriteFile(filepath.Join(target, "written"), nil, 0o600)
	assert.ErrorIs(t, err, syscall.EROFS)
}

func TestABtrfsRootIsMounted(t *testing.T) {
	// arrange
	target := t.TempDir()

	// act
	err := agent.MountRoot("miso-test-flat", target)

	// assert
	require.NoError(t, err)
	t.Cleanup(func() { _ = syscall.Unmount(target, 0) })
	release, err := os.ReadFile(filepath.Join(target, "etc/os-release"))
	require.NoError(t, err)
	assert.Contains(t, string(release), "ID=miso-test")
}

func TestTheRootSubvolumeIsFoundByItsOwnFstab(t *testing.T) {
	// arrange
	target := t.TempDir()

	// act
	err := agent.MountRoot("miso-test-subvolumes", target)

	// assert
	require.NoError(t, err)
	t.Cleanup(func() { _ = syscall.Unmount(target, 0) })
	release, err := os.ReadFile(filepath.Join(target, "etc/os-release"))
	require.NoError(t, err)
	assert.Contains(t, string(release), "ID=miso-test-subvolumes")
}
