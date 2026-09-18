package disk_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/disk"
)

func TestAPartitionIsFoundByItsNumber(t *testing.T) {
	// arrange
	block := fstest.MapFS{
		"vda/serial":         {Data: []byte("abc")},
		"vda/vda1/partition": {Data: []byte("1\n")},
		"vda/vda2/partition": {Data: []byte("2\n")},
	}

	// act
	name, err := disk.PartitionName(block, "vda", 2)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "vda2", name)
}
