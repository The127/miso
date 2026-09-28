//go:build kvm

package qemu_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/qemu"
)

func TestAHostWithKVMRunsTheMachineOnKVMWithTheHostsCPU(t *testing.T) {
	// arrange
	device := "/dev/kvm"

	// act
	args, why := qemu.Accel(device)

	// assert
	require.NoError(t, why)
	assert.Equal(t, "kvm", valueOf(t, args, "-accel"))
	assert.Equal(t, "host", valueOf(t, args, "-cpu"))
}
