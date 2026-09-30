package download_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/download/downloadtest"
)

func TestAPinnedFileThatIsThereIsMarkedUsed(t *testing.T) {
	// arrange
	dir := t.TempDir()
	digest := downloadtest.Kept(t, dir, []byte("the kernel"))
	store := download.Open(dir, nil)
	old := time.Now().Add(-time.Hour)
	require.NoError(t, store.Use(old, digest))

	// act
	_, err := store.Pinned(context.Background(), download.Pin{Digest: digest})

	// assert
	require.NoError(t, err)
	blobs, err := store.List()
	require.NoError(t, err)
	require.Len(t, blobs, 1)
	assert.True(t, blobs[0].Used.After(old))
}

func TestTheListSaysWhenABlobWasLastUsed(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{"/a": "hello"})
	store := download.Open(t.TempDir(), server.Client())
	digest, err := store.Get(context.Background(), server.URL+"/a")
	require.NoError(t, err)
	at := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	require.NoError(t, store.Use(at, digest))

	// act
	blobs, err := store.List()

	// assert
	require.NoError(t, err)
	require.Len(t, blobs, 1)
	assert.True(t, at.Equal(blobs[0].Used))
}
