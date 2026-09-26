package qemu_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/qemu"
)

func TestAReadOnlyDiskIsFoundByItsSerial(t *testing.T) {
	// arrange
	machine := qemu.Machine{Disks: []qemu.Disk{{Path: "/c/base.qcow2", Format: "qcow2", Serial: "base", ReadOnly: true}}}

	// act
	args := qemu.Arguments(machine)

	// assert
	assert.Equal(t, "file=/c/base.qcow2,format=qcow2,if=none,id=disk0,readonly=on", valueOf(t, args, "-drive"))
	assert.Equal(t, "virtio-blk-pci,drive=disk0,serial=base", valueOf(t, args, "-device"))
}
