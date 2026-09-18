package disk_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/disk"
)

func TestADiskIsFoundByItsSerial(t *testing.T) {
	// arrange
	block := fstest.MapFS{
		"vda/serial": {Data: []byte("abc")},
		"vdb/serial": {Data: []byte("0123456789abcdef0123")},
	}

	// act
	name, err := disk.BySerial(block, "0123456789abcdef0123")

	// assert
	require.NoError(t, err)
	assert.Equal(t, "vdb", name)
}

func TestASerialNoDiskHasIsNoDisk(t *testing.T) {
	// arrange
	block := fstest.MapFS{
		"vda/serial": {Data: []byte("abc")},
	}

	// act
	_, err := disk.BySerial(block, "0123456789abcdef0123")

	// assert
	assert.ErrorIs(t, err, disk.ErrNoDisk)
	assert.ErrorContains(t, err, "0123456789abcdef0123")
}
