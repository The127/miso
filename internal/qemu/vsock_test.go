package qemu_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/qemu"
)

func TestAMachineReachesItsHostOverVsockAtTheCIDItWasGiven(t *testing.T) {
	// arrange
	cid, fd := uint32(1234), 3

	// act
	args := qemu.Vsock(cid, fd, qemu.Machine{})

	// assert
	assert.Equal(t, "vhost-vsock-pci,guest-cid=1234,vhostfd=3", valueOf(t, args, "-device"))
}

func TestAMicrovmReachesItsHostOverAVsockDeviceWithoutPCI(t *testing.T) {
	// arrange
	cid, fd := uint32(1234), 3

	// act
	args := qemu.Vsock(cid, fd, qemu.Machine{Microvm: true})

	// assert
	assert.Equal(t, "vhost-vsock-device,guest-cid=1234,vhostfd=3", valueOf(t, args, "-device"))
}

func TestAnArm64MachineReachesItsHostOverAVsockDeviceOnPCI(t *testing.T) {
	// arrange
	cid, fd := uint32(1234), 3

	// act
	args := qemu.Vsock(cid, fd, qemu.Machine{Arch: "arm64"})

	// assert
	assert.Equal(t, "vhost-vsock-pci,guest-cid=1234,vhostfd=3", valueOf(t, args, "-device"))
}

func TestTheVirtioPortOfAMicrovmHasAControllerWithoutPCI(t *testing.T) {
	// act
	args := qemu.Port(3, qemu.Machine{Microvm: true})

	// assert
	assert.Contains(t, args, "virtio-serial-device")
}

func TestTheVirtioPortOfOtherMachinesHasAPCIController(t *testing.T) {
	// act
	args := qemu.Port(3, qemu.Machine{})

	// assert
	assert.Contains(t, args, "virtio-serial-pci")
}

func TestTheVirtioPortOfAnArm64MachineHasAPCIController(t *testing.T) {
	// act
	args := qemu.Port(3, qemu.Machine{Arch: "arm64"})

	// assert
	assert.Contains(t, args, "virtio-serial-pci")
}
