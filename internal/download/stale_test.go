package download_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/download"
)

func TestABlobNotUsedForLongerThanTheLimitIsStale(t *testing.T) {
	// arrange
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	blobs := []download.Blob{
		{Digest: "sha256:old", Used: now.Add(-10 * 24 * time.Hour)},
		{Digest: "sha256:new", Used: now.Add(-24 * time.Hour)},
	}

	// act
	stale := download.Stale(blobs, download.Policy{Now: now, OlderThan: 5 * 24 * time.Hour})

	// assert
	assert.Equal(t, []string{"sha256:old"}, stale)
}
