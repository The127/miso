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

func TestAnImportThatFindsItsLayerMarksItUsed(t *testing.T) {
	// arrange
	layers := t.TempDir()
	store := layer.Open(layers)
	work, err := store.Begin("abc")
	require.NoError(t, err)
	require.NoError(t, work.Finish())
	old := time.Now().Add(-time.Hour)
	require.NoError(t, store.Use(old, "abc"))
	worker := agent.New(layers, t.TempDir())

	// act
	err = worker.Import(context.Background(), protocol.Import{Key: "abc"}, io.Discard)

	// assert
	require.NoError(t, err)
	entries, err := store.List()
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.True(t, entries[0].Used.After(old))
}
