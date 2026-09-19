package download_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/download"
)

// serving is a server in memory, one file per path.
func serving(t *testing.T, files map[string]string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, found := files[r.URL.Path]
		if !found {
			http.NotFound(w, r)

			return
		}

		_, _ = w.Write([]byte(file))
	}))
	t.Cleanup(server.Close)

	return server
}

func TestADownloadThatIsRefusedIsAnErrorAndLeavesNothingBehind(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{})
	dir := t.TempDir()
	store := download.Open(dir, server.Client())

	// act
	_, err := store.Get(context.Background(), server.URL+"/gone.qcow2")

	// assert
	assert.ErrorContains(t, err, server.URL+"/gone.qcow2")
	assert.ErrorContains(t, err, "404 Not Found")
	assert.NoDirExists(t, filepath.Join(dir, "sha256"))
}

func TestADownloadThatStopsShortIsNotKept(t *testing.T) {
	// arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "1000")
		_, _ = w.Write([]byte("short"))
	}))
	t.Cleanup(server.Close)
	dir := t.TempDir()
	store := download.Open(dir, server.Client())

	// act
	_, err := store.Get(context.Background(), server.URL+"/sid.qcow2")

	// assert
	assert.Error(t, err)
	leftovers, _ := filepath.Glob(filepath.Join(dir, "sha256", "*"))
	assert.Empty(t, leftovers)
}

func TestACancelledDownloadStops(t *testing.T) {
	// arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)
	store := download.Open(t.TempDir(), server.Client())
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)

	// act
	_, err := store.Get(ctx, server.URL+"/sid.qcow2")

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}
