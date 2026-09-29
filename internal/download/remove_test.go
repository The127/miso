package download_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/download"
)

func TestARemovedBlobIsNotThereAnymore(t *testing.T) {
	// arrange
	server := serving(t, map[string]string{"/a": "hello"})
	store := download.Open(t.TempDir(), server.Client())
	digest, err := store.Get(context.Background(), server.URL+"/a")
	require.NoError(t, err)

	// act
	err = store.Remove(digest)

	// assert
	require.NoError(t, err)
	assert.NoFileExists(t, store.Path(digest))
}

func TestRemovingABlobThatIsNotThereIsNoError(t *testing.T) {
	// arrange
	store := download.Open(t.TempDir(), nil)

	// act
	err := store.Remove("sha256:abc")

	// assert
	assert.NoError(t, err)
}
