package agent_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/layer"
	"github.com/The127/miso/internal/protocol"
)

// finishedLongAgo finishes layers and marks them used an hour ago.
func finishedLongAgo(t *testing.T, store *layer.Store, keys ...string) time.Time {
	t.Helper()

	old := time.Now().Add(-time.Hour)
	for _, key := range keys {
		work, err := store.Begin(key)
		require.NoError(t, err)
		require.NoError(t, work.Finish())
	}

	require.NoError(t, store.Use(old, keys...))

	return old
}

func TestARunThatFindsItsLayerMarksItAndTheLayersBelowUsed(t *testing.T) {
	// arrange
	layers := t.TempDir()
	store := layer.Open(layers)
	old := finishedLongAgo(t, store, "base", "run")
	worker := agent.New(layers, t.TempDir())

	// act
	code, err := worker.Run(context.Background(), protocol.Run{Key: "run", Layers: []string{"base"}}, io.Discard)

	// assert
	require.NoError(t, err)
	require.Zero(t, code)
	entries, err := store.List()
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.True(t, entries[0].Used.After(old))
	assert.True(t, entries[1].Used.After(old))
}

func TestARunThatMustBuildMarksTheLayersBelowUsed(t *testing.T) {
	// arrange
	layers := t.TempDir()
	store := layer.Open(layers)
	old := finishedLongAgo(t, store, "base")
	worker := agent.New(layers, t.TempDir())

	// act
	_, _ = worker.Run(context.Background(), protocol.Run{Key: "run", Layers: []string{"base"}}, io.Discard)

	// assert
	entries, err := store.List()
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.True(t, entries[0].Used.After(old))
}

func TestACopyThatFindsItsLayerMarksItAndTheLayersOfBothStagesUsed(t *testing.T) {
	// arrange
	layers := t.TempDir()
	store := layer.Open(layers)
	old := finishedLongAgo(t, store, "base", "other", "copy")
	worker := agent.New(layers, t.TempDir())
	request := protocol.Copy{Key: "copy", Layers: []string{"base"}, Stage: "tools", From: []string{"other"}}

	// act
	err := worker.Copy(context.Background(), request, nil, io.Discard)

	// assert
	require.NoError(t, err)
	entries, err := store.List()
	require.NoError(t, err)
	require.Len(t, entries, 3)
	for _, entry := range entries {
		assert.True(t, entry.Used.After(old), entry.Key)
	}
}

func TestADiskThatFindsItsLayerMarksItAndTheLayersOfImageAndToolsUsed(t *testing.T) {
	// arrange
	layers := t.TempDir()
	store := layer.Open(layers)
	old := finishedLongAgo(t, store, "image", "tools", "disk")
	worker := agent.New(layers, t.TempDir())
	request := protocol.Disk{Key: "disk", Layers: []string{"image"}, Tools: []string{"tools"}}

	// act
	err := worker.Disk(context.Background(), request, io.Discard)

	// assert
	require.NoError(t, err)
	entries, err := store.List()
	require.NoError(t, err)
	require.Len(t, entries, 3)
	for _, entry := range entries {
		assert.True(t, entry.Used.After(old), entry.Key)
	}
}

func TestAnImportThatFindsItsLayerMarksItUsed(t *testing.T) {
	// arrange
	layers := t.TempDir()
	store := layer.Open(layers)
	old := finishedLongAgo(t, store, "abc")
	worker := agent.New(layers, t.TempDir())

	// act
	err := worker.Import(context.Background(), protocol.Import{Key: "abc"}, io.Discard)

	// assert
	require.NoError(t, err)
	entries, err := store.List()
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.True(t, entries[0].Used.After(old))
}
