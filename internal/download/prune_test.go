package download_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/download/downloadtest"
	"github.com/The127/miso/internal/lru"
)

func TestAPruneRemovesTheStaleBlobsAndSaysWhatItRemoved(t *testing.T) {
	// arrange
	dir := t.TempDir()
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	old := downloadtest.Kept(t, dir, []byte("the old kernel"))
	fresh := downloadtest.Kept(t, dir, []byte("the new kernel"))
	store := download.Open(dir, nil)
	require.NoError(t, store.Use(now.Add(-10*24*time.Hour), old))
	require.NoError(t, store.Use(now.Add(-time.Hour), fresh))

	// act
	swept, err := store.Prune(lru.Policy{Now: now, OlderThan: 5 * 24 * time.Hour})

	// assert
	require.NoError(t, err)
	assert.Equal(t, download.Swept{Count: 1, Bytes: int64(len("the old kernel"))}, swept)
	assert.NoFileExists(t, store.Path(old))
	assert.FileExists(t, store.Path(fresh))
}
