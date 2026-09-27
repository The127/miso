package builder_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
)

func TestTheBuildersCacheDiskIsWritableUnderTheCacheSerial(t *testing.T) {
	// act
	machine := builder.Machine("/cache/disk.img")

	// assert
	assert.Equal(t, []qemu.Disk{{Path: "/cache/disk.img", Format: "raw", Serial: protocol.CacheSerial}}, machine.Disks)
}
