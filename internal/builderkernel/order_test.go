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
