//go:build vmtest

package agent_test

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/kernel"
	"github.com/The127/miso/internal/protocol"
)

func TestADiskOfAnImageWithoutAKernelFailsNamingIt(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	strip := protocol.Run{Key: "stripped", Layers: []string{"base"}, Command: "rm -r /boot/vmlinuz-* /usr/lib/modules"}
	code, err := worker.Run(context.Background(), strip, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	disk := protocol.Disk{Key: "disk", Layers: []string{"base", "stripped"}, Tools: []string{"base"}}

	// act
	err = worker.Disk(context.Background(), disk, io.Discard)

	// assert
	assert.ErrorIs(t, err, kernel.ErrNoKernel)
}

func TestADiskOfAnImageWithAKernelFailsBecauseTheAgentMakesNoDisksYet(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	disk := protocol.Disk{Key: "disk", Layers: []string{"base"}, Tools: []string{"base"}}

	// act
	err := worker.Disk(context.Background(), disk, io.Discard)

	// assert
	assert.EqualError(t, err, "the agent makes no disks yet")
}
