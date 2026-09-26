package qemu_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/qemu"
)

func TestAMachineBootsItsKernelDirectly(t *testing.T) {
	// arrange
	machine := qemu.Machine{Kernel: "/k/vmlinuz", Initramfs: "/k/initrd", CommandLine: "console=ttyS0 -- agent"}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "/k/vmlinuz", valueOf(t, args, "-kernel"))
	assert.Equal(t, "/k/initrd", valueOf(t, args, "-initrd"))
	assert.Equal(t, "console=ttyS0 -- agent", valueOf(t, args, "-append"))
}
