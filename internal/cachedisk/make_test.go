package cachedisk_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/cachedisk"
	"github.com/The127/miso/internal/disk"
)

func TestANewCacheDiskIsExt4(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.img")

	// act
	err := cachedisk.Make(path)

	// assert
	require.NoError(t, err)
	file, err := os.Open(path)
	require.NoError(t, err)

	defer func() { _ = file.Close() }()

	kind, err := disk.FileSystem(file)
	require.NoError(t, err)
	assert.Equal(t, "ext4", kind)
}
