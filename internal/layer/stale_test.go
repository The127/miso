package layer_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/layer"
)

func TestALayerNotUsedForLongerThanTheLimitIsStale(t *testing.T) {
	// arrange
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	entries := []layer.Entry{
		{Key: "old", Used: now.Add(-10 * 24 * time.Hour)},
		{Key: "new", Used: now.Add(-24 * time.Hour)},
	}

	// act
	stale := layer.Stale(entries, layer.Policy{Now: now, OlderThan: 5 * 24 * time.Hour})

	// assert
	assert.Equal(t, []string{"old"}, stale)
}

func TestAPolicyWithoutAgeLimitLetsNoLayerGo(t *testing.T) {
	// arrange
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	entries := []layer.Entry{{Key: "old", Used: now.Add(-1000 * 24 * time.Hour)}}

	// act
	stale := layer.Stale(entries, layer.Policy{Now: now})

	// assert
	assert.Empty(t, stale)
}
