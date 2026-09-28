package download_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/download/downloadtest"
)

// arrived is the digest bytes have once they are in.
func arrived(t *testing.T, content string) string {
	t.Helper()

	sum := sha256.Sum256([]byte(content))

	return "sha256:" + hex.EncodeToString(sum[:])
}

func read(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	return string(content)
}

func TestAPinnedFileThatIsThereIsAnsweredWithoutARequest(t *testing.T) {
	// arrange
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests++
	}))
	t.Cleanup(server.Close)
	dir := t.TempDir()
	digest := downloadtest.Kept(t, dir, []byte("the kernel"))
	store := download.Open(dir, server.Client())

	// act
	path, err := store.Pinned(context.Background(), download.Pin{URL: server.URL + "/kernel.deb", Digest: digest})

	// assert
	require.NoError(t, err)
	assert.Equal(t, "the kernel", read(t, path))
	assert.Zero(t, requests)
}

func TestAPinnedFileThatIsNotThereIsDownloadedAndKept(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{"/kernel.deb": "the kernel"})
	dir := t.TempDir()
	store := download.Open(dir, server.Client())
	digest := arrived(t, "the kernel")

	// act
	path, err := store.Pinned(context.Background(), download.Pin{URL: server.URL + "/kernel.deb", Digest: digest})

	// assert
	require.NoError(t, err)
	assert.Equal(t, "the kernel", read(t, path))
	assert.Equal(t, store.Path(digest), path)
}

func TestADownloadThatDoesNotMatchThePinIsRefused(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{"/kernel.deb": "another kernel"})
	dir := t.TempDir()
	store := download.Open(dir, server.Client())
	digest := arrived(t, "the kernel")

	// act
	_, err := store.Pinned(context.Background(), download.Pin{URL: server.URL + "/kernel.deb", Digest: digest})

	// assert
	require.Error(t, err)
	assert.ErrorContains(t, err, server.URL+"/kernel.deb")
	assert.ErrorContains(t, err, digest)
	assert.ErrorContains(t, err, arrived(t, "another kernel"))
	kept, _ := filepath.Glob(filepath.Join(dir, "sha256", "*"))
	assert.Empty(t, kept)
}
