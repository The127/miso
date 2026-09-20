package builderkernel_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builderkernel"
)

func TestAWantedModuleComesOutUnpacked(t *testing.T) {
	// arrange
	deb := wholePackage(t, map[string]string{
		"./lib/modules/" + testRelease + "/kernel/fs/btrfs/btrfs.ko.xz": packedModule(t, "name=btrfs"),
	})

	held, err := builderkernel.Read(bytes.NewReader(deb))
	require.NoError(t, err)

	// act
	load, err := builderkernel.Unpack(held, "btrfs")

	// assert
	require.NoError(t, err)
	require.Len(t, load, 1)
	assert.Equal(t, "btrfs", load[0].Name)
	assert.Equal(t, module(t, "name=btrfs"), load[0].Content)
}
