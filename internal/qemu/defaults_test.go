package qemu_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/qemu"
)

func TestAMachineHasNoneOfQEMUsDefaultDevices(t *testing.T) {
	// arrange
	machine := qemu.Machine{}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Contains(t, args, "-nodefaults")
}

func TestAMachineReadsNoneOfTheHostsQEMUConfig(t *testing.T) {
	// arrange
	machine := qemu.Machine{}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Contains(t, args, "-no-user-config")
}

func TestAMachineIsTheSameBoardOnEveryHost(t *testing.T) {
	// arrange
	machine := qemu.Machine{}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "q35", valueOf(t, args, "-machine"))
}
