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
	"github.com/The127/miso/internal/download/downloadtest"
)

func TestANameNobodyKnowsCannotBeFetched(t *testing.T) {
	// arrange
	cache := baseimage.Open(t.TempDir(), download.Open(t.TempDir(), http.DefaultClient), map[string]baseimage.Source{})

	// act
	_, err := cache.Digest("nope")

	// assert
	assert.ErrorIs(t, err, baseimage.ErrUnknownBase)
}

func TestAKnownImageNotFetchedYetHasNoDigest(t *testing.T) {
	// arrange
	cache := baseimage.Open(filepath.Join(t.TempDir(), "not-there-yet"), download.Open(filepath.Join(t.TempDir(), "not-there-yet"), http.DefaultClient), map[string]baseimage.Source{"debian:sid": {URL: "https://example.invalid/sid.qcow2"}})

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
	cache := baseimage.Open(dir, download.Open(dir, server.Client()), map[string]baseimage.Source{"debian:sid": {URL: server.URL + "/sid.qcow2", Format: "raw"}})
	fetched, err := cache.Fetch(context.Background(), "debian:sid")
	require.NoError(t, err)
	require.NoError(t, os.Remove(filepath.Join(dir, "sha256", strings.TrimPrefix(fetched, "sha256:"))))

	// act
	digest, err := cache.Digest("debian:sid")

	// assert
	require.NoError(t, err)
	assert.Empty(t, digest)
}

func TestAnArm64CacheKeepsItsNamesApartFromTheAmd64Ones(t *testing.T) {
	// arrange
	dir := t.TempDir()
	digest := downloadtest.Kept(t, dir, []byte("the amd64 image"))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "names"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "names", "debian:sid"), []byte(digest), 0o600))
	blobs := download.Open(dir, http.DefaultClient)

	// act
	amd64, err := baseimage.OpenFor(dir, "amd64", blobs).Digest("debian:sid")
	require.NoError(t, err)
	arm64, err := baseimage.OpenFor(dir, "arm64", blobs).Digest("debian:sid")
	require.NoError(t, err)

	// assert
	assert.Equal(t, digest, amd64)
	assert.Empty(t, arm64)
}
