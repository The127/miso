package qemu_test

import (
	"testing"

	"golang.org/x/sys/unix"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/qemu"
)

func TestAMachineGetsTheCIDTheDeviceTakes(t *testing.T) {
	// arrange
	take := func(uint32) error { return nil }
	random := func() uint32 { return 7 }

	// act
	cid, err := qemu.Claim(take, random)

	// assert
	require.NoError(t, err)
	assert.Equal(t, uint32(7), cid)
}

func TestACIDAnotherMachineHoldsIsPassedOver(t *testing.T) {
	// arrange
	take := func(cid uint32) error {
		if cid == 7 {
			return unix.EADDRINUSE
		}

		return nil
	}
	candidates := []uint32{7, 9}
	random := func() uint32 {
		next := candidates[0]
		candidates = candidates[1:]

		return next
	}

	// act
	cid, err := qemu.Claim(take, random)

	// assert
	require.NoError(t, err)
	assert.Equal(t, uint32(9), cid)
}

func TestAMachineThatFindsNoFreeCIDGivesUpAfterTenTries(t *testing.T) {
	// arrange
	tries := 0
	take := func(uint32) error {
		tries++

		return unix.EADDRINUSE
	}
	random := func() uint32 { return 7 }

	// act
	_, err := qemu.Claim(take, random)

	// assert
	assert.Equal(t, 10, tries)
	assert.ErrorIs(t, err, unix.EADDRINUSE)
	assert.ErrorContains(t, err, "after 10 tries")
}

func TestAMachineIsNeverGivenACIDThatMeansSomethingElse(t *testing.T) {
	// arrange
	var tried []uint32
	take := func(cid uint32) error {
		tried = append(tried, cid)

		return nil
	}
	candidates := []uint32{unix.VMADDR_CID_HYPERVISOR, unix.VMADDR_CID_LOCAL, unix.VMADDR_CID_HOST, unix.VMADDR_CID_ANY, 7}
	random := func() uint32 {
		next := candidates[0]
		candidates = candidates[1:]

		return next
	}

	// act
	cid, err := qemu.Claim(take, random)

	// assert
	require.NoError(t, err)
	assert.Equal(t, uint32(7), cid)
	assert.Equal(t, []uint32{7}, tried)
}
