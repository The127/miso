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

func TestAPruneRemovesTheLayersNotUsedForLongerThanTheLimitAndSaysSo(t *testing.T) {
	// arrange
	layers := t.TempDir()
	store := layer.Open(layers)
	finishedLongAgo(t, store, "old", "new")
	require.NoError(t, store.Use(time.Now().Add(-10*24*time.Hour), "old"))
	worker := agent.New(layers, t.TempDir())
	var out bytes.Buffer

	// act
	err := worker.Prune(context.Background(), protocol.Prune{OlderThan: 5 * 24 * time.Hour}, &out)

	// assert
	require.NoError(t, err)
	entries, err := store.List()
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "new", entries[0].Key)
	assert.Contains(t, out.String(), "removed 1 layers")
}
