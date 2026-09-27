package baseimage_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/baseimage"
	"github.com/The127/miso/internal/download"
)

func TestEachFetchedImageHasTheFormatItsSourceDeclares(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{"/a.img": "one image", "/b.qcow2": qcow2()})
	cache := baseimage.Open(t.TempDir(), download.Open(t.TempDir(), server.Client()), map[string]baseimage.Source{
		"a": {URL: server.URL + "/a.img", Format: "raw"},
		"b": {URL: server.URL + "/b.qcow2", Format: "qcow2"},
	})
	a, err := cache.Fetch(context.Background(), "a")
	require.NoError(t, err)
	b, err := cache.Fetch(context.Background(), "b")
	require.NoError(t, err)

	// act
	formatOfA, errOfA := cache.Format(a)
	formatOfB, errOfB := cache.Format(b)

	// assert
	require.NoError(t, errOfA)
	require.NoError(t, errOfB)
	assert.Equal(t, "raw", formatOfA)
	assert.Equal(t, "qcow2", formatOfB)
}

func TestAFetchedImagesFormatIsKnownToACacheOpenedLaterOnTheSameDirectory(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{"/sid.img": "the image"})
	sources := map[string]baseimage.Source{"debian:sid": {URL: server.URL + "/sid.img", Format: "raw"}}
	dir := t.TempDir()
	digest, err := baseimage.Open(dir, download.Open(dir, server.Client()), sources).Fetch(context.Background(), "debian:sid")
	require.NoError(t, err)

	// act
	format, err := baseimage.Open(dir, download.Open(dir, server.Client()), sources).Format(digest)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "raw", format)
}

func TestADigestNoFetchBroughtHasNoFormat(t *testing.T) {
	// arrange
	cache := baseimage.Open(t.TempDir(), download.Open(t.TempDir(), http.DefaultClient), map[string]baseimage.Source{})

	// act
	_, err := cache.Format("sha256:" + strings.Repeat("0", 64))

	// assert
	assert.ErrorIs(t, err, baseimage.ErrNotFetched)
}

// qcow2 is the start of an image in qcow2, all the header a check reads.
func qcow2() string {
	return "QFI\xfb" + strings.Repeat("\x00", 12)
}

func TestAnImageThatIsNotTheQcow2ItsSourceDeclaresIsNotFetched(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{"/sid.qcow2": "the image"})
	cache := baseimage.Open(t.TempDir(), download.Open(t.TempDir(), server.Client()), map[string]baseimage.Source{"debian:sid": {URL: server.URL + "/sid.qcow2", Format: "qcow2"}})

	// act
	_, err := cache.Fetch(context.Background(), "debian:sid")

	// assert
	assert.ErrorIs(t, err, baseimage.ErrFormat)
	digest, err := cache.Digest("debian:sid")
	require.NoError(t, err)
	assert.Empty(t, digest)
}

func TestAnImageDeclaredRawThatStartsLikeQcow2IsNotFetched(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{"/sid.img": qcow2()})
	cache := baseimage.Open(t.TempDir(), download.Open(t.TempDir(), server.Client()), map[string]baseimage.Source{"debian:sid": {URL: server.URL + "/sid.img", Format: "raw"}})

	// act
	_, err := cache.Fetch(context.Background(), "debian:sid")

	// assert
	assert.ErrorIs(t, err, baseimage.ErrFormat)
}

func TestAnImageOfAFormatMisoCannotCheckIsNotFetched(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{"/sid.vmdk": "the image"})
	cache := baseimage.Open(t.TempDir(), download.Open(t.TempDir(), server.Client()), map[string]baseimage.Source{"debian:sid": {URL: server.URL + "/sid.vmdk", Format: "vmdk"}})

	// act
	_, err := cache.Fetch(context.Background(), "debian:sid")

	// assert
	assert.ErrorIs(t, err, baseimage.ErrFormat)
}

func TestAnImageDeclaredQcow2ThatEndsBeforeItsMagicIsNotFetched(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{"/sid.qcow2": "QF"})
	cache := baseimage.Open(t.TempDir(), download.Open(t.TempDir(), server.Client()), map[string]baseimage.Source{"debian:sid": {URL: server.URL + "/sid.qcow2", Format: "qcow2"}})

	// act
	_, err := cache.Fetch(context.Background(), "debian:sid")

	// assert
	assert.ErrorIs(t, err, baseimage.ErrFormat)
}

func TestAnEmptyImageIsNotFetched(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{"/sid.img": ""})
	cache := baseimage.Open(t.TempDir(), download.Open(t.TempDir(), server.Client()), map[string]baseimage.Source{"debian:sid": {URL: server.URL + "/sid.img", Format: "raw"}})

	// act
	_, err := cache.Fetch(context.Background(), "debian:sid")

	// assert
	assert.ErrorIs(t, err, baseimage.ErrFormat)
}
