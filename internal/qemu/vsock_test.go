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
	args := qemu.Vsock(cid, fd, false)

	// assert
	assert.Equal(t, "vhost-vsock-pci,guest-cid=1234,vhostfd=3", valueOf(t, args, "-device"))
}

func TestAMicrovmReachesItsHostOverAVsockDeviceWithoutPCI(t *testing.T) {
	// arrange
	cid, fd := uint32(1234), 3

	// act
	args := qemu.Vsock(cid, fd, true)

	// assert
	assert.Equal(t, "vhost-vsock-device,guest-cid=1234,vhostfd=3", valueOf(t, args, "-device"))
}

func TestTheVirtioPortOfAMicrovmHasAControllerWithoutPCI(t *testing.T) {
	// act
	args := qemu.Port(3, true)

	// assert
	assert.Contains(t, args, "virtio-serial-device")
}

func TestTheVirtioPortOfOtherMachinesHasAPCIController(t *testing.T) {
	// act
	args := qemu.Port(3, false)

	// assert
	assert.Contains(t, args, "virtio-serial-pci")
}
