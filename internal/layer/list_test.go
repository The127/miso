package layer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/layer"
)

func TestTheListHoldsOnlyFinishedLayers(t *testing.T) {
	// arrange
	store := layer.Open(t.TempDir())
	require.NoError(t, begun(t, store, "abc", "hi").Finish())
	begun(t, store, "def", "hi")
	_, err := store.Scratch()
	require.NoError(t, err)

	// act
	entries, err := store.List()

	// assert
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "abc", entries[0].Key)
}
