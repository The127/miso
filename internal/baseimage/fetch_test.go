package baseimage_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

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
