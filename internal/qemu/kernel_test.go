package qemu_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/qemu"
)

func TestAMachineBootsItsKernelDirectly(t *testing.T) {
	// arrange
	machine := qemu.Machine{Kernel: "/k/vmlinuz", Initramfs: "/k/initrd", CommandLine: "console=ttyS0 -- agent"}

	// act
	args := qemu.Arguments(machine)

	// assert
	assert.Equal(t, []string{"-kernel", "/k/vmlinuz", "-initrd", "/k/initrd", "-append", "console=ttyS0 -- agent"}, args)
}
