package download_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/download"
)

func TestAStoreThatHoldsNothingYetListsNothing(t *testing.T) {
	// arrange
	store := download.Open(t.TempDir(), nil)

	// act
	blobs, err := store.List()

	// assert
	require.NoError(t, err)
	assert.Empty(t, blobs)
}

func TestTheListHoldsTheBlobsItHasWithTheirDigestAndSize(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{"/a": "hello"})
	dir := t.TempDir()
	store := download.Open(dir, server.Client())
	digest, err := store.Get(context.Background(), server.URL+"/a")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sha256", "download-123"), []byte("half"), 0o600))

	// act
	blobs, err := store.List()

	// assert
	require.NoError(t, err)
	require.Len(t, blobs, 1)
	assert.Equal(t, digest, blobs[0].Digest)
	assert.Equal(t, int64(5), blobs[0].Size)
}
