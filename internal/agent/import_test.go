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
	t.Cleanup(func() { _ = syscall.Unmount(target, syscall.MNT_DETACH) })
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
	t.Cleanup(func() { _ = syscall.Unmount(target, syscall.MNT_DETACH) })
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
	t.Cleanup(func() { _ = syscall.Unmount(target, syscall.MNT_DETACH) })
	release, err := os.ReadFile(filepath.Join(target, "etc/os-release"))
	require.NoError(t, err)
	assert.Contains(t, string(release), "ID=miso-test-subvolumes")
}

func TestTheSubmountsOfTheRootAreMountedInIt(t *testing.T) {
	// arrange
	target := t.TempDir()

	// act
	err := agent.MountRoot("miso-test-subvolumes", target)

	// assert
	require.NoError(t, err)
	t.Cleanup(func() { _ = syscall.Unmount(target, syscall.MNT_DETACH) })
	variable, err := os.ReadFile(filepath.Join(target, "var/marker"))
	require.NoError(t, err)
	assert.Equal(t, "var\n", string(variable))
	boot, err := os.ReadFile(filepath.Join(target, "boot/marker"))
	require.NoError(t, err)
	assert.Equal(t, "boot\n", string(boot))
}

func TestTheSubmountsOfADefaultRootAreMountedInIt(t *testing.T) {
	// arrange
	target := t.TempDir()

	// act
	err := agent.MountRoot("miso-test-default", target)

	// assert
	require.NoError(t, err)
	t.Cleanup(func() { _ = syscall.Unmount(target, syscall.MNT_DETACH) })
	release, err := os.ReadFile(filepath.Join(target, "etc/os-release"))
	require.NoError(t, err)
	assert.Contains(t, string(release), "ID=miso-test-default")
	variable, err := os.ReadFile(filepath.Join(target, "var/marker"))
	require.NoError(t, err)
	assert.Equal(t, "var\n", string(variable))
}

func TestASubmountThroughALinkOutOfTheImageIsRefused(t *testing.T) {
	// arrange
	target := t.TempDir()
	require.NoError(t, os.Mkdir("/escaped", 0o700))
	t.Cleanup(func() {
		_ = syscall.Unmount("/escaped", syscall.MNT_DETACH)
		_ = os.Remove("/escaped")
	})

	// act
	err := agent.MountRoot("miso-test-escape", target)

	// assert
	t.Cleanup(func() { _ = syscall.Unmount(target, syscall.MNT_DETACH) })
	require.Error(t, err)
	assert.NoFileExists(t, "/escaped/marker")
}
