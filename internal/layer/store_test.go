package layer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/layer"
)

func TestAFinishedLayerIsFoundUnderItsKey(t *testing.T) {
	// arrange
	store := layer.Open(t.TempDir())
	work, err := store.Begin("abc")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(work.Dir(), "hello"), []byte("hi"), 0o600))

	// act
	err = work.Finish()

	// assert
	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(store.Path("abc"), "hello"))
	require.NoError(t, err)
	assert.Equal(t, "hi", string(got))
}
