package qemu_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/qemu"
)

func TestAMachineGetsTheMemoryAndCPUsItAsksFor(t *testing.T) {
	// arrange
	machine := qemu.Machine{MemoryMiB: 4096, CPUs: 2}

	// act
	args := qemu.Arguments(machine)

	// assert
	assert.Equal(t, "4096M", valueOf(t, args, "-m"))
	assert.Equal(t, "2", valueOf(t, args, "-smp"))
}
