package qemu_test

import (
	"testing"

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
