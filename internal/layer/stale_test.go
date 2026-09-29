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

func TestTheLeastRecentlyUsedLayersGoUntilTheRestFitsTheStorageLimit(t *testing.T) {
	// arrange
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	entries := []layer.Entry{
		{Key: "newest", Used: now.Add(-time.Hour), Size: 100},
		{Key: "oldest", Used: now.Add(-3 * time.Hour), Size: 100},
		{Key: "middle", Used: now.Add(-2 * time.Hour), Size: 100},
	}

	// act
	stale := layer.Stale(entries, layer.Policy{Now: now, KeepStorage: 150})

	// assert
	assert.Equal(t, []string{"oldest", "middle"}, stale)
}

func TestLayersUsedAtTheSameTimeGoTogetherForTheStorageLimit(t *testing.T) {
	// arrange
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	tied := now.Add(-2 * time.Hour)
	entries := []layer.Entry{
		{Key: "parent", Used: tied, Size: 100},
		{Key: "child", Used: tied, Size: 100},
		{Key: "later", Used: now.Add(-time.Hour), Size: 100},
	}

	// act
	stale := layer.Stale(entries, layer.Policy{Now: now, KeepStorage: 250})

	// assert
	assert.ElementsMatch(t, []string{"parent", "child"}, stale)
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
