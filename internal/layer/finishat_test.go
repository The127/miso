package layer_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/layer"
)

func TestALayerFinishedAtATimeIsUsedAtThatTime(t *testing.T) {
	// arrange
	store := layer.Open(t.TempDir())
	work := begun(t, store, "abc", "hi")
	at := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	// act
	err := work.FinishAt(at)

	// assert
	require.NoError(t, err)
	entries, err := store.List()
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.True(t, at.Equal(entries[0].Used))
}
