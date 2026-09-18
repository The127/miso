package disk_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/disk"
)

const (
	esp        = "c12a7328-f81f-11d2-ba4b-00a0c93ec93b"
	rootX86_64 = "4f68bce3-e8cd-4db1-96e7-fbcaf984b709"
)

func TestTheRootPartitionIsFoundByItsType(t *testing.T) {
	// arrange
	image := gpt(t, entry{kind: esp, first: 34, last: 99}, entry{kind: rootX86_64, first: 100, last: 199})

	// act
	root, err := disk.Root(bytes.NewReader(image))

	// assert
	require.NoError(t, err)
	assert.Equal(t, disk.Partition{Offset: 100 * 512, Size: 100 * 512}, root)
}

func TestAPartitionPastAnyDiskIsABrokenTable(t *testing.T) {
	// arrange
	image := gpt(t, entry{kind: rootX86_64, first: 1 << 62, last: 1<<62 + 99})

	// act
	_, err := disk.Root(bytes.NewReader(image))

	// assert
	assert.ErrorIs(t, err, disk.ErrBrokenTable)
}

func TestAnEntryTooSmallForAPartitionIsABrokenTable(t *testing.T) {
	// arrange
	image := gpt(t, entry{kind: rootX86_64, first: 100, last: 199})
	binary.LittleEndian.PutUint32(image[512+84:], 0)

	// act
	_, err := disk.Root(bytes.NewReader(image))

	// assert
	assert.ErrorIs(t, err, disk.ErrBrokenTable)
}

func TestADiskWithoutARootPartitionHasNoRoot(t *testing.T) {
	// arrange
	image := gpt(t, entry{kind: esp, first: 34, last: 99})

	// act
	_, err := disk.Root(bytes.NewReader(image))

	// assert
	assert.ErrorIs(t, err, disk.ErrNoRoot)
}

func TestADiskWithoutASignatureHasNoTable(t *testing.T) {
	// arrange
	image := make([]byte, 200*512)

	// act
	_, err := disk.Root(bytes.NewReader(image))

	// assert
	assert.ErrorIs(t, err, disk.ErrNoTable)
}

type entry struct {
	kind        string
	first, last uint64
}

// gpt is a disk of 512-byte sectors with a GPT header in sector 1 and 128
// entries of 128 bytes from sector 2 on.
func gpt(t *testing.T, entries ...entry) []byte {
	t.Helper()

	image := make([]byte, 200*512)
	header := image[512:]
	copy(header, "EFI PART")
	binary.LittleEndian.PutUint64(header[72:], 2)
	binary.LittleEndian.PutUint32(header[80:], 128)
	binary.LittleEndian.PutUint32(header[84:], 128)

	for i, e := range entries {
		at := image[1024+i*128:]
		copy(at, guid(t, e.kind))
		binary.LittleEndian.PutUint64(at[32:], e.first)
		binary.LittleEndian.PutUint64(at[40:], e.last)
	}

	return image
}

// guid is the on-disk form of a GUID, whose first three fields are little
// endian.
func guid(t *testing.T, text string) []byte {
	t.Helper()

	raw, err := hex.DecodeString(strings.ReplaceAll(text, "-", ""))
	require.NoError(t, err)

	return []byte{
		raw[3], raw[2], raw[1], raw[0],
		raw[5], raw[4],
		raw[7], raw[6],
		raw[8], raw[9], raw[10], raw[11], raw[12], raw[13], raw[14], raw[15],
	}
}
