package emulator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/emulator"
)

func TestTheEmulatorOfAnArm64ImageIsTheHostsStaticQEMUOnThePath(t *testing.T) {
	// arrange
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "qemu-aarch64-static"), []byte("the emulator"), 0o600))
	t.Setenv("PATH", dir)

	// act
	path, found, err := emulator.Host("arm64")

	// assert
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "qemu-aarch64-static"), path)
	assert.Equal(t, []byte("the emulator"), found)
}

func TestTheEmulatorOfAnAmd64ImageIsTheHostsStaticQEMUOnThePath(t *testing.T) {
	// arrange
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "qemu-aarch64-static"), []byte("another emulator"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "qemu-x86_64-static"), []byte("the emulator"), 0o600))
	t.Setenv("PATH", dir)

	// act
	_, found, err := emulator.Host("amd64")

	// assert
	require.NoError(t, err)
	assert.Equal(t, []byte("the emulator"), found)
}

func TestAnEmulatorThatIsNotInstalledIsRefusedNamingThePackageItComesWith(t *testing.T) {
	// arrange
	t.Setenv("PATH", t.TempDir())

	// act
	_, _, err := emulator.Host("amd64")

	// assert
	require.ErrorContains(t, err, "amd64")
	assert.ErrorContains(t, err, "qemu-x86_64-static")
	assert.ErrorContains(t, err, "qemu-user-static")
}

func TestAnEmulatorOfAnArchitectureMisoHasNoQEMUForIsRefusedNamingIt(t *testing.T) {
	// arrange
	t.Setenv("PATH", t.TempDir())

	// act
	_, _, err := emulator.Host("sparc")

	// assert
	require.ErrorContains(t, err, "sparc")
	assert.NotContains(t, err.Error(), "qemu-user-static")
}
