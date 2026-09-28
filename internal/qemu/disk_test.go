package qemu_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/qemu"
)

func TestAReadOnlyDiskIsFoundByItsSerial(t *testing.T) {
	// arrange
	machine := qemu.Machine{Disks: []qemu.Disk{{Path: "/c/base.qcow2", Format: "qcow2", Serial: "base", Access: qemu.ReadOnly}}}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "file=/c/base.qcow2,format=qcow2,if=none,id=disk0,readonly=on", valueOf(t, args, "-drive"))
	assert.Equal(t, "virtio-blk-pci,drive=disk0,serial=base", valueOf(t, args, "-device"))
}

func TestAWritableDiskKeepsTheGuestsFlushes(t *testing.T) {
	// arrange
	machine := qemu.Machine{Disks: []qemu.Disk{{Path: "/c/cache.img", Format: "raw", Serial: "miso-cache"}}}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "file=/c/cache.img,format=raw,if=none,id=disk0,cache=writeback", valueOf(t, args, "-drive"))
}

func TestACommaInADisksPathStaysPartOfThePath(t *testing.T) {
	// arrange
	machine := qemu.Machine{Disks: []qemu.Disk{{Path: "/c/a,readonly=off.img", Format: "raw", Serial: "base", Access: qemu.ReadOnly}}}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "file=/c/a,,readonly=off.img,format=raw,if=none,id=disk0,readonly=on", valueOf(t, args, "-drive"))
}

func TestACommaInADisksSerialStaysPartOfTheSerial(t *testing.T) {
	// arrange
	machine := qemu.Machine{Disks: []qemu.Disk{{Path: "/c/base.qcow2", Format: "qcow2", Serial: "a,drive=b", Access: qemu.ReadOnly}}}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "virtio-blk-pci,drive=disk0,serial=a,,drive=b", valueOf(t, args, "-device"))
}

func TestASerialLongerThanAVirtioDiskShowsIsRefused(t *testing.T) {
	// arrange
	machine := qemu.Machine{Disks: []qemu.Disk{{Path: "/c/base.qcow2", Format: "qcow2", Serial: "0123456789abcdef01234", Access: qemu.ReadOnly}}}

	// act
	_, err := qemu.Arguments(machine)

	// assert
	assert.ErrorContains(t, err, "0123456789abcdef01234")
	assert.ErrorContains(t, err, "20 bytes")
}

func TestASerialOfAllTheBytesAVirtioDiskShowsIsKept(t *testing.T) {
	// arrange
	machine := qemu.Machine{Disks: []qemu.Disk{{Path: "/c/base.qcow2", Format: "qcow2", Serial: "0123456789abcdef0123", Access: qemu.ReadOnly}}}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "virtio-blk-pci,drive=disk0,serial=0123456789abcdef0123", valueOf(t, args, "-device"))
}

func TestASnapshotDiskIsWrittenToAndForgotten(t *testing.T) {
	// arrange
	machine := qemu.Machine{Disks: []qemu.Disk{{Path: "/o/image.raw", Format: "raw", Serial: "image", Access: qemu.Snapshot}}}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "file=/o/image.raw,format=raw,if=none,id=disk0,snapshot=on", valueOf(t, args, "-drive"))
}
