package baseimage_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
