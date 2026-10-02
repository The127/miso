package cachedisk_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/cachedisk"
)

func TestTheDiskOfACacheDirectoryIsLayersImgInIt(t *testing.T) {
	// act
	disk := cachedisk.Dir("/cache/builder").Disk()

	// assert
	assert.Equal(t, "/cache/builder/layers.img", disk)
}

func TestTheLockOfACacheDirectoryIsLayersLockInIt(t *testing.T) {
	// act
	lock := cachedisk.Dir("/cache/builder").LockFile()

	// assert
	assert.Equal(t, "/cache/builder/layers.lock", lock)
}

func TestTheBuilderLogOfACacheDirectoryIsBuilderLogInIt(t *testing.T) {
	// act
	log := cachedisk.Dir("/cache/builder").BuilderLog()

	// assert
	assert.Equal(t, "/cache/builder/builder.log", log)
}

func TestTheCheckLogOfACacheDirectoryIsCheckLogInIt(t *testing.T) {
	// act
	log := cachedisk.Dir("/cache/builder").CheckLog()

	// assert
	assert.Equal(t, "/cache/builder/check.log", log)
}
