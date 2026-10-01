package lru_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/lru"
)

func TestAnItemNotUsedForLongerThanTheLimitIsStale(t *testing.T) {
	// arrange
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	items := []lru.Item{
		{ID: "old", Used: now.Add(-10 * 24 * time.Hour)},
		{ID: "new", Used: now.Add(-24 * time.Hour)},
	}

	// act
	stale := lru.Stale(items, lru.Policy{Now: now, OlderThan: 5 * 24 * time.Hour})

	// assert
	assert.Equal(t, []string{"old"}, stale)
}

func TestTheLeastRecentlyUsedItemsGoUntilTheRestFitsTheStorageLimit(t *testing.T) {
	// arrange
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	items := []lru.Item{
		{ID: "newest", Used: now.Add(-time.Hour), Size: 100},
		{ID: "oldest", Used: now.Add(-3 * time.Hour), Size: 100},
		{ID: "middle", Used: now.Add(-2 * time.Hour), Size: 100},
	}

	// act
	stale := lru.Stale(items, lru.Policy{Now: now, KeepStorage: 150})

	// assert
	assert.Equal(t, []string{"oldest", "middle"}, stale)
}

func TestItemsUsedAtTheSameTimeGoTogetherForTheStorageLimit(t *testing.T) {
	// arrange
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	tied := now.Add(-2 * time.Hour)
	items := []lru.Item{
		{ID: "parent", Used: tied, Size: 100},
		{ID: "child", Used: tied, Size: 100},
		{ID: "later", Used: now.Add(-time.Hour), Size: 100},
	}

	// act
	stale := lru.Stale(items, lru.Policy{Now: now, KeepStorage: 250})

	// assert
	assert.ElementsMatch(t, []string{"parent", "child"}, stale)
}

func TestAPolicyWithoutAgeLimitLetsNoItemGo(t *testing.T) {
	// arrange
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	items := []lru.Item{{ID: "old", Used: now.Add(-1000 * 24 * time.Hour)}}

	// act
	stale := lru.Stale(items, lru.Policy{Now: now})

	// assert
	assert.Empty(t, stale)
}
