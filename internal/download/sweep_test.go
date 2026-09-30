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

func TestASweepOfAStoreThatHoldsNothingYetRemovesNothing(t *testing.T) {
	// arrange
	store := download.Open(t.TempDir(), nil)

	// act
	swept, err := store.Sweep()

	// assert
	require.NoError(t, err)
	assert.Equal(t, download.Swept{}, swept)
}

func TestASweepRemovesADownloadThatNeverArrivedAndKeepsTheBlobs(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{"/a": "hello"})
	dir := t.TempDir()
	store := download.Open(dir, server.Client())
	digest, err := store.Get(context.Background(), server.URL+"/a")
	require.NoError(t, err)
	left := filepath.Join(dir, "sha256", "download-123")
	require.NoError(t, os.WriteFile(left, []byte("half"), 0o600))

	// act
	swept, err := store.Sweep()

	// assert
	require.NoError(t, err)
	assert.NoFileExists(t, left)
	assert.FileExists(t, store.Path(digest))
	assert.Equal(t, download.Swept{Count: 1, Bytes: 4}, swept)
}
