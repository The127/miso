package initramfs_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/u-root/u-root/pkg/cpio"

	"github.com/The127/miso/internal/initramfs"
)

// record is the file of a name in an archive, which must hold it.
func record(t *testing.T, archive []byte, name string) cpio.Record {
	t.Helper()

	records, err := cpio.ReadAllRecords(cpio.Newc.Reader(bytes.NewReader(archive)))
	require.NoError(t, err)
	for _, found := range records {
		if found.Name == name {
			return found
		}
	}

	require.Failf(t, "not in the archive", "no %s in the archive", name)

	return cpio.Record{}
}

func TestAnInitramfsHoldsItsInitAsAnExecutableInit(t *testing.T) {
	// arrange
	var archive bytes.Buffer

	// act
	err := initramfs.Write(&archive, []byte("the init"), nil)

	// assert
	require.NoError(t, err)
	init := record(t, archive.Bytes(), "init")
	assert.Equal(t, uint64(cpio.S_IFREG|0o700), init.Mode)
	content := make([]byte, init.FileSize)
	_, err = init.ReadAt(content, 0)
	require.NoError(t, err)
	assert.Equal(t, "the init", string(content))
}

func TestAnInitramfsHoldsTheConsole(t *testing.T) {
	// arrange
	var archive bytes.Buffer

	// act
	err := initramfs.Write(&archive, []byte("the init"), nil)

	// assert
	require.NoError(t, err)
	console := record(t, archive.Bytes(), "dev/console")
	assert.Equal(t, uint64(cpio.S_IFCHR|0o600), console.Mode)
	assert.Equal(t, uint64(5), console.Rmajor)
	assert.Equal(t, uint64(1), console.Rminor)
}

func TestAnInitramfsHoldsItsModulesInTheOrderToLoadThem(t *testing.T) {
	// arrange
	var archive bytes.Buffer
	modules := []initramfs.Module{
		{Name: "virtio_blk", Content: []byte("first")},
		{Name: "btrfs", Content: []byte("second")},
	}

	// act
	err := initramfs.Write(&archive, []byte("the init"), modules)

	// assert
	require.NoError(t, err)
	first := record(t, archive.Bytes(), "modules/01-virtio_blk.ko")
	second := record(t, archive.Bytes(), "modules/02-btrfs.ko")
	assert.Equal(t, uint64(5), first.FileSize)
	assert.Equal(t, uint64(6), second.FileSize)
}
