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
