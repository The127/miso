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
