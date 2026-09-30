package download

import (
	"slices"
	"time"
)

// Policy says which blobs are stale. Now is the time it is judged at. A
// limit that is zero is not set, and lets nothing go. KeepStorage is the
// bytes the blobs that stay may take.
type Policy struct {
	Now         time.Time
	OlderThan   time.Duration
	KeepStorage int64
}

// Stale gives the digests of the blobs the policy lets go.
func Stale(blobs []Blob, policy Policy) []string {
	var stale []string
	var kept []Blob

	for _, blob := range blobs {
		if policy.OlderThan != 0 && policy.Now.Sub(blob.Used) > policy.OlderThan {
			stale = append(stale, blob.Digest)

			continue
		}

		kept = append(kept, blob)
	}

	if policy.KeepStorage == 0 {
		return stale
	}

	slices.SortStableFunc(kept, func(a, b Blob) int { return a.Used.Compare(b.Used) })

	var total int64
	for _, blob := range kept {
		total += blob.Size
	}

	for i := 0; i < len(kept) && total > policy.KeepStorage; i++ {
		stale = append(stale, kept[i].Digest)
		total -= kept[i].Size
	}

	return stale
}
