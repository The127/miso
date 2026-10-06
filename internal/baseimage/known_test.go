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
	cache := baseimage.Open(t.TempDir(), download.Open(t.TempDir(), http.DefaultClient), baseimage.Known["amd64"])

	// act
	digest, err := cache.Digest("debian:sid")

	// assert
	require.NoError(t, err)
	assert.Empty(t, digest)
}

func TestMisoKnowsTheDebianCloudImagesAreInQcow2(t *testing.T) {
	// arrange
	known := baseimage.Known["amd64"]

	// act
	formats := []string{known["debian:sid"].Format, known["debian:13"].Format, known["debian:trixie"].Format}

	// assert
	assert.Equal(t, []string{"qcow2", "qcow2", "qcow2"}, formats)
}

func TestMisoKnowsTheDebianArm64CloudImages(t *testing.T) {
	// arrange
	known := baseimage.Known["arm64"]

	// act
	urls := []string{known["debian:sid"].URL, known["debian:13"].URL, known["debian:trixie"].URL}

	// assert
	assert.Equal(t, []string{
		"https://cloud.debian.org/images/cloud/sid/daily/latest/debian-sid-genericcloud-arm64-daily.qcow2",
		"https://cloud.debian.org/images/cloud/trixie/latest/debian-13-genericcloud-arm64.qcow2",
		"https://cloud.debian.org/images/cloud/trixie/latest/debian-13-genericcloud-arm64.qcow2",
	}, urls)
}
