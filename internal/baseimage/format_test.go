package baseimage_test

import (
	"context"
	"encoding/binary"
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

// declaring is a cache that knows debian:sid as an image of some bytes, whose
// source declares a format.
func declaring(t *testing.T, image, format string) *baseimage.Cache {
	t.Helper()

	server := serving(t, map[string]string{"/image": image})

	return baseimage.Open(t.TempDir(), download.Open(t.TempDir(), server.Client()), map[string]baseimage.Source{"debian:sid": {URL: server.URL + "/image", Format: format}})
}

// qcow2 is the header of an image in qcow2 version 3, naming nothing
// outside itself.
func qcow2() string {
	return "QFI\xfb" + "\x00\x00\x00\x03" + strings.Repeat("\x00", 96)
}

func TestAnImageThatIsNotTheQcow2ItsSourceDeclaresIsNotFetched(t *testing.T) {
	// arrange
	cache := declaring(t, "the image", "qcow2")

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
	cache := declaring(t, qcow2(), "raw")

	// act
	_, err := cache.Fetch(context.Background(), "debian:sid")

	// assert
	assert.ErrorIs(t, err, baseimage.ErrFormat)
}

func TestAnImageOfAFormatMisoCannotCheckIsNotFetched(t *testing.T) {
	// arrange
	cache := declaring(t, "the image", "vmdk")

	// act
	_, err := cache.Fetch(context.Background(), "debian:sid")

	// assert
	assert.ErrorIs(t, err, baseimage.ErrFormat)
}

func TestAnImageDeclaredQcow2ThatEndsBeforeItsMagicIsNotFetched(t *testing.T) {
	// arrange
	cache := declaring(t, "QF", "qcow2")

	// act
	_, err := cache.Fetch(context.Background(), "debian:sid")

	// assert
	assert.ErrorIs(t, err, baseimage.ErrFormat)
}

func TestAnEmptyImageIsNotFetched(t *testing.T) {
	// arrange
	cache := declaring(t, "", "raw")

	// act
	_, err := cache.Fetch(context.Background(), "debian:sid")

	// assert
	assert.ErrorIs(t, err, baseimage.ErrFormat)
}

// with is an image with a big-endian number written at an offset.
func with(image string, offset int, number uint64) string {
	bytes := []byte(image)
	binary.BigEndian.PutUint64(bytes[offset:], number)

	return string(bytes)
}

func TestAQcow2ImageThatNamesABackingFileIsNotFetched(t *testing.T) {
	// arrange
	cache := declaring(t, with(qcow2(), 8, 512), "qcow2")

	// act
	_, err := cache.Fetch(context.Background(), "debian:sid")

	// assert
	assert.ErrorIs(t, err, baseimage.ErrExternalFile)
}

func TestAQcow2ImageWhoseHeaderEndsEarlyIsNotFetched(t *testing.T) {
	// arrange
	cache := declaring(t, qcow2()[:10], "qcow2")

	// act
	_, err := cache.Fetch(context.Background(), "debian:sid")

	// assert
	assert.ErrorIs(t, err, baseimage.ErrFormat)
}

func TestAQcow2ImageWithAnExternalDataFileIsNotFetched(t *testing.T) {
	// arrange
	cache := declaring(t, with(qcow2(), 72, 1<<2), "qcow2")

	// act
	_, err := cache.Fetch(context.Background(), "debian:sid")

	// assert
	assert.ErrorIs(t, err, baseimage.ErrExternalFile)
}

// version2 is a qcow2 image made version 2, whose header ends before any
// features.
func version2(image string) string {
	return image[:4] + "\x00\x00\x00\x02" + image[8:]
}

func TestAVersion2Qcow2ImageIsFetchedWhateverFollowsItsHeader(t *testing.T) {
	// arrange
	cache := declaring(t, with(version2(qcow2()), 72, 1<<2), "qcow2")

	// act
	_, err := cache.Fetch(context.Background(), "debian:sid")

	// assert
	assert.NoError(t, err)
}
