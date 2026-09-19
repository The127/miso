package builderkernel_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/builderkernel"
)

func TestAModuleThatNeedsNothingLoadsOnItsOwn(t *testing.T) {
	// arrange
	have := []builderkernel.Info{{Name: "virtio_blk"}}

	// act
	loaded := builderkernel.Order(have, "virtio_blk")

	// assert
	assert.Equal(t, []string{"virtio_blk"}, loaded)
}

func TestAModuleLoadsAfterWhatItDependsOn(t *testing.T) {
	// arrange
	have := []builderkernel.Info{
		{Name: "btrfs", Depends: []string{"libcrc32c"}},
		{Name: "libcrc32c"},
	}

	// act
	loaded := builderkernel.Order(have, "btrfs")

	// assert
	assert.Equal(t, []string{"libcrc32c", "btrfs"}, loaded)
}
