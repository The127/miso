package lru_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/lru"
)

func TestAPruneRemovesTheStaleItemsAndSaysHowMuchWent(t *testing.T) {
	// arrange
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	items := []lru.Item{
		{ID: "old", Used: now.Add(-10 * 24 * time.Hour), Size: 7},
		{ID: "new", Used: now.Add(-time.Hour), Size: 5},
	}

	var removed []string

	// act
	swept, err := lru.Prune(items, lru.Policy{Now: now, OlderThan: 5 * 24 * time.Hour}, func(id string) error {
		removed = append(removed, id)

		return nil
	})

	// assert
	require.NoError(t, err)
	assert.Equal(t, []string{"old"}, removed)
	assert.Equal(t, lru.Swept{Count: 1, Bytes: 7}, swept)
}

func TestAPruneStopsAtTheItemItCannotRemoveAndCountsWhatWent(t *testing.T) {
	// arrange
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	items := []lru.Item{
		{ID: "a", Used: now.Add(-10 * 24 * time.Hour), Size: 1},
		{ID: "b", Used: now.Add(-10 * 24 * time.Hour), Size: 1},
	}
	stuck := errors.New("stuck")

	// act
	swept, err := lru.Prune(items, lru.Policy{Now: now, OlderThan: time.Hour}, func(id string) error {
		if id == "b" {
			return stuck
		}

		return nil
	})

	// assert
	require.ErrorIs(t, err, stuck)
	assert.Equal(t, lru.Swept{Count: 1, Bytes: 1}, swept)
}
