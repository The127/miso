package qemu_test

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/qemu"
)

func TestACIDOneMachineHoldsIsRefusedToTheNext(t *testing.T) {
	// arrange
	first, err := hostVsock()
	require.NoError(t, err)
	t.Cleanup(func() { _ = first.Close() })
	second, err := hostVsock()
	require.NoError(t, err)
	t.Cleanup(func() { _ = second.Close() })
	// random, so a VM running on this host is left alone
	cid := 3 + rand.Uint32N(1<<31) //nolint:gosec // a CID keeps no secret
	require.NoError(t, qemu.TakeCID(first, cid))

	// act
	err = qemu.TakeCID(second, cid)

	// assert
	assert.ErrorIs(t, err, unix.EADDRINUSE)
}
