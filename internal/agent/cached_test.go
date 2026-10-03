package agent_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/layer"
	"github.com/The127/miso/internal/protocol"
)

func TestACachedSaysTheKeysWhoseLayersAreThereOneToALine(t *testing.T) {
	// arrange
	layers := t.TempDir()
	store := layer.Open(layers)
	finishedLongAgo(t, store, "one", "three")
	worker := agent.New(layers, t.TempDir())
	var out bytes.Buffer

	// act
	err := worker.Cached(context.Background(), protocol.Cached{Keys: []string{"one", "two", "three"}}, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "one\nthree\n", out.String())
}

func TestACachedDoesNotMarkAnyLayerAsUsed(t *testing.T) {
	// arrange
	layers := t.TempDir()
	store := layer.Open(layers)
	usedAt := finishedLongAgo(t, store, "one")
	worker := agent.New(layers, t.TempDir())

	// act
	err := worker.Cached(context.Background(), protocol.Cached{Keys: []string{"one"}}, &bytes.Buffer{})

	// assert
	require.NoError(t, err)
	entries, err := store.List()
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.WithinDuration(t, usedAt, entries[0].Used, time.Second)
}
