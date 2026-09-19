package baseimage_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/baseimage"
	"github.com/The127/miso/internal/download"
)

func TestAFetchDownloadsAnImageAndAnswersTheDigestOfItsBytes(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{"/sid.qcow2": "the image"})
	cache := baseimage.Open(t.TempDir(), download.Open(t.TempDir(), server.Client()), map[string]string{"debian:sid": server.URL + "/sid.qcow2"})
	sum := sha256.Sum256([]byte("the image"))

	// act
	digest, err := cache.Fetch(context.Background(), "debian:sid")

	// assert
	require.NoError(t, err)
	assert.Equal(t, "sha256:"+hex.EncodeToString(sum[:]), digest)
}

func TestAFetchedImageIsKnownToACacheOpenedLaterOnTheSameDirectory(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{"/sid.qcow2": "the image"})
	sources := map[string]string{"debian:sid": server.URL + "/sid.qcow2"}
	dir := t.TempDir()
	fetched, err := baseimage.Open(dir, download.Open(dir, server.Client()), sources).Fetch(context.Background(), "debian:sid")
	require.NoError(t, err)

	// act
	digest, err := baseimage.Open(dir, download.Open(dir, server.Client()), sources).Digest("debian:sid")

	// assert
	require.NoError(t, err)
	assert.Equal(t, fetched, digest)
}

func TestAFetchThatIsRefusedIsAnErrorAndLeavesNothingBehind(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{})
	dir := t.TempDir()
	cache := baseimage.Open(dir, download.Open(dir, server.Client()), map[string]string{"debian:sid": server.URL + "/gone.qcow2"})

	// act
	_, err := cache.Fetch(context.Background(), "debian:sid")

	// assert
	assert.ErrorContains(t, err, server.URL+"/gone.qcow2")
	assert.ErrorContains(t, err, "404 Not Found")
	digest, digestErr := cache.Digest("debian:sid")
	require.NoError(t, digestErr)
	assert.Empty(t, digest)
	assert.NoDirExists(t, filepath.Join(dir, "sha256"))
}

func TestADownloadThatStopsShortIsNotKept(t *testing.T) {
	// arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "1000")
		_, _ = w.Write([]byte("short"))
	}))
	t.Cleanup(server.Close)
	dir := t.TempDir()
	cache := baseimage.Open(dir, download.Open(dir, server.Client()), map[string]string{"debian:sid": server.URL + "/sid.qcow2"})

	// act
	_, err := cache.Fetch(context.Background(), "debian:sid")

	// assert
	assert.Error(t, err)
	digest, digestErr := cache.Digest("debian:sid")
	require.NoError(t, digestErr)
	assert.Empty(t, digest)
	leftovers, _ := filepath.Glob(filepath.Join(dir, "sha256", "*"))
	assert.Empty(t, leftovers)
}

func TestAFetchAgainReplacesWhatANamePointsAt(t *testing.T) {
	// arrange
	images := map[string]string{"/sid.qcow2": "yesterday's image"}
	server := serving(t, images)
	cache := baseimage.Open(t.TempDir(), download.Open(t.TempDir(), server.Client()), map[string]string{"debian:sid": server.URL + "/sid.qcow2"})
	_, err := cache.Fetch(context.Background(), "debian:sid")
	require.NoError(t, err)
	images["/sid.qcow2"] = "today's image"
	sum := sha256.Sum256([]byte("today's image"))

	// act
	_, err = cache.Fetch(context.Background(), "debian:sid")

	// assert
	require.NoError(t, err)
	digest, err := cache.Digest("debian:sid")
	require.NoError(t, err)
	assert.Equal(t, "sha256:"+hex.EncodeToString(sum[:]), digest)
}

func TestACancelledFetchStops(t *testing.T) {
	// arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)
	cache := baseimage.Open(t.TempDir(), download.Open(t.TempDir(), server.Client()), map[string]string{"debian:sid": server.URL + "/sid.qcow2"})
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)

	// act
	_, err := cache.Fetch(ctx, "debian:sid")

	// assert
	assert.ErrorIs(t, err, context.Canceled)
}
