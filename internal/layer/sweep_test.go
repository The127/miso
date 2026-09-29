package layer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/layer"
)

func TestASweepRemovesALayerLeftUnfinished(t *testing.T) {
	// arrange
	store := layer.Open(t.TempDir())
	left := begun(t, store, "def", "hi")

	// act
	_, err := store.Sweep()

	// assert
	require.NoError(t, err)
	assert.NoDirExists(t, left.Dir())
}

func TestASweepSaysWhatItRemoved(t *testing.T) {
	// arrange
	store := layer.Open(t.TempDir())
	left := begun(t, store, "def", "hi")
	require.NoError(t, os.WriteFile(filepath.Join(left.Dir(), "big"), make([]byte, 1<<20), 0o600))
	_, err := store.Scratch()
	require.NoError(t, err)

	// act
	swept, err := store.Sweep()

	// assert
	require.NoError(t, err)
	assert.Equal(t, 2, swept.Count)
	assert.GreaterOrEqual(t, swept.Bytes, int64(1<<20))
}

func TestASweepKeepsAFinishedLayer(t *testing.T) {
	// arrange
	store := layer.Open(t.TempDir())
	require.NoError(t, begun(t, store, "abc", "hi").Finish())

	// act
	_, err := store.Sweep()

	// assert
	require.NoError(t, err)
	has, err := store.Has("abc")
	require.NoError(t, err)
	assert.True(t, has)
}

func TestASweepRemovesAScratchLeftBehind(t *testing.T) {
	// arrange
	store := layer.Open(t.TempDir())
	scratch, err := store.Scratch()
	require.NoError(t, err)

	// act
	_, err = store.Sweep()

	// assert
	require.NoError(t, err)
	assert.NoDirExists(t, scratch)
}
