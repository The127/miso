package qemu_test

import (
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

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
