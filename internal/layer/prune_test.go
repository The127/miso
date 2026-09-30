package layer_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/layer"
)

func TestAPruneRemovesTheStaleLayersAndKeepsTheRest(t *testing.T) {
	// arrange
	store := layer.Open(t.TempDir())
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	require.NoError(t, begun(t, store, "old", "hi").Finish())
	require.NoError(t, begun(t, store, "new", "hi").Finish())
	require.NoError(t, store.Use(now.Add(-10*24*time.Hour), "old"))
	require.NoError(t, store.Use(now.Add(-time.Hour), "new"))

	// act
	err := store.Prune(layer.Policy{Now: now, OlderThan: 5 * 24 * time.Hour})

	// assert
	require.NoError(t, err)
	has, err := store.Has("old")
	require.NoError(t, err)
	assert.False(t, has)
	has, err = store.Has("new")
	require.NoError(t, err)
	assert.True(t, has)
}
