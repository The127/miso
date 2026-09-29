package layer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/layer"
)

func TestARemovedLayerIsNotThereAnymore(t *testing.T) {
	// arrange
	store := layer.Open(t.TempDir())
	require.NoError(t, begun(t, store, "abc", "hi").Finish())

	// act
	err := store.Remove("abc")

	// assert
	require.NoError(t, err)
	has, err := store.Has("abc")
	require.NoError(t, err)
	assert.False(t, has)
}

func TestARemovalStoppedHalfwayLeavesNoLayerUnderItsKey(t *testing.T) {
	// arrange
	store := layer.Open(t.TempDir())
	require.NoError(t, begun(t, store, "abc", "hi").Finish())

	// act
	left, err := layer.Retire(store, "abc")

	// assert
	require.NoError(t, err)
	has, err := store.Has("abc")
	require.NoError(t, err)
	assert.False(t, has)
	require.NoError(t, store.Sweep())
	assert.NoDirExists(t, left)
}
