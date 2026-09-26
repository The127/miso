package qemu_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/qemu"
)

func TestAMachineGetsTheMemoryAndCPUsItAsksFor(t *testing.T) {
	// arrange
	machine := qemu.Machine{MemoryMiB: 4096, CPUs: 2}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "4096M", valueOf(t, args, "-m"))
	assert.Equal(t, "2", valueOf(t, args, "-smp"))
}
