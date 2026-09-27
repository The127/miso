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

func TestMisoKnowsTheDebianCloudImagesAreInQcow2(t *testing.T) {
	// arrange
	known := baseimage.Known

	// act
	formats := []string{known["debian:sid"].Format, known["debian:13"].Format, known["debian:trixie"].Format}

	// assert
	assert.Equal(t, []string{"qcow2", "qcow2", "qcow2"}, formats)
}
