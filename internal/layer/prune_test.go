package layer_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/layer"
	"github.com/The127/miso/internal/lru"
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
	_, err := store.Prune(lru.Policy{Now: now, OlderThan: 5 * 24 * time.Hour})

	// assert
	require.NoError(t, err)
	has, err := store.Has("old")
	require.NoError(t, err)
	assert.False(t, has)
	has, err = store.Has("new")
	require.NoError(t, err)
	assert.True(t, has)
}

func TestAPruneSaysHowManyLayersAndBytesItRemoved(t *testing.T) {
	// arrange
	store := layer.Open(t.TempDir())
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	work, err := store.Begin("old")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(work.Dir(), "big"), make([]byte, 1<<20), 0o600))
	require.NoError(t, work.Finish())
	require.NoError(t, store.Use(now.Add(-10*24*time.Hour), "old"))

	// act
	swept, err := store.Prune(lru.Policy{Now: now, OlderThan: 5 * 24 * time.Hour})

	// assert
	require.NoError(t, err)
	assert.Equal(t, 1, swept.Count)
	assert.GreaterOrEqual(t, swept.Bytes, int64(1<<20))
	assert.Less(t, swept.Bytes, int64(2<<20))
}
