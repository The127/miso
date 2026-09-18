package baseimage_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/baseimage"
)

// serving is an image server in memory, one image per path.
func serving(t *testing.T, images map[string]string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		image, found := images[r.URL.Path]
		if !found {
			http.NotFound(w, r)
			return
		}

		_, _ = w.Write([]byte(image))
	}))
	t.Cleanup(server.Close)

	return server
}

func TestAFetchDownloadsAnImageAndAnswersTheDigestOfItsBytes(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{"/sid.qcow2": "the image"})
	cache := baseimage.Open(t.TempDir(), server.Client(), map[string]string{"debian:sid": server.URL + "/sid.qcow2"})
	sum := sha256.Sum256([]byte("the image"))

	// act
	digest, err := cache.Fetch(context.Background(), "debian:sid")

	// assert
	require.NoError(t, err)
	assert.Equal(t, "sha256:"+hex.EncodeToString(sum[:]), digest)
}
