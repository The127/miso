//go:build vmtest

package place_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/place"
)

func TestAHoleInACopiedFileStaysAHole(t *testing.T) {
	// arrange
	root := t.TempDir()
	sparse, err := os.Create(filepath.Join(t.TempDir(), "sparse"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = sparse.Close() })
	_, err = sparse.WriteAt([]byte("x"), 64<<20)
	require.NoError(t, err)

	// act
	err = place.Open(root).Sparse("/sparse", 0o644, sparse)

	// assert
	require.NoError(t, err)
	var copied unix.Stat_t
	require.NoError(t, unix.Stat(filepath.Join(root, "sparse"), &copied))
	assert.Equal(t, int64(64<<20+1), copied.Size)
	assert.Less(t, copied.Blocks*512, int64(1<<20), "the hole takes room")
}
