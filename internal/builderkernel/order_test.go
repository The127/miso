package builderkernel_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builderkernel"
)

func TestAModuleThatNeedsNothingLoadsOnItsOwn(t *testing.T) {
	// arrange
	have := []builderkernel.Info{{Name: "virtio_blk"}}

	// act
	loaded, err := builderkernel.Order(have, nil, "virtio_blk")

	// assert
	require.NoError(t, err)
	assert.Equal(t, []string{"virtio_blk"}, loaded)
}

func TestAModuleLoadsAfterWhatItDependsOn(t *testing.T) {
	// arrange
	have := []builderkernel.Info{
		{Name: "btrfs", Depends: []string{"libcrc32c"}},
		{Name: "libcrc32c"},
	}

	// act
	loaded, err := builderkernel.Order(have, nil, "btrfs")

	// assert
	require.NoError(t, err)
	assert.Equal(t, []string{"libcrc32c", "btrfs"}, loaded)
}

func TestAModuleReachedTwiceLoadsOnce(t *testing.T) {
	// arrange
	have := []builderkernel.Info{
		{Name: "btrfs", Depends: []string{"xor", "raid6_pq"}},
		{Name: "xor"},
		{Name: "raid6_pq", Depends: []string{"xor"}},
	}

	// act
	loaded, err := builderkernel.Order(have, nil, "btrfs")

	// assert
	require.NoError(t, err)
	assert.Equal(t, []string{"xor", "raid6_pq", "btrfs"}, loaded)
}

func TestAModuleThatIsNotThereIsRefused(t *testing.T) {
	// arrange
	have := []builderkernel.Info{{Name: "btrfs", Depends: []string{"libcrc32c"}}}

	// act
	_, err := builderkernel.Order(have, nil, "btrfs")

	// assert
	assert.ErrorContains(t, err, "libcrc32c")
}

func TestAModuleBuiltIntoTheKernelIsSkipped(t *testing.T) {
	// arrange
	have := []builderkernel.Info{{Name: "btrfs", Depends: []string{"libcrc32c"}}}
	builtin := []string{"libcrc32c"}

	// act
	loaded, err := builderkernel.Order(have, builtin, "btrfs")

	// assert
	require.NoError(t, err)
	assert.Equal(t, []string{"btrfs"}, loaded)
}
