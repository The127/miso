package layer_test

import (
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
	err := store.Sweep()

	// assert
	require.NoError(t, err)
	assert.NoDirExists(t, left.Dir())
}

func TestASweepKeepsAFinishedLayer(t *testing.T) {
	// arrange
	store := layer.Open(t.TempDir())
	require.NoError(t, begun(t, store, "abc", "hi").Finish())

	// act
	err := store.Sweep()

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
	err = store.Sweep()

	// assert
	require.NoError(t, err)
	assert.NoDirExists(t, scratch)
}
