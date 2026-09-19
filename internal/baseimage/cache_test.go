package baseimage_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/baseimage"
	"github.com/The127/miso/internal/download"
)

func TestANameNobodyKnowsCannotBeFetched(t *testing.T) {
	// arrange
	cache := baseimage.Open(t.TempDir(), download.Open(t.TempDir(), http.DefaultClient), map[string]string{})

	// act
	_, err := cache.Digest("nope")

	// assert
	assert.ErrorIs(t, err, baseimage.ErrUnknownBase)
}

func TestAKnownImageNotFetchedYetHasNoDigest(t *testing.T) {
	// arrange
	cache := baseimage.Open(filepath.Join(t.TempDir(), "not-there-yet"), download.Open(filepath.Join(t.TempDir(), "not-there-yet"), http.DefaultClient), map[string]string{"debian:sid": "https://example.invalid/sid.qcow2"})

	// act
	digest, err := cache.Digest("debian:sid")

	// assert
	require.NoError(t, err)
	assert.Empty(t, digest)
}

func TestANameWhoseImageIsGoneHasNoDigest(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{"/sid.qcow2": "the image"})
	dir := t.TempDir()
	cache := baseimage.Open(dir, download.Open(dir, server.Client()), map[string]string{"debian:sid": server.URL + "/sid.qcow2"})
	fetched, err := cache.Fetch(context.Background(), "debian:sid")
	require.NoError(t, err)
	require.NoError(t, os.Remove(filepath.Join(dir, "sha256", strings.TrimPrefix(fetched, "sha256:"))))

	// act
	digest, err := cache.Digest("debian:sid")

	// assert
	require.NoError(t, err)
	assert.Empty(t, digest)
}
