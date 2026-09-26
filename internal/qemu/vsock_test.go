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
	args := qemu.Vsock(cid, fd)

	// assert
	assert.Equal(t, "vhost-vsock-pci,guest-cid=1234,vhostfd=3", valueOf(t, args, "-device"))
}
