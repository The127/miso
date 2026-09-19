package baseimage_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/baseimage"
	"github.com/The127/miso/internal/download"
)

func TestMisoKnowsTheDebianCloudImages(t *testing.T) {
	// arrange
	cache := baseimage.Open(t.TempDir(), download.Open(t.TempDir(), http.DefaultClient), baseimage.Known)

	// act
	digest, err := cache.Digest("debian:sid")

	// assert
	require.NoError(t, err)
	assert.Empty(t, digest)
}
