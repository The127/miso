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

func TestTheLeastRecentlyUsedBlobsGoUntilTheRestFitsTheStorageLimit(t *testing.T) {
	// arrange
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	blobs := []download.Blob{
		{Digest: "sha256:newest", Used: now.Add(-time.Hour), Size: 100},
		{Digest: "sha256:oldest", Used: now.Add(-3 * time.Hour), Size: 100},
		{Digest: "sha256:middle", Used: now.Add(-2 * time.Hour), Size: 100},
	}

	// act
	stale := download.Stale(blobs, download.Policy{Now: now, KeepStorage: 150})

	// assert
	assert.Equal(t, []string{"sha256:oldest", "sha256:middle"}, stale)
}
