package builderkernel_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builderkernel"
)

func TestWhatTheKernelBuildsInIsNamedByItsModulesBuiltin(t *testing.T) {
	// arrange
	deb := packaged(t, map[string]string{
		"./lib/modules/6.12.107+deb13-cloud-amd64/modules.builtin": "kernel/fs/btrfs/btrfs.ko\nkernel/crypto/xor.ko\n",
	})

	// act
	found, err := builderkernel.Builtin(bytes.NewReader(deb))

	// assert
	require.NoError(t, err)
	assert.Equal(t, []string{"btrfs", "xor"}, found)
}
