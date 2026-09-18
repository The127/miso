package disk_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/disk"
)

func TestAnExt4IsKnownByItsMagic(t *testing.T) {
	// arrange
	partition := make([]byte, 4096)
	binary.LittleEndian.PutUint16(partition[1080:], 0xef53)

	// act
	kind, err := disk.FileSystem(bytes.NewReader(partition))

	// assert
	require.NoError(t, err)
	assert.Equal(t, "ext4", kind)
}

func TestABtrfsIsKnownByItsMagic(t *testing.T) {
	// arrange
	partition := make([]byte, 128*1024)
	copy(partition[65600:], "_BHRfS_M")

	// act
	kind, err := disk.FileSystem(bytes.NewReader(partition))

	// assert
	require.NoError(t, err)
	assert.Equal(t, "btrfs", kind)
}

func TestAPartitionWithoutAMagicIsAnUnknownFileSystem(t *testing.T) {
	// arrange
	partition := make([]byte, 128*1024)

	// act
	_, err := disk.FileSystem(bytes.NewReader(partition))

	// assert
	assert.ErrorIs(t, err, disk.ErrUnknownFileSystem)
}

func TestAPartitionTooSmallForAMagicIsAnUnknownFileSystem(t *testing.T) {
	// arrange
	partition := make([]byte, 4096)

	// act
	_, err := disk.FileSystem(bytes.NewReader(partition))

	// assert
	assert.ErrorIs(t, err, disk.ErrUnknownFileSystem)
}
