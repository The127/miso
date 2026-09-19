package initramfs_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/u-root/u-root/pkg/cpio"

	"github.com/The127/miso/internal/initramfs"
)

func TestAnInitramfsHoldsItsInitAsAnExecutableInit(t *testing.T) {
	// arrange
	var archive bytes.Buffer

	// act
	err := initramfs.Write(&archive, []byte("the init"))

	// assert
	require.NoError(t, err)
	records, err := cpio.ReadAllRecords(cpio.Newc.Reader(bytes.NewReader(archive.Bytes())))
	require.NoError(t, err)
	var init *cpio.Record
	for i := range records {
		if records[i].Name == "init" {
			init = &records[i]
		}
	}

	require.NotNil(t, init, "no init in the archive")
	assert.Equal(t, uint64(cpio.S_IFREG|0o700), init.Mode)
	content := make([]byte, init.FileSize)
	_, err = init.ReadAt(content, 0)
	require.NoError(t, err)
	assert.Equal(t, "the init", string(content))
}
