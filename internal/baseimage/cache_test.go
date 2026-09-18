package baseimage_test

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/baseimage"
)

func TestANameNobodyKnowsCannotBeFetched(t *testing.T) {
	// arrange
	cache := baseimage.Open(t.TempDir(), http.DefaultClient, map[string]string{})

	// act
	_, err := cache.Digest("nope")

	// assert
	assert.ErrorIs(t, err, baseimage.ErrUnknownBase)
	assert.ErrorContains(t, err, "nope")
}

func TestAKnownImageNotFetchedYetHasNoDigest(t *testing.T) {
	// arrange
	cache := baseimage.Open(filepath.Join(t.TempDir(), "not-there-yet"), http.DefaultClient, map[string]string{"debian:sid": "https://example.invalid/sid.qcow2"})

	// act
	digest, err := cache.Digest("debian:sid")

	// assert
	require.NoError(t, err)
	assert.Empty(t, digest)
}
