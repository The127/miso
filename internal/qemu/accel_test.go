package qemu_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/qemu"
)

func TestAHostWithoutKVMRunsTheMachineOnTCGWithTheMostCapableCPU(t *testing.T) {
	// arrange
	missing := filepath.Join(t.TempDir(), "kvm")

	// act
	args, why := qemu.Accel(missing)

	// assert
	assert.ErrorIs(t, why, fs.ErrNotExist)
	assert.Equal(t, "tcg", valueOf(t, args, "-accel"))
	assert.Equal(t, "max", valueOf(t, args, "-cpu"))
}

func TestAHostWhoseKVMDeviceIsNoKVMRunsTheMachineOnTCG(t *testing.T) {
	// arrange
	device := filepath.Join(t.TempDir(), "kvm")
	require.NoError(t, os.WriteFile(device, nil, 0o600))

	// act
	args, why := qemu.Accel(device)

	// assert
	assert.ErrorIs(t, why, unix.ENOTTY)
	assert.Equal(t, "tcg", valueOf(t, args, "-accel"))
}
