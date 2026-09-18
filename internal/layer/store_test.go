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

func TestOnlyAFinishedLayerIsThere(t *testing.T) {
	// arrange
	store := layer.Open(t.TempDir())
	finished, err := store.Begin("abc")
	require.NoError(t, err)
	_, err = store.Begin("def")
	require.NoError(t, err)
	require.NoError(t, finished.Finish())

	// act
	hasFinished, errFinished := store.Has("abc")
	hasBegun, errBegun := store.Has("def")

	// assert
	require.NoError(t, errFinished)
	require.NoError(t, errBegun)
	assert.True(t, hasFinished)
	assert.False(t, hasBegun)
}

func TestALayerFinishedTwiceKeepsTheFirst(t *testing.T) {
	// arrange
	dir := t.TempDir()
	store := layer.Open(dir)
	first := begun(t, store, "abc", "first")
	second := begun(t, store, "abc", "second")
	require.NoError(t, first.Finish())

	// act
	err := second.Finish()

	// assert
	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(store.Path("abc"), "hello"))
	require.NoError(t, err)
	assert.Equal(t, "first", string(got))
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

// begun is the layer of a key with a file hello that says what is given.
func begun(t *testing.T, store *layer.Store, key, says string) *layer.Work {
	t.Helper()

	work, err := store.Begin(key)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(work.Dir(), "hello"), []byte(says), 0o600))

	return work
}
