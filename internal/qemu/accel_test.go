package qemu_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/qemu"
)

func TestAMachineRunsOnKVMWithTheHostsCPU(t *testing.T) {
	// arrange
	machine := qemu.Machine{}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Contains(t, args, "-enable-kvm")
	assert.Equal(t, "host", valueOf(t, args, "-cpu"))
}
