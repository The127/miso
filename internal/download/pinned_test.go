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
)

// kept writes bytes into a store by hand and answers their digest.
func kept(t *testing.T, dir, content string) string {
	t.Helper()

	sum := sha256.Sum256([]byte(content))
	hash := hex.EncodeToString(sum[:])
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sha256"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sha256", hash), []byte(content), 0o600))

	return "sha256:" + hash
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
	digest := kept(t, dir, "the kernel")
	store := download.Open(dir, server.Client())

	// act
	path, err := store.Pinned(context.Background(), server.URL+"/kernel.deb", digest)

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
	sum := sha256.Sum256([]byte("the kernel"))
	digest := "sha256:" + hex.EncodeToString(sum[:])

	// act
	path, err := store.Pinned(context.Background(), server.URL+"/kernel.deb", digest)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "the kernel", read(t, path))
	assert.Equal(t, store.Path(digest), path)
}
