package qemu_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/qemu"
)

func TestAMachinesCardHasTheMACItAsksForOnQEMUsUserNetwork(t *testing.T) {
	// arrange
	machine := qemu.Machine{Card: &qemu.Card{MAC: "52:54:00:6d:69:73"}}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "user,id=card", valueOf(t, args, "-netdev"))
	assert.Equal(t, "virtio-net-pci,netdev=card,mac=52:54:00:6d:69:73", valueOf(t, args, "-device"))
}
